FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /thesis-service ./src/service/thesis

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /thesis-service .

# Copy certs (needed for TLS)
COPY certs /app/certs

# Create necessary directories
RUN mkdir -p /app/logs /app/env

# Set environment variable for project root
ENV PROJECT_ROOT=/app

# Expose port
EXPOSE 50055

# Run the service
CMD ["./thesis-service"]
