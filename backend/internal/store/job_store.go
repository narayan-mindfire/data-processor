package store

import (
	"context"
	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

// JobStore defines the contract for any database interacting with Jobs.
type JobStore interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
}
