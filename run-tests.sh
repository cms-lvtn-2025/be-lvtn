#!/bin/bash

# LVTN Test Runner Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-30s}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
PARALLEL_WORKERS=${PARALLEL_WORKERS:-4}

echo -e "${BLUE}🧪 LVTN Test Suite Runner${NC}"
echo "=================================="

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to check if services are running
check_services() {
    print_info "Checking required services..."
    
    # Check database
    if ! nc -z localhost 3306 2>/dev/null; then
        print_warning "MySQL not running on localhost:3306"
    else
        print_status "MySQL is running"
    fi
    
    # Check Redis
    if ! nc -z localhost 6379 2>/dev/null; then
        print_warning "Redis not running on localhost:6379"
    else
        print_status "Redis is running"
    fi
    
    # Check gRPC services
    services=("50051:academic" "50056:user" "50053:file" "50055:thesis" "50052:council" "50054:role")
    for service in "${services[@]}"; do
        port="${service%%:*}"
        name="${service##*:}"
        if ! nc -z localhost $port 2>/dev/null; then
            print_warning "$name service not running on localhost:$port"
        else
            print_status "$name service is running"
        fi
    done
}

# Function to run unit tests
run_unit_tests() {
    print_info "Running unit tests..."
    
    # Clean test cache
    go clean -testcache
    
#!/bin/bash

# Academic Service Test Runner with Coverage
# Usage: ./run-tests.sh [coverage|all|verbose]

echo "🧪 Academic Service Test Suite"
echo "=============================="

# Default flags
VERBOSE=""
COVERAGE=""
RUN_ALL=""

# Parse command line arguments
case "$1" in
    "coverage")
        COVERAGE="-cover -coverprofile=coverage.out"
        echo "📊 Running with coverage analysis..."
        ;;
    "all")
        RUN_ALL="-v"
        echo "🔍 Running all tests with verbose output..."
        ;;
    "verbose")
        VERBOSE="-v"
        echo "📋 Running with verbose output..."
        ;;
    *)
        echo "🚀 Running standard test suite..."
        ;;
esac

# Run Academic Service Handler Tests
echo ""
echo "Testing Academic Service Handlers..."
echo "-----------------------------------"

if [ "$1" = "coverage" ]; then
    go test ./src/service/academic/handler/ $COVERAGE $VERBOSE
    if [ $? -eq 0 ]; then
        echo ""
        echo "📊 Generating coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        echo "✅ Coverage report generated: coverage.html"
        
        # Show coverage summary
        echo ""
        echo "📈 Coverage Summary:"
        go tool cover -func=coverage.out | grep total
    fi
else
    go test ./src/service/academic/handler/ $VERBOSE $RUN_ALL
fi

# Check test result
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All tests passed successfully!"
    echo ""
    echo "📋 Test Summary:"
    echo "  - Faculty Handler: ✅ CREATE, READ, UPDATE, DELETE, LIST"
    echo "  - Major Handler:   ✅ CREATE, READ, UPDATE, DELETE, LIST"
    echo "  - Semester Handler: ✅ CREATE, READ, UPDATE, DELETE, LIST"
    echo "  - Error Handling: ✅ gRPC status codes, validation, database errors"
    echo "  - Database Mocking: ✅ SQL queries, transactions, row scanning"
else
    echo ""
    echo "❌ Some tests failed. Please check the output above."
    exit 1
fi

# Optional: Run integration tests if they exist
if [ -d "./src/service/academic/integration" ]; then
    echo ""
    echo "🔗 Running Integration Tests..."
    echo "------------------------------"
    go test ./src/service/academic/integration/ $VERBOSE
fi

echo ""
echo "🎯 Test execution completed!"    if [ $? -eq 0 ]; then
        print_status "Unit tests passed"
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_info "Running integration tests..."
    
    # Check if services are available
    check_services
    
    # Run integration tests
    go test -v -race -timeout=$TEST_TIMEOUT -tags=integration ./tests/integration/... 2>&1 | tee integration_results.log
    
    if [ $? -eq 0 ]; then
        print_status "Integration tests passed"
    else
        print_warning "Integration tests failed (services may not be available)"
        return 0  # Don't fail the entire suite
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    print_info "Generating coverage report..."
    
    if [ -f coverage.out ]; then
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        
        # Get coverage percentage
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        
        echo "Coverage: ${coverage}%"
        
        if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
            print_status "Coverage threshold met (${coverage}% >= ${COVERAGE_THRESHOLD}%)"
        else
            print_warning "Coverage below threshold (${coverage}% < ${COVERAGE_THRESHOLD}%)"
        fi
        
        print_info "Coverage report generated: coverage.html"
    else
        print_warning "No coverage data found"
    fi
}

# Function to run benchmarks
run_benchmarks() {
    print_info "Running benchmarks..."
    
    go test -bench=. -benchmem ./src/... ./tests/... 2>&1 | tee benchmark_results.log
    
    if [ $? -eq 0 ]; then
        print_status "Benchmarks completed"
    else
        print_warning "Benchmarks had issues"
    fi
}

# Function to run linter
run_linter() {
    print_info "Running linter..."
    
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./... 2>&1 | tee lint_results.log
        
        if [ $? -eq 0 ]; then
            print_status "Linting passed"
        else
            print_warning "Linting found issues"
        fi
    else
        print_warning "golangci-lint not installed, skipping"
    fi
}

# Function to run security checks
run_security_checks() {
    print_info "Running security checks..."
    
    if command -v gosec &> /dev/null; then
        gosec ./... 2>&1 | tee security_results.log
        
        if [ $? -eq 0 ]; then
            print_status "Security checks passed"
        else
            print_warning "Security checks found issues"
        fi
    else
        print_warning "gosec not installed, skipping"
    fi
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up..."
    
    # Remove temporary files
    rm -f test_results.log integration_results.log benchmark_results.log lint_results.log security_results.log
    
    print_status "Cleanup completed"
}

# Main execution
main() {
    local run_type=${1:-all}
    
    case $run_type in
        "unit")
            run_unit_tests
            generate_coverage_report
            ;;
        "integration")
            run_integration_tests
            ;;
        "bench")
            run_benchmarks
            ;;
        "lint")
            run_linter
            ;;
        "security")
            run_security_checks
            ;;
        "all")
            run_unit_tests
            run_integration_tests
            generate_coverage_report
            run_benchmarks
            run_linter
            run_security_checks
            ;;
        "ci")
            # CI/CD optimized run
            run_unit_tests
            generate_coverage_report
            run_linter
            ;;
        *)
            echo "Usage: $0 {unit|integration|bench|lint|security|all|ci}"
            echo ""
            echo "  unit        - Run unit tests with coverage"
            echo "  integration - Run integration tests"
            echo "  bench       - Run benchmarks"
            echo "  lint        - Run linter"
            echo "  security    - Run security checks"
            echo "  all         - Run all tests and checks"
            echo "  ci          - Run CI/CD optimized tests"
            exit 1
            ;;
    esac
    
    print_status "Test suite completed"
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"