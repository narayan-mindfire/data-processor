package job

import (
	"context"
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
}

type PipelineService struct {
	repo JobRepository
}

func NewPipelineService(repo JobRepository) *PipelineService {
	return &PipelineService{repo: repo}
}

func (s *PipelineService) StartPipeline(ctx context.Context, job *models.Job) error {
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return err
	}

	// 2. trigger the massive concurrent pipeline here!
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
