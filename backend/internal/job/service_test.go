package job

import (
	"context"
	"testing"
	"time"
)

func TestPipelineService_CancelJob(t *testing.T) {
	repo := &MockJobRepository{}
	svc := NewPipelineService(repo)

	// mock an active job in the service maps
	ctx, cancel := context.WithCancel(context.Background())
	svc.activeJobs["job-1"] = cancel

	err := svc.CancelJob(context.Background(), "job-1")
	if err != nil {
		t.Errorf("Unexpected error cancelling active job: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Errorf("Expected context to be cancelled")
	}
}

func TestPipelineService_ListJobs(t *testing.T) {
	repo := &MockJobRepository{}
	svc := NewPipelineService(repo)

	// Safe pass-through check
	_, err := svc.ListJobs(context.Background())
	if err != nil {
		t.Errorf("Unexpected error listing jobs: %v", err)
	}
}
