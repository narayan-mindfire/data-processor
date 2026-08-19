package job

import (
	"context"
	"log/slog"
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
}

func (e *PipelineEngine) ingestJSON(ctx context.Context, source models.SourceDef, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) validationWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) transformationWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (e *PipelineEngine) aggregate(ctx context.Context) {
}
