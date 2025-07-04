package integration

import (
	"go_work_horse/pkg/jobqueue"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestEndToEndJobProcessing(t *testing.T) {
	// Enqueue a job directly in the Redis queue
	queue := jobqueue.NewRedisQueue("localhost:6379", "", 0, "jobs")
	job := &jobqueue.Job{
		ID:         "integration-job-1",
		Payload:    []byte(`{"foo":"bar"}`),
		Status:     jobqueue.JobStatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := queue.Enqueue(job); err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// Wait for the worker to process the job
	success := false
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://localhost:2112/metrics")
		if err != nil {
			t.Logf("waiting for metrics endpoint...")
			time.Sleep(2 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if strings.Contains(string(body), "jobs_processed_total 1") {
			success = true
			break
		}
		time.Sleep(2 * time.Second)
	}
	if !success {
		t.Errorf("job was not processed and metrics not updated")
	}
}
