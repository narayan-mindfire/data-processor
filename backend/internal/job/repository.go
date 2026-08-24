package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
		INSERT INTO jobs (id, config, status, created_at) 
		VALUES ($1, $2, $3, $4)
	`
	configBytes, err := json.Marshal(job.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal job config: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, query, job.ID, configBytes, job.Status, job.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT id, config, status, total_records, processed_records, error_count, created_at, finished_at, metrics 
		FROM jobs 
		WHERE id = $1
	`
	var job models.Job
	var configBytes []byte
	var metricsBytes []byte
	var finishedAt sql.NullTime

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&configBytes,
		&job.Status,
		&job.TotalRecords,
		&job.ProcessedRecords,
		&job.ErrorCount,
		&job.CreatedAt,
		&finishedAt,
		&metricsBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	if err := json.Unmarshal(configBytes, &job.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job config: %w", err)
	}

	if len(metricsBytes) > 0 {
		_ = json.Unmarshal(metricsBytes, &job.Metrics)
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

func (r *PostgresJobRepository) InsertJobError(ctx context.Context, jobError *models.JobError) error {
	query := `INSERT INTO job_errors (job_id, stage, record_index, error_message) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.ExecContext(ctx, query, jobError.JobID, jobError.Stage, jobError.RecordIndex, jobError.ErrorMessage)
	return err
}

func (r *PostgresJobRepository) InsertJobResult(ctx context.Context, result *models.JobResult) error {
	query := `INSERT INTO job_results (job_id, summary_json) VALUES ($1, $2)`
	_, err := r.DB.ExecContext(ctx, query, result.JobID, result.SummaryJSON)
	return err
}

func (r *PostgresJobRepository) GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error) {
	query := `SELECT id, job_id, stage, record_index, error_message, created_at FROM job_errors WHERE job_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorsList []models.JobError
	for rows.Next() {
		var e models.JobError
		if err := rows.Scan(&e.ID, &e.JobID, &e.Stage, &e.RecordIndex, &e.ErrorMessage, &e.CreatedAt); err != nil {
			return nil, err
		}
		errorsList = append(errorsList, e)
	}
	if errorsList == nil {
		errorsList = []models.JobError{}
	}
	return errorsList, nil
}

func (r *PostgresJobRepository) GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error) {
	query := `SELECT id, job_id, summary_json, created_at FROM job_results WHERE job_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.JobResult
	for rows.Next() {
		var res models.JobResult
		if err := rows.Scan(&res.ID, &res.JobID, &res.SummaryJSON, &res.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	if results == nil {
		results = []models.JobResult{}
	}
	return results, nil
}

func (r *PostgresJobRepository) UpdateJobStatus(ctx context.Context, id string, status string, finishedAt *time.Time) error {
	var query string
	var err error
	if finishedAt != nil {
		query = `UPDATE jobs SET status = $2, finished_at = $3 WHERE id = $1`
		_, err = r.DB.ExecContext(ctx, query, id, status, finishedAt)
	} else {
		query := `
			UPDATE jobs 
			SET status = $2, finished_at = COALESCE($3, finished_at) 
			WHERE id = $1 
			AND status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
		`
		_, err = r.DB.ExecContext(ctx, query, id, status)
	}

	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) UpdateJobMetrics(ctx context.Context, id string, metrics map[string]interface{}) error {
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	query := `UPDATE jobs SET metrics = $2 WHERE id = $1`
	_, err = r.DB.ExecContext(ctx, query, id, metricsBytes)
	if err != nil {
		return fmt.Errorf("failed to update job metrics: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) InsertExportedRecord(ctx context.Context, jobID string, sourceURL string, data map[string]any) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}
	query := `INSERT INTO job_exported_records (job_id, source_url, data) VALUES ($1, $2, $3)`
	_, err = r.DB.ExecContext(ctx, query, jobID, sourceURL, string(dataBytes))
	return err
}

func (r *PostgresJobRepository) InsertExportedRecordsBulk(ctx context.Context, jobID string, records []*PipelineRecord) error {
	if len(records) == 0 {
		return nil
	}

	query := `INSERT INTO job_exported_records (job_id, source_url, data) VALUES `
	args := make([]interface{}, 0, len(records)*3)

	for i, rec := range records {
		dataBytes, err := json.Marshal(rec.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal export data at index %d: %w", i, err)
		}

		query += fmt.Sprintf("($%d, $%d, $%d),", i*3+1, i*3+2, i*3+3)
		args = append(args, jobID, rec.SourceURL, string(dataBytes))
	}

	// Remove trailing comma
	query = query[:len(query)-1]

	_, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert exported records: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) GetExportedRecordsBySource(ctx context.Context, jobID string, sourceURL string) (*sql.Rows, error) {
	query := `SELECT data FROM job_exported_records WHERE job_id = $1 AND source_url = $2 ORDER BY created_at ASC, id ASC`
	return r.DB.QueryContext(ctx, query, jobID, sourceURL)
}

func (r *PostgresJobRepository) GetDistinctSources(ctx context.Context, jobID string) ([]string, error) {
	query := `SELECT DISTINCT source_url FROM job_exported_records WHERE job_id = $1`
	rows, err := r.DB.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var source string
		if err := rows.Scan(&source); err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func (r *PostgresJobRepository) DeleteExportedRecords(ctx context.Context, jobID string) error {
	query := `DELETE FROM job_exported_records WHERE job_id = $1`
	_, err := r.DB.ExecContext(ctx, query, jobID)
	return err
}

func (r *PostgresJobRepository) DeleteJob(ctx context.Context, id string) error {
	query := `DELETE FROM jobs WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresJobRepository) ListJobs(ctx context.Context, limit, offset int) ([]models.Job, int, error) {
	var total int
	// Optimize pagination count for massive tables using PostgreSQL internal statistics
	countQuery := `
		SELECT COALESCE(
			(SELECT reltuples::bigint FROM pg_class WHERE relname = 'jobs' AND reltuples > 1000),
			(SELECT COUNT(*) FROM jobs)
		)
	`
	err := r.DB.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, config, status, total_records, processed_records, error_count, created_at, finished_at, metrics 
		FROM jobs 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		var configBytes []byte
		var metricsBytes []byte
		var finishedAt sql.NullTime

		if err := rows.Scan(
			&job.ID,
			&configBytes,
			&job.Status,
			&job.TotalRecords,
			&job.ProcessedRecords,
			&job.ErrorCount,
			&job.CreatedAt,
			&finishedAt,
			&metricsBytes,
		); err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(configBytes, &job.Config); err != nil {
			continue
		}
		if len(metricsBytes) > 0 {
			_ = json.Unmarshal(metricsBytes, &job.Metrics)
		}
		if finishedAt.Valid {
			job.FinishedAt = &finishedAt.Time
		}
		jobs = append(jobs, job)
	}

	if jobs == nil {
		jobs = []models.Job{}
	}
	return jobs, total, nil
}
