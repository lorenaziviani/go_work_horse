// Para buildar: go build -o enqueue main.go
// Uso: REDIS_ADDR=localhost:6379 ./enqueue '{"foo":"bar"}'

package main

import (
	"encoding/json"
	"fmt"
	"go_work_horse/pkg/jobqueue"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: enqueue <payload>")
		os.Exit(1)
	}
	payload := os.Args[1]

	maxRetries := 3
	retryDelay := 5
	if v := os.Getenv("JOB_MAX_RETRIES"); v != "" {
		fmt.Sscanf(v, "%d", &maxRetries)
	}
	if v := os.Getenv("JOB_RETRY_DELAY"); v != "" {
		fmt.Sscanf(v, "%d", &retryDelay)
	}

	job := jobqueue.Job{
		ID:         uuid.NewString(),
		Payload:    []byte(payload),
		Status:     jobqueue.JobStatusPending,
		RetryCount: 0,
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	queue := jobqueue.NewRedisQueue(redisAddr, "", 0, "jobs")
	if err := queue.Enqueue(&job); err != nil {
		fmt.Println("Error enqueueing job:", err)
		os.Exit(1)
	}

	b, _ := json.MarshalIndent(job, "", "  ")
	fmt.Println("Job enqueued:")
	fmt.Println(string(b))
}
