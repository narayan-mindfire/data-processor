package controllers

import (
	"encoding/json"
	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
	"net/http"
)

// CreateJobRequest represents the payload to start a new pipeline
type CreateJobRequest struct {
	SourceType string `json:"source_type" example:"csv" enums:"csv,json,mixed"`
	SourceURL  string `json:"source_url" example:"https://covid.ourworldindata.org/data/owid-covid-data.csv"`
}

// MockResponse represents our temporary stub responses
type MockResponse struct {
	Message string `json:"message" example:"Endpoint hit successfully"`
}

// ErrorResponse represents a generic API error
type ErrorResponse struct {
	Error string `json:"error" example:"job not found"`
}

func sendMockJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(MockResponse{Message: message})
}

// --- API Handlers ---

// @Summary Start a new pipeline job
// @Description Ingests data from the specified source URL and begins concurrent processing.
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param request body CreateJobRequest true "Pipeline Configuration"
// @Success 201 {object} MockResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/pipelines [post]
func CreateJobHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Create pipeline job endpoint hit", http.StatusCreated)
	}
}

// @Summary List all pipeline jobs
// @Description Retrieves a list of all historical and running jobs
// @Tags Pipelines
// @Produce json
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines [get]
func ListJobsHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "List pipeline jobs endpoint hit", http.StatusOK)
	}
}

// @Summary Get pipeline job details
// @Description Fetches the metadata and status of a specific job by ID
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID (e.g. 1234-abcd)"
// @Success 200 {object} MockResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/pipelines/{id} [get]
func GetJobHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Get pipeline job endpoint hit for ID: "+r.PathValue("id"), http.StatusOK)
	}
}

// @Summary Get real-time job progress
// @Description Retrieves processing metrics (total vs processed records) for a running job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/progress [get]
func GetJobProgressHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Get job progress hit for ID: "+r.PathValue("id"), http.StatusOK)
	}
}

// @Summary Get final job results
// @Description Retrieves the aggregated mathematical outputs for a completed job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/results [get]
func GetJobResultsHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Get job results hit", http.StatusOK)
	}
}

// @Summary Get job error logs
// @Description Retrieves any failed records and error logs for a job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/errors [get]
func GetJobErrorsHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Get job errors hit", http.StatusOK)
	}
}

// @Summary Cancel running pipeline job
// @Description Safely cancels an active pipeline via context cancellation
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/cancel [patch]
func CancelJobHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Cancel job hit", http.StatusOK)
	}
}

// @Summary Delete job and artifacts
// @Description Removes a job and all associated artifacts from PostgreSQL
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id} [delete]
func DeleteJobHandler(db repositories.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "Delete job hit", http.StatusOK)
	}
}
