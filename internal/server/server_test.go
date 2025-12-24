package server

import "testing"

func TestShouldProcess(t *testing.T) {
	srv := New("token", "review-bot", nil)
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
