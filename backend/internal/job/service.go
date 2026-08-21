package job

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
	UpdateJobProgress(ctx context.Context, id string, processedRecords, errorCount int) error
	UpdateJobStatus(ctx context.Context, id string, status string, finishedAt *time.Time) error
	UpdateJobMetrics(ctx context.Context, id string, metrics map[string]interface{}) error
	InsertJobError(ctx context.Context, jobError *models.JobError) error
	InsertJobResult(ctx context.Context, result *models.JobResult) error
	GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error)
	GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error)
	InsertExportedRecord(ctx context.Context, jobID string, sourceURL string, data map[string]any) error
	GetExportedRecordsBySource(ctx context.Context, jobID string, sourceURL string) (*sql.Rows, error)
	GetDistinctSources(ctx context.Context, jobID string) ([]string, error)
	DeleteExportedRecords(ctx context.Context, jobID string) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context) ([]models.Job, error)
}

type PipelineService struct {
	repo       JobRepository
	s3Client   *store.S3Client
	activeJobs map[string]context.CancelFunc
	mu         sync.Mutex
	log        *slog.Logger
}

func NewPipelineService(repo JobRepository, s3Client *store.S3Client, log *slog.Logger) *PipelineService {
	return &PipelineService{
		repo:       repo,
		s3Client:   s3Client,
		activeJobs: make(map[string]context.CancelFunc),
		log:        log,
	}
}

func (s *PipelineService) StartPipeline(ctx context.Context, job *models.Job) error {
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return err
	}

	pipelineCtx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.activeJobs[job.ID] = cancel
	s.mu.Unlock()

	engine := NewPipelineEngine(job, s.repo, s.log)

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.activeJobs, job.ID)
			s.mu.Unlock()
		}()
		engine.Run(pipelineCtx)

		status := models.StatusCompleted
		if pipelineCtx.Err() == context.Canceled {
			status = models.StatusCancelled
		}

		now := time.Now()
		_ = s.repo.UpdateJobStatus(context.Background(), job.ID, status, &now)
		_ = s.repo.UpdateJobMetrics(context.Background(), job.ID, engine.GetMetrics())

		if status == models.StatusCompleted {
			// Trigger ephemeral buffer cleanup and S3 sync in the background
			go s.SyncToS3(job.ID)
		}
	}()

	return nil
}

func (s *PipelineService) CancelJob(ctx context.Context, id string) error {
	s.mu.Lock()
	cancel, exists := s.activeJobs[id]
	s.mu.Unlock()

	if !exists {
		job, err := s.repo.GetJobByID(ctx, id)
		if err != nil {
			return err
		}
		if job.Status != models.StatusRunning && job.Status != models.StatusPending {
			return fmt.Errorf("job is already %s", job.Status)
		}

		now := time.Now()
		return s.repo.UpdateJobStatus(ctx, id, models.StatusCancelled, &now)
	}

	cancel()
	return nil
}

func (s *PipelineService) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	return s.repo.GetJobByID(ctx, id)
}

func (s *PipelineService) GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error) {
	return s.repo.GetJobErrors(ctx, jobID)
}

func (s *PipelineService) GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error) {
	return s.repo.GetJobResults(ctx, jobID)
}

func (s *PipelineService) DeleteJob(ctx context.Context, id string) error {
	return s.repo.DeleteJob(ctx, id)
}

func (s *PipelineService) ListJobs(ctx context.Context) ([]models.Job, error) {
	return s.repo.ListJobs(ctx)
}
