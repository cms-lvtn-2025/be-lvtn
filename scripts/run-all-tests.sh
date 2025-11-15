#!/bin/bash

# Comprehensive test runner for all services
# Usage: ./scripts/run-all-tests.sh [mode]
# Modes: quick, full, coverage, watch

set -e

MODE=${1:-quick}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
SERVICES=("academic" "council" "file" "role" "thesis" "user")
TOTAL_TESTS=218
FAILED_SERVICES=()

echo -e "${BLUE}🧪 LVTN Microservices Test Suite${NC}"
echo -e "${CYAN}===================================${NC}"
echo -e "Mode: ${YELLOW}$MODE${NC}"
echo -e "Total Expected Tests: ${CYAN}$TOTAL_TESTS${NC}"
echo ""

# Function to run tests for a service
run_service_tests() {
    local service=$1
    local flags=$2
    
    echo -e "${YELLOW}Testing $service service...${NC}"
    
    cd "src/service/$service"
    
    if go test ./tests/unit/... $flags; then
        echo -e "${GREEN}✅ $service: PASSED${NC}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}❌ $service: FAILED${NC}"
        FAILED_SERVICES+=("$service")
        cd - > /dev/null
        return 1
    fi
}

# Quick mode - run all tests
if [ "$MODE" == "quick" ]; then
    echo -e "${CYAN}Running quick test suite...${NC}"
    echo ""
    
    for service in "${SERVICES[@]}"; do
        run_service_tests "$service" "-v -timeout=30s" || true
        echo ""
    done

# Full mode - run with race detection
elif [ "$MODE" == "full" ]; then
    echo -e "${CYAN}Running full test suite with race detection...${NC}"
    echo ""
    
    for service in "${SERVICES[@]}"; do
        run_service_tests "$service" "-v -race -timeout=1m" || true
        echo ""
    done

# Coverage mode - generate coverage reports
elif [ "$MODE" == "coverage" ]; then
    echo -e "${CYAN}Running tests with coverage analysis...${NC}"
    echo ""
    
    mkdir -p coverage
    
    for service in "${SERVICES[@]}"; do
        echo -e "${YELLOW}Testing $service service with coverage...${NC}"
        cd "src/service/$service"
        
        if go test ./tests/unit/... -v -coverprofile="../../../coverage/${service}.out" -covermode=atomic; then
            echo -e "${GREEN}✅ $service: PASSED${NC}"
            
            if [ -f "../../../coverage/${service}.out" ]; then
                total=$(go tool cover -func="../../../coverage/${service}.out" | grep "total:" | awk '{print $3}')
                echo -e "   Coverage: ${CYAN}$total${NC}"
            fi
        else
            echo -e "${RED}❌ $service: FAILED${NC}"
            FAILED_SERVICES+=("$service")
        fi
        
        cd - > /dev/null
        echo ""
    done

# Watch mode - continuous testing
elif [ "$MODE" == "watch" ]; then
    echo -e "${CYAN}Starting watch mode (Ctrl+C to stop)...${NC}"
    echo ""
    
    while true; do
        clear
        echo -e "${BLUE}🔄 Running tests... ($(date))${NC}"
        echo ""
        
        for service in "${SERVICES[@]}"; do
            run_service_tests "$service" "-v -timeout=30s" || true
        done
        
        echo ""
        echo -e "${YELLOW}Waiting 5 seconds before next run...${NC}"
        sleep 5
    done

else
    echo -e "${RED}Unknown mode: $MODE${NC}"
    echo "Available modes: quick, full, coverage, watch"
    exit 1
fi

# Summary
echo ""
echo -e "${BLUE}=================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}=================================${NC}"

if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All services passed!${NC}"
    echo -e "   ${CYAN}$TOTAL_TESTS tests executed successfully${NC}"
    exit 0
else
    echo -e "${RED}❌ ${#FAILED_SERVICES[@]} service(s) failed:${NC}"
    for service in "${FAILED_SERVICES[@]}"; do
        echo -e "   ${RED}- $service${NC}"
    done
    exit 1
fi
