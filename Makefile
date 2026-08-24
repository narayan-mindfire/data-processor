# Data Processor Makefile

# Start the docker-compose stack and rebuild images
up:
	docker compose up --build -d

# Stop the docker-compose stack
down:
	docker compose down

# Run the test suite with coverage
test:
	docker run --rm -v "$(PWD)/backend:/app" -v dp-go-cache:/go -v dp-go-build-cache:/root/.cache -w /app golang:1.24-alpine go test -cover ./...

# Run the Go linter
lint:
	docker run --rm -v "$(PWD)/backend:/app" -v dp-lint-cache:/root/.cache -w /app golangci/golangci-lint:v1.64.5 golangci-lint run -v

# Generate Swagger Documentation
swagger:
	docker run --rm -v "$(PWD)/backend:/app" -v dp-go-cache:/go -w /app golang:1.24-alpine sh -c "go install github.com/swaggo/swag/cmd/swag@v1.16.2 && swag init -g cmd/api/main.go --parseDependency --parseInternal"

# Clean up Docker system (Optional)
clean:
	docker system prune -f
	docker volume rm dp-go-cache dp-go-build-cache dp-lint-cache || true

# Tidy Go modules
tidy:
	docker run --rm -v "$(PWD)/backend:/app" -v dp-go-cache:/go -w /app golang:1.24-alpine go mod tidy
