#!/bin/bash

# Generate comprehensive coverage report
# Usage: ./scripts/coverage-report.sh [service]

set -e

SERVICE=${1:-all}
COVERAGE_DIR="coverage"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}📊 Generating Coverage Reports${NC}"
echo "=================================="

# Create coverage directory
mkdir -p $COVERAGE_DIR

# Function to generate coverage for a service
generate_service_coverage() {
    local service=$1
    echo -e "${YELLOW}Processing $service service...${NC}"
    
    cd "src/service/$service"
    
    # Run tests with coverage
    go test ./tests/unit/... -coverprofile="../../../$COVERAGE_DIR/${service}_coverage.out" \
        -covermode=atomic -v
    
    # Generate HTML report
    go tool cover -html="../../../$COVERAGE_DIR/${service}_coverage.out" \
        -o="../../../$COVERAGE_DIR/${service}_coverage.html"
    
    # Generate function-level coverage
    go tool cover -func="../../../$COVERAGE_DIR/${service}_coverage.out" \
        > "../../../$COVERAGE_DIR/${service}_coverage.txt"
    
    # Extract total coverage
    total_coverage=$(go tool cover -func="../../../$COVERAGE_DIR/${service}_coverage.out" | \
        grep "total:" | awk '{print $3}')
    
    echo -e "${GREEN}✅ $service: $total_coverage coverage${NC}"
    
    cd - > /dev/null
}

# Generate coverage for specific service or all
if [ "$SERVICE" == "all" ]; then
    services=("academic" "council" "file" "role" "thesis" "user")
    
    echo "" > "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "Coverage Summary - $TIMESTAMP" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "=============================" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    
    for service in "${services[@]}"; do
        generate_service_coverage "$service"
        
        # Add to summary
        total=$(tail -n 1 "$COVERAGE_DIR/${service}_coverage.txt" | awk '{print $3}')
        printf "%-15s: %s\n" "$service" "$total" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    done
    
    echo ""
    echo -e "${BLUE}📋 Coverage Summary:${NC}"
    cat "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    
else
    generate_service_coverage "$SERVICE"
fi

echo ""
echo -e "${GREEN}✅ Coverage reports generated in $COVERAGE_DIR/${NC}"
echo -e "   Open ${CYAN}$COVERAGE_DIR/*_coverage.html${NC} in your browser"
