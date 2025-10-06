#!/bin/bash

# Directories to watch
services=(
  "academic:50051"
  "council:50052"
  "file:50053"
  "role:50054"
  "thesis:50055"
  "user:50056"
)

cleanup() {
  echo "Stopping all services..."
  jobs -p | xargs -r kill
  exit 0
}

trap cleanup SIGINT SIGTERM

echo "Starting all services with hot reload..."

# Start each service in background
for service_info in "${services[@]}"; do
  IFS=':' read -r service port <<< "$service_info"
  echo "Starting $service on port $port..."
  (
    cd "src/service/$service" || exit
    air -c ../../../air/.air-$service.toml
  ) &
done

echo "All services started. Press Ctrl+C to stop all."
wait
