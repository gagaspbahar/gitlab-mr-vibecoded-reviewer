# gitlab-mr-vibecoded-reviewer

A GitLab merge request review bot written in Go. Mention `@review-bot` on a merge request to trigger:

- A summary of the merge request
- Inline review comments (when line numbers map to the diff) with a fallback summary note

## Features

- Webhook-driven MR note handling
- OpenAI-compatible LLM API integration
- Inline discussion creation for review suggestions
- Summary note posted to the MR

## Configuration

Create a config file (YAML or JSON) and pass it via `-config`.

Example `config.yaml`:

```yaml
gitlab_base_url: "https://gitlab.example.com"
gitlab_token: "<token>"
gitlab_webhook_token: "<optional webhook secret>"
bot_username: "review-bot"
llm_base_url: "https://llm.example.com"
llm_api_key: "<api key>"
llm_model: "internal-reviewer"
listen_addr: ":8080"
http_timeout: "30s"
worker_concurrency: 2
job_timeout: "5m"
```

`worker_concurrency` controls how many background reviewers run in parallel, and `job_timeout`
sets the per-review timeout for each queued job.

## Running

```bash
go run ./cmd/review-bot -config config.yaml
```

## API

### `POST /webhook`

GitLab webhook endpoint for Note events. Returns `202 Accepted` after validation and
queues the review work for background processing.

### `GET /healthz`

Health check endpoint. Returns `200 OK` with `ok`.

### `GET /debug/queue`

Queue visibility endpoint. Returns JSON with the current queue depth and in-flight job count.

Example response:

```json
{
  "in_flight": 0,
  "queue_depth": 0
}
```

## Webhook setup

Configure a GitLab webhook that sends **Note events** to `http://<host>:8080/webhook`.
If you set `gitlab_webhook_token`, add it as the webhook secret token in GitLab.
