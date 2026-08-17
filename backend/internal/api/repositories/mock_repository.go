package repositories

import (
	"context"
	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/pkg/errors"
)

// MockStore is an in-memory implementation of the JobStore interface for testing.
type MockStore struct {
	Jobs map[string]*models.Job
}

// NewMockStore initializes an empty fake database
func NewMockStore() *MockStore {
	return &MockStore{
		Jobs: make(map[string]*models.Job),
	}
}

// CreateJob saves a job to our in-memory map
func (m *MockStore) CreateJob(ctx context.Context, job *models.Job) error {
	m.Jobs[job.ID] = job
	return nil
}

// GetJobByID retrieves a job or returns our custom ErrNotFound
func (m *MockStore) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	job, exists := m.Jobs[id]
	if !exists {
		return nil, errors.ErrNotFound
	}
	return job, nil
}
