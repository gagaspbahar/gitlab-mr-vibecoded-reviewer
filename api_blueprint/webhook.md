- API description
  - GitLab webhook endpoint for Note events. Returns `202 Accepted` after validation and
    queues the review work for background processing.
- Parameters
  - Header: `X-Gitlab-Token` (optional when `gitlab_webhook_token` is set).
  - Body: GitLab Note event payload containing `project_id`, `merge_request.iid`,
    and `object_attributes.note`.
- Request sample(s)
  ```http
  POST /webhook HTTP/1.1
  Content-Type: application/json
  X-Gitlab-Token: <optional webhook secret>

  {
    "object_kind": "note",
    "project_id": 123,
    "merge_request": { "iid": 45 },
    "object_attributes": {
      "noteable_type": "MergeRequest",
      "note": "please review @review-bot"
    }
  }
  ```
- Response sample(s)
  ```http
  HTTP/1.1 202 Accepted
  ```
