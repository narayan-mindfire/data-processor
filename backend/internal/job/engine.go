package job

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

type PipelineRecord struct {
	Index     int
	SourceURL string
	Data      map[string]any
}

type PipelineEngine struct {
	job  *models.Job
	repo JobRepository

	// Core Pipeline Channels
	recordsCh     chan *PipelineRecord
	validatedCh   chan *PipelineRecord
	transformedCh chan *PipelineRecord

	// Side Channels for Observability
	errorCh    chan *models.JobError
	progressCh chan int
}

func NewPipelineEngine(job *models.Job, repo JobRepository) *PipelineEngine {
	return &PipelineEngine{
		job:           job,
		repo:          repo,
		recordsCh:     make(chan *PipelineRecord, 100),
		validatedCh:   make(chan *PipelineRecord, 100),
		transformedCh: make(chan *PipelineRecord, 100),
		errorCh:       make(chan *models.JobError, 100),
		progressCh:    make(chan int, 100),
	}
}

// Run executes the entire pipeline lifecycle
func (e *PipelineEngine) Run(ctx context.Context) {
	slog.Info("Starting pipeline engine", "job_id", e.job.ID)

	// Update DB to RUNNING
	_ = e.updateJobStatus(ctx, models.StatusRunning)

	// 1. Start Background Side-Channel Listeners (Observability)
	var obsWg sync.WaitGroup
	obsWg.Add(2)
	go e.progressTracker(ctx, &obsWg)
	go e.errorCollector(ctx, &obsWg)

	// 2. Stage 1: Ingestion (Multi-source Fan-out)
	var ingestWg sync.WaitGroup
	for _, source := range e.job.Config.Sources {
		ingestWg.Add(1)
		if source.Type == "csv" {
			go e.ingestCSV(ctx, source, &ingestWg)
		} else if source.Type == "json" {
			go e.ingestJSON(ctx, source, &ingestWg)
		} else {
			ingestWg.Done() // Skip unknown sources
		}
	}

	// Close recordsCh when all sources finish downloading/parsing
	go func() {
		ingestWg.Wait()
		close(e.recordsCh)
	}()

	// 3. Stage 2: Validation (Worker Pool Fan-out)
	var valWg sync.WaitGroup
	numVal := e.job.Config.Concurrency.ValidationWorkers
	if numVal <= 0 {
		numVal = 1
	}
	for i := 0; i < numVal; i++ {
		valWg.Add(1)
		go e.validationWorker(ctx, &valWg)
	}

	// Close validatedCh when all validators finish
	go func() {
		valWg.Wait()
		close(e.validatedCh)
	}()

	// 4. Stage 3: Transformation (Worker Pool Fan-out)
	var transWg sync.WaitGroup
	numTrans := e.job.Config.Concurrency.TransformWorkers
	if numTrans <= 0 {
		numTrans = 1
	}
	for i := 0; i < numTrans; i++ {
		transWg.Add(1)
		go e.transformationWorker(ctx, &transWg)
	}

	go func() {
		transWg.Wait()
		close(e.transformedCh)
	}()

	// 5. Stage 4: Aggregation (Fan-in)
	e.aggregate(ctx)

	// 6. Cleanup
	close(e.progressCh)
	close(e.errorCh)
	obsWg.Wait()

	_ = e.updateJobStatus(ctx, models.StatusCompleted)
	slog.Info("Pipeline engine finished successfully", "job_id", e.job.ID)
}

// Helper to update Job status cleanly via the Repository
func (e *PipelineEngine) updateJobStatus(ctx context.Context, status string) error {
	var finishedAt *time.Time
	if status == models.StatusCompleted || status == models.StatusFailed || status == models.StatusCancelled {
		now := time.Now()
		finishedAt = &now
	}
	return e.repo.UpdateJobStatus(ctx, e.job.ID, status, finishedAt)
}

// --- STUBS FOR COMPILATION (We will build these next!) ---

func (e *PipelineEngine) progressTracker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) errorCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) ingestCSV(ctx context.Context, source models.SourceDef, wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("Starting CSV ingestion", "url", source.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", ErrorMessage: "Failed to create request: " + err.Error()}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", ErrorMessage: "HTTP request failed: " + err.Error()}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", ErrorMessage: fmt.Sprintf("Bad HTTP status: %d", resp.StatusCode)}
		return
	}

	reader := csv.NewReader(resp.Body)
	headers, err := reader.Read()
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", ErrorMessage: "Failed to read CSV headers: " + err.Error()}
		return
	}

	recordIndex := 0
	for {
		select {
		case <-ctx.Done():
			slog.Info("CSV Ingestion cancelled", "url", source.URL)
			return
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", RecordIndex: recordIndex, ErrorMessage: "Failed to read CSV row: " + err.Error()}
			continue
		}

		// Convert CSV string row into a generic map based on headers!
		data := make(map[string]any)
		for i, value := range row {
			if i < len(headers) {
				data[headers[i]] = value
			}
		}

		// Send it down the pipe!
		e.recordsCh <- &PipelineRecord{
			Index:     recordIndex,
			SourceURL: source.URL,
			Data:      data,
		}
		recordIndex++
	}
	slog.Info("Finished CSV ingestion", "url", source.URL, "records_read", recordIndex)
}

func (e *PipelineEngine) ingestJSON(ctx context.Context, source models.SourceDef, wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("Starting JSON ingestion", "url", source.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: "Failed to create request: " + err.Error()}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: "HTTP request failed: " + err.Error()}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: fmt.Sprintf("Bad HTTP status: %d", resp.StatusCode)}
		return
	}

	var rawData []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: "Failed to decode JSON array: " + err.Error()}
		return
	}

	for i, data := range rawData {
		select {
		case <-ctx.Done():
			slog.Info("JSON Ingestion cancelled", "url", source.URL)
			return
		default:
		}

		e.recordsCh <- &PipelineRecord{
			Index:     i,
			SourceURL: source.URL,
			Data:      data,
		}
	}
	slog.Info("Finished JSON ingestion", "url", source.URL, "records_read", len(rawData))
}

func (e *PipelineEngine) validationWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) transformationWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) aggregate(ctx context.Context) {
}
