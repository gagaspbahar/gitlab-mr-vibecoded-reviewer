- API description
  - Queue visibility endpoint. Returns current queue depth and in-flight job count.
- Parameters
  - None.
- Request sample(s)
  ```http
  GET /debug/queue HTTP/1.1
  ```
- Response sample(s)
  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json

  {
    "in_flight": 0,
    "queue_depth": 0
  }
  ```
