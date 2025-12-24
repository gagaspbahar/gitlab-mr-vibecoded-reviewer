package reviewer

import (
	"testing"
	"time"
)

func TestQueueEnqueueDequeue(t *testing.T) {
	queue := NewQueue(1, time.Minute)
	job := Job{ProjectID: 1, MRIID: 2, Note: "please review"}

	if err := queue.Enqueue(job); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	select {
	case got := <-queue.Jobs():
		if got != job {
			t.Fatalf("unexpected job: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job")
	}
}
