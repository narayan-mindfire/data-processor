package repositories

import (
	"context"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
}

type PostgresJobRepository struct {
	DB *store.DB
}

func NewPostgresJobRepository(db *store.DB) *PostgresJobRepository {
	return &PostgresJobRepository{DB: db}
}

func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *models.Job) error {
	return nil
}

func (r *PostgresJobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	return nil, nil
}