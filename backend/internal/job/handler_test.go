package job

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

// MockJobService to intercept HTTP Handler calls
type MockJobService struct{}

func (m *MockJobService) StartPipeline(ctx context.Context, job *models.Job) error { return nil }
func (m *MockJobService) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	return &models.Job{ID: id, Status: models.StatusRunning}, nil
}
func (m *MockJobService) GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error) {
	return nil, nil
}
func (m *MockJobService) GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error) {
	return nil, nil
}
func (m *MockJobService) GetExportURLs(ctx context.Context, jobID string, format string) ([]string, error) {
	return nil, nil
}
func (m *MockJobService) CancelJob(ctx context.Context, id string) error { return nil }
func (m *MockJobService) DeleteJob(ctx context.Context, id string) error { return nil }
func (m *MockJobService) ListJobs(ctx context.Context, limit, offset int) ([]models.Job, int, error) {
	return nil, 0, nil
}

func TestCreateJobHandler(t *testing.T) {
	svc := &MockJobService{}
	handler := CreateJobHandler(svc)

	// Valid JSON Body
	body := []byte(`{"concurrency": {"validation_workers": 2}}`)
	req := httptest.NewRequest("POST", "/api/v1/pipelines", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var responseJob models.Job
	if err := json.NewDecoder(rr.Body).Decode(&responseJob); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	if responseJob.ID == "" {
		t.Errorf("Expected job ID to be generated")
	}
}

func TestGetJobProgressHandler(t *testing.T) {
	svc := &MockJobService{}
	handler := GetJobProgressHandler(svc)

	req := httptest.NewRequest("GET", "/api/v1/pipelines/test-id/progress", nil)
	req.SetPathValue("id", "test-id")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var progress ProgressResponse
	if err := json.NewDecoder(rr.Body).Decode(&progress); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	if progress.Status != models.StatusRunning {
		t.Errorf("Expected status %s, got %s", models.StatusRunning, progress.Status)
	}
}

func TestCancelJobHandler(t *testing.T) {
	svc := &MockJobService{}
	handler := CancelJobHandler(svc)

	req := httptest.NewRequest("PATCH", "/api/v1/pipelines/test-id/cancel", nil)
	req.SetPathValue("id", "test-id")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestAllHandlers(t *testing.T) {
	svc := &MockJobService{}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
	}{
		{"GetJob", GetJobHandler(svc), "GET", "/api/v1/pipelines/test-id"},
		{"GetResults", GetJobResultsHandler(svc), "GET", "/api/v1/pipelines/test-id/results"},
		{"GetErrors", GetJobErrorsHandler(svc), "GET", "/api/v1/pipelines/test-id/errors"},
		{"DeleteJob", DeleteJobHandler(svc), "DELETE", "/api/v1/pipelines/test-id"},
		{"ListJobs", ListJobsHandler(svc), "GET", "/api/v1/pipelines"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.SetPathValue("id", "test-id")
			rr := httptest.NewRecorder()
			tt.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("%s returned wrong status code: got %v want %v", tt.name, rr.Code, http.StatusOK)
			}
		})
	}
}
