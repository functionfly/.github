#!/bin/bash

set -euo pipefail

echo "🧪 FunctionFly Test Coverage Analysis"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "FAIL")
            echo -e "${RED}❌ $message${NC}"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Check if we're in the project root
if [ ! -f "go.mod" ]; then
    print_status "FAIL" "Must be run from project root directory"
    exit 1
fi

# Run unit tests with coverage
echo ""
echo "Running unit tests..."
if go test -race -coverprofile=coverage-unit.out ./... > /dev/null 2>&1; then
    print_status "PASS" "Unit tests completed successfully"
else
    print_status "FAIL" "Unit tests failed"
    exit 1
fi

# Run integration tests with coverage
echo ""
echo "Running integration tests..."
if go test -race -tags=integration -coverprofile=coverage-integration.out ./... > /dev/null 2>&1; then
    print_status "PASS" "Integration tests completed successfully"
else
    print_status "WARN" "Integration tests failed (may require database setup)"
fi

# Calculate coverage percentages
UNIT_COVERAGE=$(go tool cover -func=coverage-unit.out | grep total | awk '{print substr($3, 1, length($3)-1)}' || echo "0")
INTEGRATION_COVERAGE=$(go tool cover -func=coverage-integration.out | grep total | awk '{print substr($3, 1, length($3)-1)}' || echo "0")

# Calculate combined coverage (weighted)
if [ "$INTEGRATION_COVERAGE" = "0" ]; then
    COMBINED_COVERAGE=$UNIT_COVERAGE
else
    COMBINED_COVERAGE=$(echo "scale=2; ($UNIT_COVERAGE * 0.7) + ($INTEGRATION_COVERAGE * 0.3)" | bc -l)
fi

echo ""
echo "📊 Coverage Results:"
echo "===================="
echo "Unit Test Coverage:      $UNIT_COVERAGE%"
echo "Integration Coverage:    $INTEGRATION_COVERAGE%"
echo "Combined Coverage:       $COMBINED_COVERAGE%"

# Check coverage threshold
THRESHOLD=80.0
if (( $(echo "$COMBINED_COVERAGE >= $THRESHOLD" | bc -l) )); then
    print_status "PASS" "Coverage meets threshold ($THRESHOLD%)"
else
    print_status "FAIL" "Coverage below threshold ($THRESHOLD%)"
fi

# Generate coverage reports
echo ""
echo "📄 Generating coverage reports..."
go tool cover -html=coverage-unit.out -o coverage-unit.html
echo "Unit test coverage report: coverage-unit.html"

if [ -f "coverage-integration.out" ] && [ "$INTEGRATION_COVERAGE" != "0" ]; then
    go tool cover -html=coverage-integration.out -o coverage-integration.html
    echo "Integration coverage report: coverage-integration.html"
fi

# Coverage by package
echo ""
echo "📦 Coverage by Package (Unit Tests):"
echo "===================================="
go tool cover -func=coverage-unit.out | head -n -1 | sort -k3 -nr

# Test statistics
echo ""
echo "📈 Test Statistics:"
echo "=================="

# Count test files
TEST_FILES=$(find . -name "*_test.go" | wc -l)
echo "Test files: $TEST_FILES"

# Count test functions (rough estimate)
UNIT_TESTS=$(grep -r "func Test" --include="*_test.go" . | wc -l)
INTEGRATION_TESTS=$(grep -r "func Test" --include="*integration_test.go" . | wc -l || echo "0")
echo "Unit test functions: $UNIT_TESTS"
echo "Integration test functions: $INTEGRATION_TESTS"

# Recommendations
echo ""
echo "💡 Recommendations:"
echo "=================="

if (( $(echo "$UNIT_COVERAGE < 85" | bc -l) )); then
    echo "- Consider adding more unit tests for better coverage"
fi

if (( $(echo "$INTEGRATION_COVERAGE < 70" | bc -l) )) && [ "$INTEGRATION_COVERAGE" != "0" ]; then
    echo "- Integration test coverage could be improved"
fi

if [ "$TEST_FILES" -lt 10 ]; then
    echo "- Consider adding more test files for comprehensive testing"
fi

echo ""
print_status "PASS" "Coverage analysis completed"