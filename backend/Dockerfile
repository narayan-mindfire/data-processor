# Stage 1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy module definition and download dependencies (for layer caching)
COPY go.mod ./
RUN go mod download

# Copy application source code
COPY . .

# Build a static, lightweight binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api-server ./cmd/api

# Stage 2
FROM alpine:3.20

WORKDIR /app

# Install SSL certificates so our Go app can make HTTPS calls to external APIs
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]
