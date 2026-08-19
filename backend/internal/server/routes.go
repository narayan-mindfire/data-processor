package server

import (
	"net/http"

	_ "github.com/narayan-mindfire/data-processor/backend/docs"
	"github.com/narayan-mindfire/data-processor/backend/internal/job"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterRoutes initializes the master router for the application
func RegisterRoutes(svc job.JobService) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/api-docs/", httpSwagger.WrapHandler)

	mux.HandleFunc("POST /api/v1/pipelines", job.CreateJobHandler(svc))
	mux.HandleFunc("GET /api/v1/pipelines", job.ListJobsHandler(svc))
	mux.HandleFunc("GET /api/v1/pipelines/{id}", job.GetJobHandler(svc))
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", job.GetJobProgressHandler(svc))
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", job.GetJobResultsHandler(svc))
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", job.GetJobErrorsHandler(svc))
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", job.CancelJobHandler(svc))
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", job.DeleteJobHandler(svc))

	return mux
}
