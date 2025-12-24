package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`gitlab_base_url: "https://gitlab.example.com"
` +
		`gitlab_token: "token"
` +
		`bot_username: "review-bot"
` +
		`llm_base_url: "https://llm.example.com"
` +
		`llm_api_key: "key"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected config load to succeed: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Fatalf("expected default listen addr, got %q", cfg.ListenAddr)
	}
	if cfg.LLMModel != "internal-reviewer" {
		t.Fatalf("expected default model, got %q", cfg.LLMModel)
	}
}

func TestLoadConfigMissingRequired(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`gitlab_base_url: "https://gitlab.example.com"`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing required fields")
	}
}
