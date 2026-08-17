package routes

import (
	"net/http"

	"github.com/narayan-mindfire/data-processor/backend/internal/api/controllers"
	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
)

// RegisterJobRoutes attaches Job endpoints to the main router under a specific base path.
func RegisterJobRoutes(mux *http.ServeMux, basePath string, db repositories.JobRepository) {
	mux.HandleFunc("POST "+basePath, controllers.CreateJobHandler(db))
	mux.HandleFunc("GET "+basePath, controllers.ListJobsHandler(db))
	mux.HandleFunc("GET "+basePath+"/{id}", controllers.GetJobHandler(db))
	mux.HandleFunc("GET "+basePath+"/{id}/progress", controllers.GetJobProgressHandler(db))
	mux.HandleFunc("GET "+basePath+"/{id}/results", controllers.GetJobResultsHandler(db))
	mux.HandleFunc("GET "+basePath+"/{id}/errors", controllers.GetJobErrorsHandler(db))
	mux.HandleFunc("PATCH "+basePath+"/{id}/cancel", controllers.CancelJobHandler(db))
	mux.HandleFunc("DELETE "+basePath+"/{id}", controllers.DeleteJobHandler(db))
}
