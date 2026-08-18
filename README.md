# Data Processor Pipeline

A high-performance, concurrent data ingestion and processing pipeline built in Go.

## Features

- **Concurrent Processing:** Uses Go routines and channels to process massive datasets rapidly.
- **REST API:** Fully decoupled MVC architecture.
- **PostgreSQL Database:** Embedded SQL migrations and robust connection pooling.
- **Pure Docker Tooling:** Run tests, linting, and the entire application without installing Go locally!

## Quick Start

Because we use a Pure Docker philosophy, all you need is Docker installed on your machine.

### 1. Start the Database & API:

```bash
docker compose up --build
```

View Interactive API Documentation (Swagger): Open your browser to: `http://localhost:8080/api-docs/index.html`

## 🛠 Architecture

This project follows strict Package-Oriented Design (MVC-style):

- `cmd/api/` - The entry point and main HTTP router.
- `internal/api/routes/` - Domain-specific route registration.
- `internal/api/controllers/` - HTTP request parsing and response handling.
- `internal/api/repositories/` - Database abstractions and SQL queries.