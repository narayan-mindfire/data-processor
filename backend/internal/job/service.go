package job

import (
	"context"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
	UpdateJobProgress(ctx context.Context, id string, processedRecords, errorCount int) error
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
