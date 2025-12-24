package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab-mr-vibecoded-reviewer/internal/reviewer"
)

func TestShouldProcess(t *testing.T) {
	srv := New("token", "review-bot", reviewer.NewQueue(1, time.Minute))
	event := NoteEvent{
		ObjectKind: "note",
		ProjectID:  1,
		MergeRequest: struct {
			IID int `json:"iid"`
		}{IID: 2},
		ObjectAttributes: struct {
			Note         string `json:"note"`
			NoteableType string `json:"noteable_type"`
		}{
			Note:         "please review @review-bot",
			NoteableType: "MergeRequest",
		},
	}
	if !srv.shouldProcess(event) {
		t.Fatal("expected event to be processed")
	}
}

func TestHandleWebhookAccepted(t *testing.T) {
	queue := reviewer.NewQueue(1, time.Minute)
	srv := New("token", "review-bot", queue)
	event := NoteEvent{
		ObjectKind: "note",
		ProjectID:  42,
		MergeRequest: struct {
			IID int `json:"iid"`
		}{IID: 7},
		ObjectAttributes: struct {
			Note         string `json:"note"`
			NoteableType string `json:"noteable_type"`
		}{
			Note:         "please review @review-bot",
			NoteableType: "MergeRequest",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-Gitlab-Token", "token")
	rec := httptest.NewRecorder()

	srv.handleWebhook(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	select {
	case job := <-queue.Jobs():
		if job.ProjectID != 42 || job.MRIID != 7 {
			t.Fatalf("unexpected job: %+v", job)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job")
	}
}

func TestHandleQueueStats(t *testing.T) {
	queue := reviewer.NewQueue(2, time.Minute)
	srv := New("token", "review-bot", queue)
	if err := queue.Enqueue(reviewer.Job{ProjectID: 1, MRIID: 2, Note: "note"}); err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/queue", nil)
	rec := httptest.NewRecorder()

	srv.handleQueueStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload reviewer.QueueStats
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if payload.QueueDepth != 1 {
		t.Fatalf("expected queue_depth 1, got %d", payload.QueueDepth)
	}
}
