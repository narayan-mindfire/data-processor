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
	log  *slog.Logger

	// Core Pipeline Channels
	recordsCh     chan *PipelineRecord
	validatedCh   chan *PipelineRecord
	transformedCh chan *PipelineRecord
	exportCh      chan interface{} // Can hold *PipelineRecord or *models.JobResult

	// Side Channels for Observability
	errorCh    chan *models.JobError
	progressCh chan int
}

func NewPipelineEngine(job *models.Job, repo JobRepository, log *slog.Logger) *PipelineEngine {
	return &PipelineEngine{
		job:           job,
		repo:          repo,
		log:           log,
		recordsCh:     make(chan *PipelineRecord, 100),
		validatedCh:   make(chan *PipelineRecord, 100),
		transformedCh: make(chan *PipelineRecord, 100),
		exportCh:      make(chan interface{}, 100),
		errorCh:       make(chan *models.JobError, 100),
		progressCh:    make(chan int, 100),
	}
}

func (e *PipelineEngine) Run(ctx context.Context) {
	e.log.Info("Starting pipeline engine", "job_id", e.job.ID)

	_ = e.updateJobStatus(ctx, models.StatusRunning)

	var obsWg sync.WaitGroup
	obsWg.Add(1)
	go e.observabilityTracker(ctx, &obsWg)

	var ingestWg sync.WaitGroup
	for _, source := range e.job.Config.Sources {
		ingestWg.Add(1)
		if source.Type == "csv" {
			go e.ingestCSV(ctx, source, &ingestWg)
		} else if source.Type == "json" {
			go e.ingestJSON(ctx, source, &ingestWg)
		} else {
			ingestWg.Done()
		}
	}

	go func() {
		ingestWg.Wait()
		close(e.recordsCh)
	}()

	var valWg sync.WaitGroup
	numVal := e.job.Config.Concurrency.ValidationWorkers
	if numVal <= 0 {
		numVal = 1
	}
	for i := 0; i < numVal; i++ {
		valWg.Add(1)
		go e.validationWorker(ctx, &valWg)
	}

	go func() {
		valWg.Wait()
		close(e.validatedCh)
	}()

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

	var exportWg sync.WaitGroup
	exportWg.Add(1)
	go e.exportWorker(ctx, &exportWg)

	e.aggregate(ctx)

	exportWg.Wait()

	// Wait for all worker WaitGroups to finish their shutdown sequences
	ingestWg.Wait()
	valWg.Wait()
	transWg.Wait()

	close(e.progressCh)
	close(e.errorCh)
	obsWg.Wait()

	_ = e.updateJobStatus(ctx, models.StatusCompleted)
	e.log.Info("Pipeline engine finished successfully", "job_id", e.job.ID)
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
	e.log.Info("Starting Aggregation Fan-In stage", "job_id", e.job.ID)

	results := make(map[string]float64)
	counts := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			e.log.Info("Aggregation cancelled by user", "job_id", e.job.ID)
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

				e.exportCh <- jobResult
				close(e.exportCh)

				e.log.Info("Aggregation complete! Final Results passed to Export Stage", "job_id", e.job.ID, "results", string(summaryBytes))
				return
			}

			e.progressCh <- 1
			e.exportCh <- record

			for _, agg := range e.job.Config.Aggregations {
				val, exists := record.Data[agg.Field]
				if !exists || val == nil {
					continue
				}

				if agg.Type == "count" {
					results[agg.OutputName]++
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
				}
			}
		}
	}
}

func (e *PipelineEngine) observabilityTracker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	processed := 0
	errors := 0

	for {
		select {
		case _, ok := <-e.progressCh:
			if !ok {
				e.progressCh = nil
			} else {
				processed++
			}
		case jobErr, ok := <-e.errorCh:
			if !ok {
				e.errorCh = nil
			} else {
				errors++
				e.log.Error("Pipeline Error Occurred", "stage", jobErr.Stage, "msg", jobErr.ErrorMessage)
				if err := e.repo.InsertJobError(context.Background(), jobErr); err != nil {
					e.log.Error("Failed to persist job error", "error", err)
				}
			}
		case <-ticker.C:
			_ = e.repo.UpdateJobProgress(context.Background(), e.job.ID, processed, errors)
		}

		if e.progressCh == nil && e.errorCh == nil {
			_ = e.repo.UpdateJobProgress(context.Background(), e.job.ID, processed, errors)
			return
		}
	}
}

func (e *PipelineEngine) ingestCSV(ctx context.Context, source models.SourceDef, wg *sync.WaitGroup) {
	defer wg.Done()
	e.log.Info("Starting CSV ingestion", "url", source.URL)

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
	reader.LazyQuotes = true       // Ignore missing/malformed quotes
	reader.TrimLeadingSpace = true // Ignore spaces after commas
	headers, err := reader.Read()
	if err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-CSV", ErrorMessage: "Failed to read CSV headers: " + err.Error()}
		return
	}

	recordIndex := 0
	for {
		select {
		case <-ctx.Done():
			e.log.Info("CSV Ingestion cancelled", "url", source.URL)
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
	e.log.Info("Finished CSV ingestion", "url", source.URL, "records_read", recordIndex)
}

func (e *PipelineEngine) ingestJSON(ctx context.Context, source models.SourceDef, wg *sync.WaitGroup) {
	defer wg.Done()
	e.log.Info("Starting JSON ingestion", "url", source.URL)

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

	var rawData interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: "Failed to decode JSON: " + err.Error()}
		return
	}

	var records []map[string]any

	switch v := rawData.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				records = append(records, m)
			}
		}
	case map[string]interface{}:
		if source.JSONArrayPath != "" && source.JSONArrayPath != "$" {
			if nested, ok := v[source.JSONArrayPath]; ok {
				if nestedArr, ok := nested.([]interface{}); ok {
					for _, item := range nestedArr {
						if m, ok := item.(map[string]interface{}); ok {
							records = append(records, m)
						}
					}
				} else if m, ok := nested.(map[string]interface{}); ok {
					records = append(records, m)
				}
			}
		} else {
			records = append(records, v)
		}
	default:
		e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Ingest-JSON", ErrorMessage: "Unsupported JSON root type"}
		return
	}

	for i, data := range records {
		select {
		case <-ctx.Done():
			e.log.Info("JSON Ingestion cancelled", "url", source.URL)
			return
		default:
		}

		e.recordsCh <- &PipelineRecord{
			Index:     i,
			SourceURL: source.URL,
			Data:      data,
		}
	}
	e.log.Info("Finished JSON ingestion", "url", source.URL, "records_read", len(records))
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

func (e *PipelineEngine) exportWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	e.log.Info("Starting Export Worker", "job_id", e.job.ID)

	for {
		select {
		case <-ctx.Done():
			e.log.Info("Export Worker cancelled", "job_id", e.job.ID)
			return
		case data, ok := <-e.exportCh:
			if !ok {
				e.log.Info("Export Worker finished", "job_id", e.job.ID)
				return
			}

			switch v := data.(type) {
			case *PipelineRecord:
				if err := e.repo.InsertExportedRecord(context.Background(), e.job.ID, v.Data); err != nil {
					e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Export", RecordIndex: v.Index, ErrorMessage: "Failed to persist record: " + err.Error()}
				}
			case *models.JobResult:
				if err := e.repo.InsertJobResult(context.Background(), v); err != nil {
					e.errorCh <- &models.JobError{JobID: e.job.ID, Stage: "Export", ErrorMessage: "Failed to save final results: " + err.Error()}
				}
			}
		}
	}
}
