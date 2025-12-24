package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	GitLabBaseURL     string        `mapstructure:"gitlab_base_url"`
	GitLabToken       string        `mapstructure:"gitlab_token"`
	WebhookToken      string        `mapstructure:"gitlab_webhook_token"`
	BotUsername       string        `mapstructure:"bot_username"`
	ListenAddr        string        `mapstructure:"listen_addr"`
	LLMBaseURL        string        `mapstructure:"llm_base_url"`
	LLMAPIKey         string        `mapstructure:"llm_api_key"`
	LLMModel          string        `mapstructure:"llm_model"`
	HTTPTimeout       time.Duration `mapstructure:"http_timeout"`
	WorkerConcurrency int           `mapstructure:"worker_concurrency"`
	JobTimeout        time.Duration `mapstructure:"job_timeout"`
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetDefault("listen_addr", ":8080")
	v.SetDefault("llm_model", "internal-reviewer")
	v.SetDefault("http_timeout", "30s")
	v.SetDefault("worker_concurrency", 2)
	v.SetDefault("job_timeout", "5m")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.GitLabBaseURL == "" {
		return Config{}, fmt.Errorf("gitlab_base_url is required")
	}
	if cfg.GitLabToken == "" {
		return Config{}, fmt.Errorf("gitlab_token is required")
	}
	if cfg.BotUsername == "" {
		return Config{}, fmt.Errorf("bot_username is required")
	}
	if cfg.LLMBaseURL == "" {
		return Config{}, fmt.Errorf("llm_base_url is required")
	}
	if cfg.LLMAPIKey == "" {
		return Config{}, fmt.Errorf("llm_api_key is required")
	}
	if cfg.WorkerConcurrency <= 0 {
		return Config{}, fmt.Errorf("worker_concurrency must be positive")
	}
	if cfg.JobTimeout <= 0 {
		return Config{}, fmt.Errorf("job_timeout must be positive")
	}

	return cfg, nil
}
