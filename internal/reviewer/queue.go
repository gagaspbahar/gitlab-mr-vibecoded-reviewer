package reviewer

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

var ErrQueueFull = errors.New("queue full")

type Job struct {
	ProjectID int
	MRIID     int
	Note      string
}

type Queue struct {
	jobs     chan Job
	timeout  time.Duration
	inFlight atomic.Int64
}

type QueueStats struct {
	InFlight   int64 `json:"in_flight"`
	QueueDepth int   `json:"queue_depth"`
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
		log.Printf("queue enqueue ok project_id=%d mr_iid=%d note_len=%d queue_depth=%d", job.ProjectID, job.MRIID, len(job.Note), len(q.jobs))
		return nil
	default:
		log.Printf("queue enqueue failed project_id=%d mr_iid=%d note_len=%d error=%v", job.ProjectID, job.MRIID, len(job.Note), ErrQueueFull)
		return ErrQueueFull
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) QueueStats() QueueStats {
	return QueueStats{
		InFlight:   q.inFlight.Load(),
		QueueDepth: len(q.jobs),
	}
}

func (q *Queue) StartWorkers(concurrency int, reviewer *Reviewer) {
	if concurrency <= 0 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for job := range q.jobs {
				q.inFlight.Add(1)
				start := time.Now()
				log.Printf("queue job start project_id=%d mr_iid=%d note_len=%d", job.ProjectID, job.MRIID, len(job.Note))
				ctx, cancel := context.WithTimeout(context.Background(), q.timeout)
				err := reviewer.Run(ctx, job.ProjectID, job.MRIID, job.Note)
				cancel()
				duration := time.Since(start)
				if err != nil {
					log.Printf("queue job finish project_id=%d mr_iid=%d note_len=%d duration=%s status=error error=%v", job.ProjectID, job.MRIID, len(job.Note), duration, err)
				} else {
					log.Printf("queue job finish project_id=%d mr_iid=%d note_len=%d duration=%s status=ok", job.ProjectID, job.MRIID, len(job.Note), duration)
				}
				q.inFlight.Add(-1)
			}
		}()
	}
}
