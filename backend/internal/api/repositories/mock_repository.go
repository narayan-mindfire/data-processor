package repositories

import (
	"context"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/pkg/apperrors"
)

type MockRepository struct {
	Jobs map[string]*models.Job
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Jobs: make(map[string]*models.Job),
	}
}

func (m *MockRepository) CreateJob(ctx context.Context, job *models.Job) error {
	m.Jobs[job.ID] = job
	return nil
}

func (m *MockRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	job, exists := m.Jobs[id]
	if !exists {
		return nil, errors.ErrNotFound
	}
	return job, nil
}
