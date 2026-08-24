package job

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/models"
)

func TestPipelineService_CancelJob(t *testing.T) {
	repo := &MockJobRepository{}
	svc := NewPipelineService(repo, nil, slog.Default())

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
	svc := NewPipelineService(repo, nil, slog.Default())

	// Safe pass-through check
	_, _, err := svc.ListJobs(context.Background(), 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing jobs: %v", err)
	}
}

func TestPipelineService_PassThroughs(t *testing.T) {
	repo := &MockJobRepository{}
	svc := NewPipelineService(repo, nil, slog.Default())
	ctx := context.Background()

	_, _ = svc.GetJobByID(ctx, "test")
	_, _ = svc.GetJobErrors(ctx, "test")
	_, _ = svc.GetJobResults(ctx, "test")
	_ = svc.DeleteJob(ctx, "test")
	_ = svc.StartPipeline(ctx, &models.Job{ID: "test-start"})
}
