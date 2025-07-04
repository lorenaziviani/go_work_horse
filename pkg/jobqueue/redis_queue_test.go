package jobqueue

import (
	"os"
	"testing"
	"time"
)

func getTestRedisQueue() *RedisQueue {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return NewRedisQueue(addr, "", 0, "test_jobs")
}

func TestRedisQueueEnqueueDequeue(t *testing.T) {
	queue := getTestRedisQueue()
	job := &Job{
		ID:         "job-test-1",
		Payload:    []byte(`{"foo":"bar"}`),
		Status:     JobStatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := queue.Enqueue(job)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}
	job2, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}
	if job2 == nil || job2.ID != job.ID {
		t.Errorf("expected job id %s, got %v", job.ID, job2)
	}
}
