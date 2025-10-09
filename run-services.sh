#!/bin/bash

# Script để chỉ chạy các microservices (không chạy server)
# Sử dụng: ./run-services.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AIR_DIR="$PROJECT_ROOT/air"

# Array to store background process IDs
PIDS=()

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Đang dừng tất cả services...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    wait
    echo -e "${GREEN}Đã dừng tất cả services!${NC}"
    exit 0
}

# Trap Ctrl+C and cleanup
trap cleanup SIGINT SIGTERM

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo -e "${RED}Air chưa được cài đặt!${NC}"
    echo -e "${YELLOW}Cài đặt Air bằng lệnh:${NC}"
    echo "go install github.com/air-verse/air@latest"
    exit 1
fi

# Create log directory if not exists
mkdir -p "$PROJECT_ROOT/logs"

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}  Khởi động Microservices với Air${NC}"
echo -e "${BLUE}========================================${NC}"

# Function to run a service with Air
run_service() {
    local service_name=$1
    local config_file=$2
    local service_dir=$3

    echo -e "${GREEN}▶ Đang khởi động ${service_name}...${NC}"

    cd "$service_dir"
    air -c "$config_file" > "$PROJECT_ROOT/logs/${service_name}.log" 2>&1 &
    PIDS+=($!)

    cd "$PROJECT_ROOT"
    sleep 0.5
}

# Run all microservices
echo -e "\n${YELLOW}[1/6] Khởi động User Service...${NC}"
run_service "user-service" "$AIR_DIR/.air-user.toml" "$PROJECT_ROOT/src/service/user"

echo -e "${YELLOW}[2/6] Khởi động Academic Service...${NC}"
run_service "academic-service" "$AIR_DIR/.air-academic.toml" "$PROJECT_ROOT/src/service/academic"

echo -e "${YELLOW}[3/6] Khởi động Council Service...${NC}"
run_service "council-service" "$AIR_DIR/.air-council.toml" "$PROJECT_ROOT/src/service/council"

echo -e "${YELLOW}[4/6] Khởi động File Service...${NC}"
run_service "file-service" "$AIR_DIR/.air-file.toml" "$PROJECT_ROOT/src/service/file"

echo -e "${YELLOW}[5/6] Khởi động Role Service...${NC}"
run_service "role-service" "$AIR_DIR/.air-role.toml" "$PROJECT_ROOT/src/service/role"

echo -e "${YELLOW}[6/6] Khởi động Thesis Service...${NC}"
run_service "thesis-service" "$AIR_DIR/.air-thesis.toml" "$PROJECT_ROOT/src/service/thesis"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Tất cả microservices đã được khởi động!${NC}"
echo -e "${BLUE}========================================${NC}"

echo -e "\n${YELLOW}Logs:${NC}"
echo -e "  • User Service:     ${PROJECT_ROOT}/logs/user-service.log"
echo -e "  • Academic Service: ${PROJECT_ROOT}/logs/academic-service.log"
echo -e "  • Council Service:  ${PROJECT_ROOT}/logs/council-service.log"
echo -e "  • File Service:     ${PROJECT_ROOT}/logs/file-service.log"
echo -e "  • Role Service:     ${PROJECT_ROOT}/logs/role-service.log"
echo -e "  • Thesis Service:   ${PROJECT_ROOT}/logs/thesis-service.log"

echo -e "\n${YELLOW}Để xem logs real-time:${NC}"
echo -e "  tail -f ${PROJECT_ROOT}/logs/*-service.log"

echo -e "\n${YELLOW}Nhấn Ctrl+C để dừng tất cả services${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Wait for all background processes
wait
