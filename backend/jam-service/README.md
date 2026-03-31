# Backend (Go)

Initial Go backend scaffold for the video streaming project.

## Structure

```text
backend/
├── cmd/server/main.go
├── internal/config/
├── internal/server/
├── .env.example
└── go.mod
```

## Run

```bash
go run ./cmd/server
```

Server starts on `http://localhost:8080` by default.

## Health Check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{
  "status": "ok",
  "service": "video-streaming-backend",
  "timestamp": "2026-01-01T00:00:00Z"
}
```

## Swagger API Docs

Open Swagger UI in your browser for interactive API testing:

```text
http://localhost:8080/swagger
```

Raw OpenAPI spec:

```text
http://localhost:8080/swagger/openapi.json
```

## Jams Queue API (MVP)

Queue command/read endpoints are served under:

```text
/api/v1/jams/{jamId}/queue/add       (POST)
/api/v1/jams/{jamId}/queue/remove    (POST)
/api/v1/jams/{jamId}/queue/reorder   (POST)
/api/v1/jams/{jamId}/queue/snapshot  (GET)
```

The service enforces:
- atomic queue mutations with monotonic `queueVersion`
- add idempotency via `idempotencyKey`
- `409 version_conflict` on stale reorder commands
- `409 session_ended` when write commands target ended sessions

## Jam Session Lifecycle API (MVP)

Session lifecycle endpoints are served under:

```text
/api/v1/jams/create             (POST)
/api/v1/jams/{jamId}/join       (POST)
/api/v1/jams/{jamId}/leave      (POST)
/api/v1/jams/{jamId}/end        (POST)
```

Lifecycle and authorization rules:
- create requires valid premium entitlement (`403 premium_required` otherwise)
- end requires session host ownership (`403 host_only` for non-host actors)
- host leave transitions session status to ended (`endCause=host_leave`)
- queue/playback write operations are rejected after session end with `409 session_ended`

## Rollout / Rollback (Feature Flag)

1. Deploy with queue endpoints disabled by feature flag.
2. Enable for internal traffic only and watch queue version conflict/idempotency metrics.
3. Gradually increase traffic percentage to the `jams` queue path.
4. Rollback by toggling the feature flag off and routing to previous queue path.

## Test

```bash
go test ./...
go test -race ./...
go vet ./...
```
