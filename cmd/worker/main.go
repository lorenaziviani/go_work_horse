package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"go_work_horse/pkg/jobqueue"
)

func processJob(job *jobqueue.Job) error {
	start := time.Now()
	// Simulate job processing
	if os.Getenv("SIMULATE_FAIL") == "1" && job.RetryCount < 2 {
		err := fmt.Errorf("simulated error on job %s", job.ID)
		log.Printf("{\"job_id\":\"%s\",\"status\":\"failed\",\"duration_ms\":%d,\"error\":\"%s\",\"stack\":\"%s\"}", job.ID, time.Since(start).Milliseconds(), err.Error(), string(debug.Stack()))
		return err
	}
	time.Sleep(2 * time.Second)
	log.Printf("{\"job_id\":\"%s\",\"status\":\"completed\",\"duration_ms\":%d}", job.ID, time.Since(start).Milliseconds())
	return nil
}

func main() {
	cfg := jobqueue.LoadConfig()
	workerCount := cfg.MaxRetries // Reusing max_retries as example, ideally create a specific config
	if envWorkers := os.Getenv("WORKER_COUNT"); envWorkers != "" {
		fmt.Sscanf(envWorkers, "%d", &workerCount)
	}
	if workerCount <= 0 {
		workerCount = 5
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	queue := jobqueue.NewRedisQueue(redisAddr, "", 0, "jobs")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down workers...")
		cancel()
	}()

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					job, err := queue.Dequeue()
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					}
					if job == nil {
						time.Sleep(1 * time.Second)
						continue
					}
					job.Status = jobqueue.JobStatusRunning
					start := time.Now()
					maxRetries := job.MaxRetries
					if maxRetries == 0 {
						maxRetries = 3
					}
					retryDelay := job.RetryDelay
					if retryDelay == 0 {
						retryDelay = 5
					}
					err = processJob(job)
					if err != nil {
						job.Status = jobqueue.JobStatusFailed
						job.LastError = err.Error()
						job.RetryCount++
						if job.RetryCount <= maxRetries {
							backoff := time.Duration(retryDelay) * time.Second * time.Duration(1<<uint(job.RetryCount-1))
							log.Printf("{\"job_id\":\"%s\",\"retry\":%d,\"backoff_seconds\":%d}", job.ID, job.RetryCount, int(backoff.Seconds()))
							time.Sleep(backoff)
							_ = queue.Enqueue(job)
						}
					} else {
						job.Status = jobqueue.JobStatusSuccess
						job.LastError = ""
					}
					job.UpdatedAt = time.Now()
					dur := time.Since(start)
					b, _ := json.Marshal(map[string]interface{}{
						"worker":      workerID,
						"job_id":      job.ID,
						"status":      job.Status,
						"duration_ms": dur.Milliseconds(),
						"updated_at":  job.UpdatedAt,
					})
					log.Println(string(b))
				}
			}
		}(i)
	}
	wg.Wait()
}
