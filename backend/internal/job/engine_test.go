package job

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

// MockJobRepository simulates the database for unit testing
type MockJobRepository struct {
	errors  int
	results int
	mu      sync.Mutex
}

func (m *MockJobRepository) CreateJob(ctx context.Context, job *models.Job) error { return nil }
func (m *MockJobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	return nil, nil
}
func (m *MockJobRepository) UpdateJobStatus(ctx context.Context, id string, status string, finishedAt *time.Time) error {
	return nil
}
func (m *MockJobRepository) UpdateJobMetrics(ctx context.Context, id string, metrics map[string]interface{}) error {
	return nil
}
func (m *MockJobRepository) UpdateJobProgress(ctx context.Context, id string, processed, errors int) error {
	return nil
}
func (m *MockJobRepository) InsertJobResult(ctx context.Context, res *models.JobResult) error {
	m.mu.Lock()
	m.results++
	m.mu.Unlock()
	return nil
}
func (m *MockJobRepository) InsertJobError(ctx context.Context, err *models.JobError) error {
	m.mu.Lock()
	m.errors++
	m.mu.Unlock()
	return nil
}
func (m *MockJobRepository) GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error) {
	return nil, nil
}
func (m *MockJobRepository) GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error) {
	return nil, nil
}
func (m *MockJobRepository) InsertExportedRecord(ctx context.Context, jobID string, sourceURL string, data map[string]any) error {
	return nil
}
func (m *MockJobRepository) GetExportedRecordsBySource(ctx context.Context, jobID string, sourceURL string) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockJobRepository) GetDistinctSources(ctx context.Context, jobID string) ([]string, error) {
	return nil, nil
}
func (m *MockJobRepository) DeleteExportedRecords(ctx context.Context, jobID string) error {
	return nil
}

func (m *MockJobRepository) DeleteJob(ctx context.Context, id string) error { return nil }
func (m *MockJobRepository) ListJobs(ctx context.Context, limit, offset int) ([]models.Job, int, error) {
	return nil, 0, nil
}

func TestEngine_ValidationAndTransformation(t *testing.T) {
	job := &models.Job{
		ID: "test-job-1",
		Config: models.JobConfig{
			Validations: []models.ValidationRule{
				{Field: "age", Rule: "is_numeric"},
			},
			Transformations: []models.TransformationRule{
				{Field: "age", Action: "convert_to_float"},
				{Field: "status", Action: "fill_empty", DefaultValue: "active"},
			},
		},
	}

	engine := NewPipelineEngine(job, &MockJobRepository{}, slog.Default())

	// Push Valid Record
	engine.recordsCh <- &PipelineRecord{Index: 1, Data: map[string]any{"age": "25"}}
	// Push Invalid Record (Text)
	engine.recordsCh <- &PipelineRecord{Index: 2, Data: map[string]any{"age": "twenty-five"}}
	close(engine.recordsCh)

	var wg sync.WaitGroup
	wg.Add(1)
	go engine.validationWorker(context.Background(), &wg)
	wg.Wait()
	close(engine.validatedCh)

	wg.Add(1)
	go engine.transformationWorker(context.Background(), &wg)
	wg.Wait()
	close(engine.transformedCh)

	validCount := 0
	for rec := range engine.transformedCh {
		validCount++
		if _, ok := rec.Data["age"].(float64); !ok {
			t.Errorf("Expected age to be float64 after transformation")
		}
		if rec.Data["status"] != "active" {
			t.Errorf("Expected empty status to be filled with 'active'")
		}
	}

	if validCount != 1 {
		t.Errorf("Expected 1 valid record, got %d", validCount)
	}
}

func TestEngine_Aggregation(t *testing.T) {
	job := &models.Job{
		ID: "test-job-2",
		Config: models.JobConfig{
			Aggregations: []models.AggregationDef{
				{Type: "sum", Field: "amount", OutputName: "total_amount"},
				{Type: "count", Field: "user", OutputName: "total_users"},
			},
		},
	}

	repo := &MockJobRepository{}
	engine := NewPipelineEngine(job, repo, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.transformedCh <- &PipelineRecord{Index: 1, Data: map[string]any{"amount": 100.0, "user": "alice"}}
	engine.transformedCh <- &PipelineRecord{Index: 2, Data: map[string]any{"amount": 50.0, "user": "bob"}}
	close(engine.transformedCh)

	// Since aggregate saves to the repo, we verify it through the mock repo
	var wg sync.WaitGroup
	wg.Add(1)
	go engine.exportWorker(ctx, &wg)

	engine.aggregate(ctx)
	wg.Wait()

	if repo.results != 1 {
		t.Errorf("Expected exactly 1 job result saved, got %d", repo.results)
	}
}

func TestEngine_IngestJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"id": 1, "name": "Alice"}]`)
	}))
	defer ts.Close()

	job := &models.Job{ID: "test-job", Config: models.JobConfig{}}
	engine := NewPipelineEngine(job, &MockJobRepository{}, slog.Default())

	var wg sync.WaitGroup
	wg.Add(1)

	source := models.SourceDef{Type: "json", URL: ts.URL}
	go engine.ingestJSON(context.Background(), source, &wg)

	wg.Wait()
	close(engine.recordsCh)

	count := 0
	for range engine.recordsCh {
		count++
	}
	if count != 1 {
		t.Errorf("Expected 1 record ingested from fake JSON API, got %d", count)
	}
}
