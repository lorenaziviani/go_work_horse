package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go_work_horse/pkg/jobqueue"
)

func processJob(job *jobqueue.Job) error {
	start := time.Now()
	// Simulate job processing
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
					err = processJob(job)
					if err != nil {
						job.Status = jobqueue.JobStatusFailed
					} else {
						job.Status = jobqueue.JobStatusSuccess
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
