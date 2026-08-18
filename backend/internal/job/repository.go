package job

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
	"github.com/narayan-mindfire/data-processor/backend/pkg/apperrors"
)

type PostgresJobRepository struct {
	DB *store.DB
}

func NewPostgresJobRepository(db *store.DB) *PostgresJobRepository {
	return &PostgresJobRepository{DB: db}
}

func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (id, source_type, status, created_at) 
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.ExecContext(ctx, query, job.ID, job.SourceType, job.Status, job.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT id, source_type, status, total_records, processed_records, error_count, created_at, finished_at 
		FROM jobs 
		WHERE id = $1
	`

	var job models.Job
	var finishedAt sql.NullTime

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.SourceType,
		&job.Status,
		&job.TotalRecords,
		&job.ProcessedRecords,
		&job.ErrorCount,
		&job.CreatedAt,
		&finishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	if finishedAt.Valid {
		job.FinishedAt = &finishedAt.Time
	}

	return &job, nil
}

func (r *PostgresJobRepository) UpdateJobProgress(ctx context.Context, id string, processedRecords, errorCount int) error {
	query := `
		UPDATE jobs 
		SET processed_records = $2, error_count = $3 
		WHERE id = $1
	`
	_, err := r.DB.ExecContext(ctx, query, id, processedRecords, errorCount)
	if err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}
	return nil
}
