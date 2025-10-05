FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /role-service ./src/service/role

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /role-service .

# Create env directory and copy env file
RUN mkdir -p env
COPY env/role.env env/role.env

# Expose port
EXPOSE 50054

# Run the service
CMD ["./role-service"]
