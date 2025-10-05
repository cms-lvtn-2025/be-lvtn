FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /council-service ./src/service/council

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /council-service .

# Create env directory and copy env file
RUN mkdir -p env
COPY env/council.env env/council.env

# Expose port
EXPOSE 50052

# Run the service
CMD ["./council-service"]
