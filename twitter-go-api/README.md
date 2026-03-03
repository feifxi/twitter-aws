# Twitter Clone (Go API)

This is the high-performance, production-ready backend API for the Twitter/X clone, built directly in Go.

It was designed to replace the original Java/Spring Boot backend to provide significantly faster execution, lower memory footprint, and highly optimized database querying using modern Go patterns.

## 🚀 Tech Stack

- **Language:** Go 1.24+
- **Framework:** Gin Web Framework
- **Database:** PostgreSQL
- **Query Builder:** sqlc (Type-safe SQL code generation)
- **Redis:** Redis v9 (notification pub/sub + distributed rate limiting)
- **Logging:** Zerolog (Structured JSON Telemetry)
- **Authentication:** Custom JWT authentication (Tokens)
- **Storage:** Azure Blob Storage (for media uploads)
- **Architecture:** Clean Architecture (Handler -> Usecase -> Repository)

## ✨ Key Features

- **Layered Architecture:** HTTP handlers depend on explicit domain service interfaces (auth, user, tweet, feed, search, discovery, notification) to keep boundaries clear and testable.
- **Memory Batching (DataLoader Pattern):** Reduces N+1 query patterns in feed hydration by batching author and referenced tweet lookups.
- **Redis Pub/Sub for Notifications:** Supports cross-instance real-time notification fanout for SSE streams.
- **Security-Focused Upload Validation:** Validates media/avatar size, extension allowlists, and server-side MIME detection before usecase execution.
- **Structured Error Contract:** Maps validation/app/db errors into consistent API error responses.
- **Full Text Search:** Uses PostgreSQL `to_tsvector` / `ts_rank` for tweet search and hashtag prefix search.
- **Media Uploads:** Multipart file streaming to Azure Blob storage.

## 🛠️ Getting Started (Local Development)

### Prerequisites

- Go 1.24 or higher
- PostgreSQL running locally (can be started via Docker Compose in the project root)
- [sqlc](https://sqlc.dev/) (for database model generation)
- [golang-migrate](https://github.com/golang-migrate/migrate) (for database migrations)

### 1. Database Setup & Migrations

Ensure your PostgreSQL database `twitter_db` is running.
```bash
# Run database migrations
make migrateup
```

*Note: The database configuration expects `postgres://root:rootpass@localhost:5432/twitter_db?sslmode=disable` by default.*

### 2. Configuration

Create a `app.env` file in the root of `twitter-go-api` based on required environment variables:
```env
DATABASE_URL=postgresql://root:rootpass@localhost:5432/twitter_db?sslmode=disable
HTTP_SERVER_ADDRESS=0.0.0.0:8080
DB_MAX_CONNS=25
DB_MIN_CONNS=0
DB_MAX_CONN_LIFETIME_MINUTES=5
MAX_MULTIPART_MEMORY_BYTES=33554432
MAX_MEDIA_BYTES=104857600
MAX_AVATAR_BYTES=5242880
TRUSTED_PROXIES=
TOKEN_SYMMETRIC_KEY=replace-with-a-strong-32-plus-char-secret
TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=30
GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
AZURE_STORAGE_CONTAINER_NAME=tweet-media
AZURE_STORAGE_CONNECTION_STRING=your-azure-storage-connection-string
REDIS_ADDRESS=redis://localhost:6379
REDIS_PASSWORD=
```

### 3. Running the API

```bash
# Generate SQLC models (if you modify queries in db/query/*.sql)
make sqlc

# Run the API
make run
```

The API will be available at `http://localhost:8080`.

## 📂 Project Structure

- `/cmd/api` - The main application entry point.
- `/db` - Contains `migration` files and `query` SQL schemas. `sqlc` reads these to generate the Go database repository.
- `/internal/db` - The generated `sqlc` database interaction code. Do not edit manually.
- `/internal/server` - Gin HTTP handlers, routes, request parsing, and JSON response mapping.
- `/internal/usecase` - Domain business services (auth/user/tweet/feed/search/discovery/notification) and shared domain helpers.
- `/internal/config` - Viper configuration management.
- `/internal/token` - JWT token creation and verification.
- `/internal/service` - Third-party integrations (like Azure Blob Storage).

## ✅ Testing

- Unit and handler tests run with:
```bash
go test ./...
```
- Database integration tests for transaction semantics are included in `internal/db/store_integration_test.go`.
  - Set `TEST_DATABASE_URL` to run them.
  - If `TEST_DATABASE_URL` is not set, those integration tests are skipped.
