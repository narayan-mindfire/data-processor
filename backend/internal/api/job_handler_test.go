package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
)

func TestGetJobHandler_Success(t *testing.T) {
	mockDB := store.NewMockStore()
	
	fakeJob := &models.Job{
		ID:         "123",
		SourceType: "csv",
		Status:     models.StatusCompleted,
		CreatedAt:  time.Now(),
	}
	mockDB.CreateJob(context.Background(), fakeJob)

	handler := GetJobHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines/123", nil)
	req.SetPathValue("id", "123") 

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rr.Code)
	}

	var responseJob models.Job
	if err := json.NewDecoder(rr.Body).Decode(&responseJob); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if responseJob.ID != "123" {
		t.Errorf("Expected job ID '123', got '%s'", responseJob.ID)
	}
}
