package job

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
	"github.com/narayan-mindfire/data-processor/backend/pkg/apperrors"
)

type JobService interface {
	StartPipeline(ctx context.Context, job *models.Job) error
	GetJobByID(ctx context.Context, id string) (*models.Job, error)
	GetJobErrors(ctx context.Context, jobID string) ([]models.JobError, error)
	GetJobResults(ctx context.Context, jobID string) ([]models.JobResult, error)
	GetExportedRecords(ctx context.Context, jobID string) (*sql.Rows, error)
	CancelJob(ctx context.Context, id string) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context) ([]models.Job, error)
}

type MockResponse struct {
	Message string `json:"message" example:"Endpoint hit successfully"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"job not found"`
}

// ProgressResponse defines the structured output required by the assignment
type ProgressResponse struct {
	Status           string     `json:"status"`
	PercentComplete  float64    `json:"percent_complete"`
	ProcessedRecords int        `json:"processed_records"`
	ErrorCount       int        `json:"error_count"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time,omitempty"`
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
// @Param request body models.JobConfig true "Pipeline Configuration"
// @Success 201 {object} models.Job
// @Router /api/v1/pipelines [post]
func CreateJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config models.JobConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON body"})
			return
		}

		job := &models.Job{
			ID:        generateUUID(),
			Config:    config,
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
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
// @Tags Pipelines
// @Produce json
// @Success 200 {array} models.Job
// @Router /api/v1/pipelines [get]
func ListJobsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := svc.ListJobs(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch jobs"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(jobs)
	}
}

// @Summary Get real-time job progress
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} ProgressResponse
// @Router /api/v1/pipelines/{id}/progress [get]
func GetJobProgressHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		job, err := svc.GetJobByID(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Job not found"})
			return
		}

		percent := 0.0
		if job.TotalRecords > 0 {
			percent = (float64(job.ProcessedRecords) / float64(job.TotalRecords)) * 100.0
		} else if job.Status == models.StatusCompleted {
			percent = 100.0
		}

		resp := ProgressResponse{
			Status:           job.Status,
			PercentComplete:  percent,
			ProcessedRecords: job.ProcessedRecords,
			ErrorCount:       job.ErrorCount,
			StartTime:        job.CreatedAt,
			EndTime:          job.FinishedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// @Summary Get final job results
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {array} models.JobResult
// @Router /api/v1/pipelines/{id}/results [get]
func GetJobResultsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		results, err := svc.GetJobResults(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch results"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(results)
	}
}

// @Summary Get job error logs
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {array} models.JobError
// @Router /api/v1/pipelines/{id}/errors [get]
func GetJobErrorsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		errorsList, err := svc.GetJobErrors(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch errors"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(errorsList)
	}
}

// @Summary Cancel running pipeline job
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id}/cancel [patch]
func CancelJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := svc.CancelJob(r.Context(), id); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(MockResponse{Message: "Job cancellation signal sent successfully"})
	}
}

// @Summary Delete job and artifacts
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} MockResponse
// @Router /api/v1/pipelines/{id} [delete]
func DeleteJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// If it's running, cancel it first before deleting
		_ = svc.CancelJob(r.Context(), id)

		if err := svc.DeleteJob(r.Context(), id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete job"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(MockResponse{Message: "Job deleted successfully"})
	}
}

// @Summary Export job processed records as JSON stream
// @Tags Pipelines
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {string} string "JSON stream"
// @Router /api/v1/pipelines/{id}/export/json [get]
func ExportJobJSONHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		rows, err := svc.GetExportedRecords(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch exported records"})
			return
		}
		defer rows.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"job_%s_export.json\"", id))
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("["))
		first := true
		for rows.Next() {
			var dataBytes []byte
			if err := rows.Scan(&dataBytes); err != nil {
				continue
			}
			if !first {
				_, _ = w.Write([]byte(","))
			}
			_, _ = w.Write(dataBytes)
			first = false
		}
		_, _ = w.Write([]byte("]"))
	}
}

// @Summary Export job processed records as CSV stream
// @Tags Pipelines
// @Produce text/csv
// @Param id path string true "Job ID"
// @Success 200 {string} string "CSV stream"
// @Router /api/v1/pipelines/{id}/export/csv [get]
func ExportJobCSVHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		rows, err := svc.GetExportedRecords(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch exported records"})
			return
		}
		defer rows.Close()

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"job_%s_export.csv\"", id))
		w.WriteHeader(http.StatusOK)

		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		headersWritten := false
		var headers []string

		for rows.Next() {
			var dataBytes []byte
			if err := rows.Scan(&dataBytes); err != nil {
				continue
			}
			var record map[string]any
			if err := json.Unmarshal(dataBytes, &record); err != nil {
				continue
			}

			if !headersWritten {
				for k := range record {
					headers = append(headers, k)
				}
				sort.Strings(headers) // Ensure consistent column order
				if err := csvWriter.Write(headers); err != nil {
					return
				}
				headersWritten = true
			}

			var row []string
			for _, h := range headers {
				val := record[h]
				row = append(row, fmt.Sprintf("%v", val))
			}
			if err := csvWriter.Write(row); err != nil {
				return
			}
		}
	}
}
