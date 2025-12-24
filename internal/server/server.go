package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab-mr-vibecoded-reviewer/internal/reviewer"
)

type Server struct {
	webhookToken string
	botUsername  string
	reviewer     *reviewer.Reviewer
}

type NoteEvent struct {
	ObjectKind   string `json:"object_kind"`
	ProjectID    int    `json:"project_id"`
	MergeRequest struct {
		IID int `json:"iid"`
	} `json:"merge_request"`
	ObjectAttributes struct {
		Note         string `json:"note"`
		NoteableType string `json:"noteable_type"`
	} `json:"object_attributes"`
}

func New(webhookToken, botUsername string, reviewer *reviewer.Reviewer) *Server {
	return &Server{
		webhookToken: webhookToken,
		botUsername:  botUsername,
		reviewer:     reviewer,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var event NoteEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !s.shouldProcess(event) {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	if err := s.reviewer.Run(ctx, event.ProjectID, event.MergeRequest.IID, event.ObjectAttributes.Note); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("review failed: %s", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) authorized(r *http.Request) bool {
	if s.webhookToken == "" {
		return true
	}
	return r.Header.Get("X-Gitlab-Token") == s.webhookToken
}

func (s *Server) shouldProcess(event NoteEvent) bool {
	if event.ObjectKind != "note" {
		return false
	}
	if event.ObjectAttributes.NoteableType != "MergeRequest" {
		return false
	}
	if event.ProjectID == 0 || event.MergeRequest.IID == 0 {
		return false
	}
	return strings.Contains(event.ObjectAttributes.Note, "@"+s.botUsername)
}
