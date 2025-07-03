package jobqueue

type Queue interface {
	Enqueue(job *Job) error
	Dequeue() (*Job, error)
	Acknowledge(jobID string) error
	Requeue(job *Job) error
}
