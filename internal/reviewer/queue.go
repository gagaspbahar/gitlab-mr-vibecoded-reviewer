package reviewer

import (
	"context"
	"errors"
	"log"
	"time"
)

var ErrQueueFull = errors.New("queue full")

type Job struct {
	ProjectID int
	MRIID     int
	Note      string
}

type Queue struct {
	jobs    chan Job
	timeout time.Duration
}

func NewQueue(buffer int, timeout time.Duration) *Queue {
	if buffer <= 0 {
		buffer = 1
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &Queue{
		jobs:    make(chan Job, buffer),
		timeout: timeout,
	}
}

func (q *Queue) Enqueue(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) StartWorkers(concurrency int, reviewer *Reviewer) {
	if concurrency <= 0 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for job := range q.jobs {
				ctx, cancel := context.WithTimeout(context.Background(), q.timeout)
				if err := reviewer.Run(ctx, job.ProjectID, job.MRIID, job.Note); err != nil {
					log.Printf("review failed for project %d mr %d: %v", job.ProjectID, job.MRIID, err)
				}
				cancel()
			}
		}()
	}
}
