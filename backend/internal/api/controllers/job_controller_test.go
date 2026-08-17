package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
)

func TestGetJobHandler_Success(t *testing.T) {
	mockRepo := repositories.NewMockRepository()

	handler := GetJobHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines/123", nil)
	req.SetPathValue("id", "123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	expectedMessage := "Get pipeline job endpoint hit for ID: 123"
	if response["message"] != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response["message"])
	}
}
