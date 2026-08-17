package api

import (
	"encoding/json"
	"net/http"

	"github.com/narayan-mindfire/data-processor/backend/internal/store"
	"github.com/narayan-mindfire/data-processor/backend/pkg/errors"
)

// GetJobHandler returns an http.HandlerFunc that fetches a job by ID.
func GetJobHandler(db store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.PathValue("id")
		if jobID == "" {
			http.Error(w, "missing job id", http.StatusBadRequest)
			return
		}

		job, err := db.GetJobByID(r.Context(), jobID)
		if err != nil {
			if err == errors.ErrNotFound {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Return 200 OK and the JSON data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(job)
	}
}
