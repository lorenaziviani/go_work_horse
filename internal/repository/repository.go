package repository

import "go_work_horse/pkg/jobqueue"

type Repository interface {
	SaveJob(job *jobqueue.Job) error
	GetJob(id string) (*jobqueue.Job, error)
	UpdateJob(job *jobqueue.Job) error
	ListJobs(status jobqueue.JobStatus) ([]*jobqueue.Job, error)
}
