package models

import (
	"time"
)

// Pipeline Job Statuses
const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
)

// Job represents the metadata of a pipeline run
type Job struct {
	ID               string     `json:"id"`
	SourceType       string     `json:"source_type"` 
	Status           string     `json:"status"`
	TotalRecords     int        `json:"total_records"`
	ProcessedRecords int        `json:"processed_records"`
	ErrorCount       int        `json:"error_count"`
	CreatedAt        time.Time  `json:"created_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
}

// JobError represents a failed record during pipeline execution
type JobError struct {
	ID           int       `json:"id"`
	JobID        string    `json:"job_id"`
	Stage        string    `json:"stage"`
	RecordIndex  int       `json:"record_index"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// JobResult represents the final aggregated output
type JobResult struct {
	ID          int       `json:"id"`
	JobID       string    `json:"job_id"`
	SummaryJSON string    `json:"summary_json"`
	CreatedAt   time.Time `json:"created_at"`
}
