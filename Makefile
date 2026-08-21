.PHONY: up down build test lint swagger clean

# Start the application
up:
	docker compose up -d

# Stop the application
down:
	docker compose down

# Rebuild and start the application
build:
	docker compose up --build -d

# Run the test suite with coverage
test:
	docker run --rm -v "$(PWD)/backend:/app" -w /app golang:1.23-alpine sh -c "GOPROXY=https://goproxy.io,direct go test -cover ./..."

# Run the Go linter
lint:
	docker run --rm -v "$(PWD)/backend:/app" -w /app golangci/golangci-lint:v1.60.1 golangci-lint run -v

# Generate Swagger Documentation
swagger:
	docker run --rm -v "$(PWD)/backend:/app" -w /app golang:1.23-alpine sh -c "GOPROXY=https://goproxy.io,direct go install github.com/swaggo/swag/cmd/swag@v1.16.2 && swag init -g cmd/api/main.go --parseDependency --parseInternal"

# Clean up Docker system (Optional)
clean:
	docker system prune -f
