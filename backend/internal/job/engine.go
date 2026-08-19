package job

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

func (e *PipelineEngine) aggregate(ctx context.Context) {
	slog.Info("Starting Aggregation Fan-In stage", "job_id", e.job.ID)

	results := make(map[string]float64)
	counts := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Aggregation cancelled by user", "job_id", e.job.ID)
			return
		case record, ok := <-e.transformedCh:
			if !ok {

				for _, agg := range e.job.Config.Aggregations {
					if agg.Type == "average" {
						if counts[agg.OutputName] > 0 {
							results[agg.OutputName] = results[agg.OutputName] / float64(counts[agg.OutputName])
						}
					}
				}

				summaryBytes, err := json.Marshal(results)
				if err != nil {
					e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Aggregation", ErrorMessage: "Failed to marshal results: " + err.Error()}
					return
				}

				jobResult := &models.JobResult{
					JobID:       e.job.ID,
					SummaryJSON: string(summaryBytes),
				}

				if err := e.repo.InsertJobResult(context.Background(), jobResult); err != nil {
					slog.Error("Failed to save final results to DB", "error", err)
				}

				slog.Info("Aggregation complete! Final Results Saved", "job_id", e.job.ID, "results", string(summaryBytes))
				return
			}

			e.progressCh <- 1

			for _, agg := range e.job.Config.Aggregations {
				val, exists := record.Data[agg.Field]
				if !exists || val == nil {
					continue
				}

				var num float64
				switch v := val.(type) {
				case int:
					num = float64(v)
				case float64:
					num = v
				default:
					continue
				}

				if agg.Type == "sum" || agg.Type == "average" {
					results[agg.OutputName] += num
					counts[agg.OutputName]++
				} else if agg.Type == "count" {
					results[agg.OutputName]++
				}
			}
		}
	}
}

func (e *PipelineEngine) progressTracker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	processed := 0
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-e.progressCh:
			if !ok {
				// Channel closed by orchestrator. Final flush before shutting down.
				_ = e.repo.UpdateJobProgress(context.Background(), e.job.ID, processed, 0)
				return
			}
			processed++
		case <-ticker.C:
			// Flush to database periodically
			_ = e.repo.UpdateJobProgress(context.Background(), e.job.ID, processed, 0)
		}
	}
}

func (e *PipelineEngine) errorCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for jobErr := range e.errorCh {
		slog.Error("Pipeline Error Occurred", "stage", jobErr.Stage, "msg", jobErr.ErrorMessage)

		if err := e.repo.InsertJobError(context.Background(), jobErr); err != nil {
			slog.Error("Failed to persist job error", "error", err)
		}
	}
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

	for {
		select {
		case <-ctx.Done():
			return // Job cancelled
		case record, ok := <-e.recordsCh:
			if !ok {
				return // No more records to validate
			}

			isValid := true
			for _, rule := range e.job.Config.Validations {
				val, exists := record.Data[rule.Field]

				if !exists || val == nil {
					if rule.Rule == "not_empty" {
						e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Validation", RecordIndex: record.Index, ErrorMessage: fmt.Sprintf("Missing required field: %s", rule.Field)}
						isValid = false
						break
					}
					continue
				}

				strVal := fmt.Sprintf("%v", val)

				if rule.Rule == "not_empty" && strings.TrimSpace(strVal) == "" {
					e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Validation", RecordIndex: record.Index, ErrorMessage: fmt.Sprintf("Field %s cannot be empty", rule.Field)}
					isValid = false
					break
				}

				if rule.Rule == "is_numeric" {
					if _, err := strconv.ParseFloat(strings.TrimSpace(strVal), 64); err != nil {
						e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Validation", RecordIndex: record.Index, ErrorMessage: fmt.Sprintf("Field %s must be numeric, got: '%s'", rule.Field, strVal)}
						isValid = false
						break
					}
				}
			}

			if isValid {
				e.validatedCh <- record
			}
		}
	}
}

func (e *PipelineEngine) transformationWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case record, ok := <-e.validatedCh:
			if !ok {
				return
			}

			for _, rule := range e.job.Config.Transformations {
				val, exists := record.Data[rule.Field]

				strVal := ""
				if exists && val != nil {
					strVal = strings.TrimSpace(fmt.Sprintf("%v", val))
				}

				if rule.Action == "fill_empty" {
					if strVal == "" {
						record.Data[rule.Field] = rule.DefaultValue
						strVal = rule.DefaultValue
					}
				}

				if rule.Action == "convert_to_int" {
					if strVal != "" {
						if intVal, err := strconv.Atoi(strVal); err == nil {
							record.Data[rule.Field] = intVal
						} else {
							e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Transformation", RecordIndex: record.Index, ErrorMessage: fmt.Sprintf("Failed to convert %s to int", rule.Field)}
						}
					}
				}

				if rule.Action == "convert_to_float" {
					if strVal != "" {
						if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
							record.Data[rule.Field] = floatVal
						} else {
							e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Transformation", RecordIndex: record.Index, ErrorMessage: fmt.Sprintf("Failed to convert %s to float", rule.Field)}
						}
					}
				}
			}

			e.transformedCh <- record
		}
	}
}
