package job

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/pkg/apperrors"
)

// Issue 3: Service Layer Boundary. The Handler calls the Service, not the Repository directly!
type JobService interface {
	StartPipeline(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
}

type CreateJobRequest struct {
	SourceType string `json:"source_type" example:"csv" enums:"csv,json,mixed"`
	SourceURL  string `json:"source_url" example:"https://covid.ourworldindata.org/data/owid-covid-data.csv"`
}

type MockResponse struct {
	Message string `json:"message" example:"Endpoint hit successfully"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"job not found"`
}

func sendMockJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(MockResponse{Message: message})
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// @Summary Start a new pipeline job
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param request body CreateJobRequest true "Pipeline Configuration"
// @Success 201 {object} models.Job
// @Router /api/v1/pipelines [post]
func CreateJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON body"})
			return
		}

		job := &models.Job{
			ID:         generateUUID(),
			SourceType: req.SourceType,
			Status:     models.StatusPending,
			CreatedAt:  time.Now(),
		}

		if err := svc.StartPipeline(r.Context(), job); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to start pipeline"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(job)
	}
}

// @Summary Get pipeline job details
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.Job
// @Router /api/v1/pipelines/{id} [get]
func GetJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		job, err := svc.GetJobByID(r.Context(), id)
		if err != nil {
			if err == apperrors.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Job not found"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch job"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(job)
	}
}

// @Summary List all pipeline jobs
// @Description Retrieves a list of all historical and running jobs
// @Tags Pipelines
// @Produce json
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines [get]
func ListJobsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendMockJSON(w, "List pipeline jobs endpoint hit", http.StatusOK)
	}
}

// @Summary Get real-time job progress
// @Description Retrieves processing metrics (total vs processed records) for a running job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/progress [get]
func GetJobProgressHandler(svc JobService) http.HandlerFunc {
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
func GetJobResultsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { sendMockJSON(w, "Get job results hit", http.StatusOK) }
}

// @Summary Get job error logs
// @Description Retrieves any failed records and error logs for a job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/errors [get]
func GetJobErrorsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { sendMockJSON(w, "Get job errors hit", http.StatusOK) }
}

// @Summary Cancel running pipeline job
// @Description Safely cancels an active pipeline via context cancellation
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/cancel [patch]
func CancelJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { sendMockJSON(w, "Cancel job hit", http.StatusOK) }
}

// @Summary Delete job and artifacts
// @Description Removes a job and all associated artifacts from PostgreSQL
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id} [delete]
func DeleteJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { sendMockJSON(w, "Delete job hit", http.StatusOK) }
}
