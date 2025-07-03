package jobqueue

import (
	"time"
)

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusSuccess JobStatus = "success"
	JobStatusFailed  JobStatus = "failed"
)

type Job struct {
	ID         string     `json:"id"`
	Payload    []byte     `json:"payload"`
	Status     JobStatus  `json:"status"`
	RetryCount int        `json:"retry_count"`
	ExecutedAt *time.Time `json:"executed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	RetryDelay int        `json:"retry_delay"` // in seconds
	MaxRetries int        `json:"max_retries"`
	LastError  string     `json:"last_error,omitempty"`
}
