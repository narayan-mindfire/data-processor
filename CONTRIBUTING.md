# Contributing Guidelines

Welcome to the Data Processor team! We enforce strict code quality gates to ensure our pipeline remains robust and performant.

## Local Development (Pure Docker)

You do not need to install Go on your local machine to contribute to this project. We utilize isolated Docker containers for all tooling.

### 1. Setup Git Hooks (Mandatory)

Before you can push any code, you must activate our custom pre-commit hooks. These hooks will spin up temporary Docker containers to run our tests and Linters before allowing a commit.

Run this command once after cloning the repository:

```bash
git config core.hooksPath .githooks
```

### 2. Managing Dependencies

If you add a new library to the codebase, do not try to run `go get` locally. Instead, use our Docker toolchain to update the module cleanly:

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.23-alpine go mod tidy
```

### 3. Manual Testing

If you want to run the tests manually without committing, use our Dockerized test runner:

```bash
docker run --rm -v "$(pwd)/backend:/app" -w /app golang:1.23-alpine go test ./...
```

### 4. Architecture Rules

- **Database:** Do not write SQL queries in Controllers. All database logic must live in `repositories/`. Migrations must be written in `internal/store/migrations/`.
- **API:** Do not add business logic to Routes. Always include Swagger annotations (`@Summary`, `@Tags`) when adding a new API endpoint.
- **Concurrency:** All heavy data processing must be delegated to the `services/` layer using goroutines and channels.