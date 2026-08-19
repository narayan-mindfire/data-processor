package job

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

type MockJobRepository struct{}

func (m *MockJobRepository) CreateJob(ctx context.Context, job *models.Job) error { return nil }
func (m *MockJobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	return nil, nil
}
func (m *MockJobRepository) UpdateJobStatus(ctx context.Context, id string, status string, finishedAt *time.Time) error {
	return nil
}
func (m *MockJobRepository) UpdateJobProgress(ctx context.Context, id string, processed, errors int) error {
	return nil
}
func (m *MockJobRepository) InsertJobResult(ctx context.Context, res *models.JobResult) error {
	return nil
}
func (m *MockJobRepository) InsertJobError(ctx context.Context, err *models.JobError) error {
	return nil
}
func (m *MockJobRepository) GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error) {
	return nil, nil
}
func (m *MockJobRepository) GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error) {
	return nil, nil
}
func (m *MockJobRepository) DeleteJob(ctx context.Context, id string) error     { return nil }
func (m *MockJobRepository) ListJobs(ctx context.Context) ([]models.Job, error) { return nil, nil }

func TestEngine_Validation(t *testing.T) {
	job := &models.Job{
		ID: "test-job-1",
		Config: models.JobConfig{
			Validations: []models.ValidationRule{
				{Field: "age", Rule: "is_numeric"},
			},
		},
	}

	engine := NewPipelineEngine(job, &MockJobRepository{})

	// Create a single test record
	record := &PipelineRecord{
		Index: 1,
		Data:  map[string]any{"age": "25"}, // Valid numeric string
	}

	engine.recordsCh <- record
	close(engine.recordsCh)

	var wg sync.WaitGroup
	wg.Add(1)
	go engine.validationWorker(context.Background(), &wg)
	wg.Wait()
	close(engine.validatedCh)

	validCount := 0
	for range engine.validatedCh {
		validCount++
	}

	if validCount != 1 {
		t.Errorf("Expected 1 valid record, got %d", validCount)
	}
}
