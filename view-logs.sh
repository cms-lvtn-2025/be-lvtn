#!/bin/bash

# Script để xem logs của tất cả services và server
# Sử dụng: ./view-logs.sh [service-name]
# Ví dụ: ./view-logs.sh user-service
#        ./view-logs.sh main-server
#        ./view-logs.sh (xem tất cả)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$PROJECT_ROOT/logs"

# Check if logs directory exists
if [ ! -d "$LOG_DIR" ]; then
    echo -e "${RED}Thư mục logs không tồn tại!${NC}"
    echo -e "${YELLOW}Vui lòng chạy services trước.${NC}"
    exit 1
fi

# If a specific service is provided
if [ -n "$1" ]; then
    LOG_FILE="$LOG_DIR/$1.log"
    if [ -f "$LOG_FILE" ]; then
        echo -e "${BLUE}========================================${NC}"
        echo -e "${GREEN}  Logs của $1${NC}"
        echo -e "${BLUE}========================================${NC}\n"
        tail -f "$LOG_FILE"
    else
        echo -e "${RED}Log file không tồn tại: $LOG_FILE${NC}"
        echo -e "\n${YELLOW}Available logs:${NC}"
        ls -1 "$LOG_DIR"/*.log 2>/dev/null || echo "  Không có log files nào"
        exit 1
    fi
else
    # Show all logs
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}  Xem tất cả logs${NC}"
    echo -e "${BLUE}========================================${NC}\n"

    if ! command -v multitail &> /dev/null; then
        echo -e "${YELLOW}Sử dụng tail -f để xem logs...${NC}"
        echo -e "${YELLOW}(Cài đặt 'multitail' để có trải nghiệm tốt hơn)${NC}\n"
        tail -f "$LOG_DIR"/*.log
    else
        multitail "$LOG_DIR"/*.log
    fi
fi
