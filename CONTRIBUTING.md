# Contributing Guidelines

Welcome to the Data Processor repository! We enforce strict code quality gates to keep the pipeline robust and performant.

## Prerequisites

Only **Docker** is required. All Go tooling (builds, tests, linting, Swagger generation) runs inside containers — no local Go installation needed.

## Getting Started

### 1. Setup Git Hooks (Mandatory)

Our pre-commit hooks spin up temporary Docker containers to run tests and linters before allowing a commit. Activate them once after cloning:

```bash
git config core.hooksPath .githooks
```

### 2. Start the Development Environment

```bash
cp .env.example .env
docker compose up --build
```

The API server starts on `http://localhost:8080` and migrations run automatically.

## Development Workflow

### Managing Dependencies

Do **not** run `go get` locally. Use the Dockerised toolchain:

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.24-alpine go mod tidy
```

### Running Tests

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.24-alpine go test ./...
```

### Running the Linter

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golangci/golangci-lint:v1.64.5 golangci-lint run ./...
```

### Regenerating Swagger Docs

Swagger is auto-generated during `docker compose up --build`. To regenerate manually:

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.24-alpine sh -c \
  "go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/main.go"
```

## Architecture Rules

This project uses **domain-driven packaging** with interface-based dependency injection. Understanding these rules is essential before contributing.

### Package Structure

```
internal/
├── job/           Domain package (handler + service + repository co-located)
├── server/        HTTP wiring and route registration
├── models/        Shared domain types
├── pipeline/      Concurrent processing engine
└── store/         Database connection and migrations
```

### Dependency Direction

Dependencies must always point **inward** via interfaces:

```
Handler  →  defines JobService interface   →  consumed by PipelineService
Service  →  defines JobRepository interface →  consumed by PostgresJobRepository
```

- **Handlers** deal with HTTP: parsing requests, encoding responses, and setting status codes. They call the `JobService` interface — never a repository directly.
- **Services** contain business logic and orchestration. They call the `JobRepository` interface.
- **Repositories** execute raw SQL. They must never import the `handler` or `service` layers.
- **Interfaces are defined by the consumer**, not the implementer. This is a core Go idiom.

### Database

- All SQL queries must live in `internal/job/repository.go` (or the relevant domain package's repository file).
- Schema changes must be written as numbered migration files in `internal/store/migrations/` (e.g., `000002_add_index.up.sql` / `000002_add_index.down.sql`).
- Never write raw SQL in handlers or services.

### API Endpoints

- All new endpoints must include Swagger annotations (`@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`).
- Route registration happens in `internal/server/routes.go`.
- Handler functions live in the relevant domain package (e.g., `internal/job/handler.go`).

### Error Handling

- Use sentinel errors from `pkg/apperrors/` for well-known failure types (`ErrNotFound`, `ErrInvalidInput`, etc.).
- Always compare errors with `errors.Is()`, not `==`, to support wrapped errors.
- Use `fmt.Errorf("context: %w", err)` to wrap errors with context as they propagate up layers.

### Logging

- Use `log/slog` (Go's stdlib structured logger). Do **not** use `fmt.Println` or `log.Println`.
- The logger is created in `main.go` and should be injected into services and repositories via constructors.

### Concurrency

- All heavy data processing must be delegated to the `pipeline/` package using goroutines and channels.
- Never spawn goroutines inside handlers. The service layer orchestrates pipeline execution.

### Testing

- **Handler tests**: Use `httptest` with a mock `JobService` to test HTTP request/response behaviour.
- **Service tests**: Unit test business logic with a mock `JobRepository`.
- **Repository tests**: Integration test against a real PostgreSQL instance (or use `DATA-DOG/go-sqlmock`).
- Test files live alongside the code they test (e.g., `internal/job/handler_test.go`).

## Commit Messages

Use clear, descriptive commit messages:

```
feat(api): add job cancellation endpoint
fix(config): handle nil pointer on missing finished_at
refactor(config): move interfaces to consumer packages
test(handler): add handler tests for CreateJobHandler
refactor(docs): update architecture diagram in README
```

## Questions?

Open an issue or reach out. We're happy to help you get started!