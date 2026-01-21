#!/bin/sh

# Wrapper to run NPM tests and output a standardized summary
cd npm
OUTPUT=$(npm test 2>&1)
EXIT_CODE=$?

echo "$OUTPUT"

# Extract counts from Jest output
# Example: Tests:       7 passed, 7 total
PASSED=$(echo "$OUTPUT" | grep "Tests:" | sed -E 's/.* ([0-9]+) passed.*/\1/')
TOTAL=$(echo "$OUTPUT" | grep "Tests:" | sed -E 's/.* ([0-9]+) total.*/\1/')

if [ -z "$PASSED" ]; then PASSED=0; fi
if [ -z "$TOTAL" ]; then TOTAL=0; fi

FAILED=$((TOTAL - PASSED))

echo "SUMMARY: PASSED=$PASSED, FAILED=$FAILED, TOTAL=$TOTAL"

exit $EXIT_CODE
