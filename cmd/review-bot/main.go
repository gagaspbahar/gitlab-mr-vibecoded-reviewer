package main

import (
	"flag"
	"log"
	"net/http"

	"gitlab-mr-vibecoded-reviewer/internal/config"
	"gitlab-mr-vibecoded-reviewer/internal/gitlab"
	"gitlab-mr-vibecoded-reviewer/internal/llm"
	"gitlab-mr-vibecoded-reviewer/internal/reviewer"
	"gitlab-mr-vibecoded-reviewer/internal/server"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	gitlabClient, err := gitlab.NewClient(cfg.GitLabBaseURL, cfg.GitLabToken, cfg.HTTPTimeout)
	if err != nil {
		log.Fatalf("gitlab client error: %v", err)
	}
	llmClient, err := llm.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.HTTPTimeout)
	if err != nil {
		log.Fatalf("llm client error: %v", err)
	}

	reviewerService := reviewer.New(gitlabClient, llmClient)
	server := server.New(cfg.WebhookToken, cfg.BotUsername, reviewerService)

	log.Printf("review bot listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, server.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
