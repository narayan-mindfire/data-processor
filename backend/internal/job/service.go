package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
	UpdateJobProgress(ctx context.Context, id string, processedRecords, errorCount int) error
	UpdateJobStatus(ctx context.Context, id string, status string, finishedAt *time.Time) error
	InsertJobError(ctx context.Context, jobError *models.JobError) error
	InsertJobResult(ctx context.Context, result *models.JobResult) error
	GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error)
	GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context) ([]models.Job, error)
}

type PipelineService struct {
	repo       JobRepository
	activeJobs map[string]context.CancelFunc
	mu         sync.Mutex
}

func NewPipelineService(repo JobRepository) *PipelineService {
	return &PipelineService{
		repo:       repo,
		activeJobs: make(map[string]context.CancelFunc),
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

	engine := NewPipelineEngine(job, s.repo)

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.activeJobs, job.ID)
			s.mu.Unlock()
		}()
		engine.Run(pipelineCtx)
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
