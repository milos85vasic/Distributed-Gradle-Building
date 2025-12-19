# Distributed Gradle Building - ML Service Dockerfile
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go/go.mod go/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY go/ ./

# Remove main.go from root to avoid conflict
RUN rm -f main.go

# Build the ML service binary
RUN cd ml && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ml-binary main.go && mv ml-binary ../

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create app user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ml-binary ./ml-service

# Copy entrypoint script
COPY docker/entrypoint-ml.sh /app/entrypoint-ml.sh
RUN chmod +x /app/entrypoint-ml.sh

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/config

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8082/health || exit 1

# Default command
CMD ["/app/entrypoint-ml.sh"]