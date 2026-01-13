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
if [ $result -eq 0 ]; then
    echo "✓ All library tests passed!" | tee -a "$LOGFILE"
else
    echo "✗ Some tests failed. See $LOGFILE for details." | tee -a "$LOGFILE"
fi
echo "=========================================" | tee -a "$LOGFILE"

exit $result
