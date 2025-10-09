#!/bin/bash

# Script để chỉ chạy main server (GraphQL/REST API)
# Sử dụng: ./run-server.sh

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
echo -e "${GREEN}  Khởi động Main Server với Air${NC}"
echo -e "${BLUE}========================================${NC}"

echo -e "\n${YELLOW}Đang khởi động GraphQL/REST API Server...${NC}\n"

# Run the main server
cd "$PROJECT_ROOT"
air -c "$AIR_DIR/.air-server.toml"
