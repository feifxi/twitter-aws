# Twitter Go API

Primary backend API for this project.

## Stack

- Go 1.24+
- Gin
- PostgreSQL
- sqlc
- Redis (SSE fanout + distributed rate limiting)
- Zerolog
- Azure Blob Storage (media)

## Architecture

- `handler -> usecase -> db/service`
- Domain-oriented usecase services with explicit dependencies:
  - auth, user, tweet, feed, search, discovery, notification
- Shared error transport model in `internal/apiresponse`
- DTO/domain mapping separated from generated sqlc models

## API Behavior (Important)

- Pagination uses cursor model:
  - Request: `cursor`, `size`
  - Response: `items`, `hasNext`, `nextCursor`
- Upload handling validates:
  - max size
  - extension allowlist
  - server-side detected MIME type
- Request correlation:
  - every response includes `X-Request-ID`
  - error responses include `requestId` for log correlation
- Health endpoints:
  - `GET /healthz` (liveness)
  - `GET /readyz` (readiness: DB/Redis dependency check)

See API contract details:
- `docs/api-contract.md`

## Local Setup

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis
- sqlc
- golang-migrate

### Run

```bash
make migrateup
make run
```

Default API address: `http://localhost:8080`

### SQLC Regeneration

If you change queries under `db/query`:

```bash
make sqlc
```

## Testing

Run all tests:

```bash
go test ./...
```

## Production Operations Baseline

For Azure Container Apps:
- Liveness probe path: `/healthz`
- Readiness probe path: `/readyz`
- Health endpoints are intentionally outside API rate limiting

Monitoring baseline:
- Track HTTP 5xx rate, p95 latency, and restart count
- Create alert rules for sustained 5xx spikes and failing readiness probes
- Use `requestId` from API error response to correlate app logs for incident triage

Integration test note:
- `internal/db/store_integration_test.go` uses `TEST_DATABASE_URL`
- It is skipped when `TEST_DATABASE_URL` is not set

## Main Directories

- `cmd/api`: entrypoint
- `db/migration`: schema migrations
- `db/query`: sqlc query source
- `internal/db`: generated sqlc code
- `internal/server`: HTTP handlers/routes/response mapping
- `internal/usecase`: business logic services and domain helpers
- `internal/service`: external integrations (e.g., storage)
- `internal/config`: environment config
- `internal/token`: JWT logic
