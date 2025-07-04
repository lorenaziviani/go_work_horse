package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go_work_horse/pkg/jobqueue"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	jobsEnqueued = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_enqueued",
		Help: "Number of jobs currently in the queue.",
	})
	jobsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_processed_total",
		Help: "Total number of jobs processed.",
	})
	jobsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_failed_total",
		Help: "Total number of jobs failed.",
	})
	jobsRetried = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_retried_total",
		Help: "Total number of job retries.",
	})

	jobsPending = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_pending",
		Help: "Number of jobs with status pending.",
	})
	jobsRunning = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_running",
		Help: "Number of jobs with status running.",
	})
	jobsSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_success",
		Help: "Number of jobs with status success.",
	})
	jobsFailedGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_failed_gauge",
		Help: "Number of jobs with status failed.",
	})
)

func initMetrics() {
	prometheus.MustRegister(jobsEnqueued, jobsProcessed, jobsFailed, jobsRetried, jobsPending, jobsRunning, jobsSuccess, jobsFailedGauge)
	go func() {
		srv := &http.Server{
			Addr:         ":2112",
			Handler:      promhttp.Handler(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Error starting metrics server: %v", err)
		}
	}()
}

func initTracer() oteltrace.Tracer {
	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("jaeger:4318"), otlptracehttp.WithInsecure())
	if err != nil {
		log.Fatalf("Error creating OTLP exporter: %v", err)
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go_work_horse"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Tracer("go_work_horse")
}

func processJob(job *jobqueue.Job) error {
	start := time.Now()
	// Simulate job processing
	if os.Getenv("SIMULATE_FAIL") == "1" && job.RetryCount < 2 {
		err := fmt.Errorf("simulated error on job %s", job.ID)
		log.Printf("{\"job_id\":\"%s\",\"status\":\"failed\",\"duration_ms\":%d,\"error\":\"%s\"}", job.ID, time.Since(start).Milliseconds(), err.Error())
		return err
	}
	time.Sleep(2 * time.Second)
	log.Printf("{\"job_id\":\"%s\",\"status\":\"completed\",\"duration_ms\":%d}", job.ID, time.Since(start).Milliseconds())
	return nil
}

func main() {
	initMetrics()
	tracer := initTracer()
	cfg := jobqueue.LoadConfig()
	workerCount := cfg.MaxRetries // Reusing max_retries as example, ideally create a specific config
	if envWorkers := os.Getenv("WORKER_COUNT"); envWorkers != "" {
		if _, err := fmt.Sscanf(envWorkers, "%d", &workerCount); err != nil {
			fmt.Printf("Error converting WORKER_COUNT: %v\n", err)
		}
	}
	if workerCount <= 0 {
		workerCount = 5
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	queue := jobqueue.NewRedisQueue(redisAddr, "", 0, "jobs")

	// Update jobsEnqueued metric with the real size of the Redis queue every 2s
	go func() {
		for {
			size, err := queue.Length()
			if err == nil {
				jobsEnqueued.Set(float64(size))
			}
			time.Sleep(2 * time.Second)
		}
	}()

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
					_, jobSpan := tracer.Start(ctx, "process-job")
					job.Status = jobqueue.JobStatusRunning
					jobsPending.Dec()
					jobsRunning.Inc()
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
						jobsFailed.Inc()
						jobsRunning.Dec()
						jobsFailedGauge.Inc()
						if job.RetryCount <= maxRetries {
							jobsRetried.Inc()
							maxShift := 20
							shift := job.RetryCount - 1
							if shift > maxShift {
								shift = maxShift
							}
							if shift < 0 {
								shift = 0
							}
							backoff := time.Duration(retryDelay) * time.Second * time.Duration(1<<shift)
							log.Printf("{\"job_id\":\"%s\",\"retry\":%d,\"backoff_seconds\":%d}", job.ID, job.RetryCount, int(backoff.Seconds()))
							time.Sleep(backoff)
							_ = queue.Enqueue(job)
						}
					} else {
						job.Status = jobqueue.JobStatusSuccess
						job.LastError = ""
						jobsProcessed.Inc()
						jobsRunning.Dec()
						jobsSuccess.Inc()
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
					jobSpan.End()
				}
			}
		}(i)
	}
	wg.Wait()
}
