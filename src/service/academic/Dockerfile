# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o academic-service ./src/service/academic

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for TLS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/academic-service .

# Create directories for mounted volumes
RUN mkdir -p /app/service /app/logs

# Expose service port
EXPOSE 50051

# Run the service
CMD ["./academic-service"]
