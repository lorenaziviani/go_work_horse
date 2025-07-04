package jobqueue

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJobCreationDefaults(t *testing.T) {
	now := time.Now()
	job := Job{
		ID:         "test-id",
		Payload:    []byte(`{"foo":"bar"}`),
		Status:     JobStatusPending,
		RetryCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if job.Status != JobStatusPending {
		t.Errorf("expected status pending, got %v", job.Status)
	}
	if job.RetryCount != 0 {
		t.Errorf("expected retry_count 0, got %d", job.RetryCount)
	}
}

func TestJobJSONSerialization(t *testing.T) {
	job := Job{
		ID:         "test-id",
		Payload:    []byte(`{"foo":"bar"}`),
		Status:     JobStatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	b, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal job: %v", err)
	}
	var job2 Job
	if err := json.Unmarshal(b, &job2); err != nil {
		t.Fatalf("failed to unmarshal job: %v", err)
	}
	if job2.ID != job.ID {
		t.Errorf("expected id %s, got %s", job.ID, job2.ID)
	}
}
