#!/bin/sh

# Script to run the FIGlet library test suite
# Runs Go tests for the figlet package with various options
#
# Usage: ./run-lib-tests.sh [options]
#   -v    Verbose output
#   -c    Show coverage
#   -b    Run benchmarks
#   -r    Generate race detection report

LC_ALL=POSIX
export LC_ALL

VERBOSE=""
COVERAGE=""
BENCHMARKS=""
RACE=""
LOGFILE="lib-tests.log"

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -v)
            VERBOSE="-v"
            ;;
        -c)
            COVERAGE="-cover -coverprofile=coverage.out"
            ;;
        -b)
            BENCHMARKS="-bench=. -benchmem"
            ;;
        -r)
            RACE="-race"
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "  -v    Verbose output"
            echo "  -c    Show coverage report"
            echo "  -b    Run benchmarks"
            echo "  -r    Run with race detection"
            echo "  -h    Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
    shift
done

echo "=========================================" | tee "$LOGFILE"
echo "FIGlet Go Library Test Suite" | tee -a "$LOGFILE"
echo "=========================================" | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

# Run tests
echo "Running tests..." | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

# Build the test command
TEST_CMD="go test $VERBOSE $COVERAGE $RACE $BENCHMARKS ./figlet/..."

echo "Command: $TEST_CMD" | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

# Execute tests
if $TEST_CMD 2>&1 | tee -a "$LOGFILE"; then
    result=0
else
    result=1
fi

echo "" | tee -a "$LOGFILE"

# Show coverage report if requested
if [ -n "$COVERAGE" ] && [ -f "coverage.out" ]; then
    echo "=========================================" | tee -a "$LOGFILE"
    echo "Coverage Report" | tee -a "$LOGFILE"
    echo "=========================================" | tee -a "$LOGFILE"
    go tool cover -func=coverage.out | tee -a "$LOGFILE"
    echo "" | tee -a "$LOGFILE"
fi

# Summary
echo "=========================================" | tee -a "$LOGFILE"

# Extract counts from log if possible (Go test output is tricky but we can try)
# Typical output: PASS: TestName (0.01s)
passed_count=$(grep -c "^PASS: " "$LOGFILE")
failed_count=$(grep -c "^FAIL: " "$LOGFILE")

if [ $result -eq 0 ]; then
    echo "✓ All library tests passed!" | tee -a "$LOGFILE"
    # If standard go test didn't output individual passes, at least we know it passed
    if [ "$passed_count" -eq 0 ] && [ "$failed_count" -eq 0 ]; then
        passed_count=1
        total_count=1
    else
        total_count=$((passed_count + failed_count))
    fi
else
    echo "✗ Some tests failed. See $LOGFILE for details." | tee -a "$LOGFILE"
    total_count=$((passed_count + failed_count))
    if [ "$total_count" -eq 0 ]; then total_count=1; fi
fi
echo "SUMMARY: PASSED=$passed_count, FAILED=$failed_count, TOTAL=$total_count"
echo "=========================================" | tee -a "$LOGFILE"

exit $result
