#!/bin/bash

# Master script to run all test suites and aggregate results

TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_TESTS=0

run_suite() {
    local name=$1
    local cmd=$2
    
    echo "================================================================================"
    echo " RUNNING SUITE: $name"
    echo "================================================================================"
    
    # Run command and capture output
    # We use temporary file to avoid issues with pipes and exit codes
    local tmpfile=$(mktemp)
    eval "$cmd" 2>&1 | tee "$tmpfile"
    local exit_code=${PIPESTATUS[0]}
    
    # Extract summary
    local summary=$(grep "SUMMARY:" "$tmpfile" | tail -1)
    if [[ $summary =~ PASSED=([0-9]+),\ FAILED=([0-9]+),\ TOTAL=([0-9]+) ]]; then
        local p=${BASH_REMATCH[1]}
        local f=${BASH_REMATCH[2]}
        local t=${BASH_REMATCH[3]}
        
        TOTAL_PASSED=$((TOTAL_PASSED + p))
        TOTAL_FAILED=$((TOTAL_FAILED + f))
        TOTAL_TESTS=$((TOTAL_TESTS + t))
        
        echo "--------------------------------------------------------------------------------"
        echo " Suite $name: $p OK, $f FAILED, $t TOTAL"
    else
        echo "--------------------------------------------------------------------------------"
        echo " Suite $name: Failed to parse summary!"
        if [ $exit_code -ne 0 ]; then
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
        else
            TOTAL_PASSED=$((TOTAL_PASSED + 1))
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
        fi
    fi
    echo "================================================================================"
    echo ""
    
    rm -f "$tmpfile"
}

# Ensure binaries are built
make build build-chkfont build-wasm || exit 1

run_suite "Functional CLI" "./run-tests.sh fonts"
run_suite "Library" "./run-lib-tests.sh"
run_suite "Chkfont Compatibility" "./run-chkfont-tests.sh"
run_suite "Colors" "./run-colors-tests.sh"
run_suite "Output Parsers" "./run-output-tests.sh"
run_suite "NPM Package" "./run-npm-tests.sh"

# Optional: Run Compatibility Tests if C figlet is installed
if command -v figlet >/dev/null 2>&1; then
    run_suite "C Compatibility" "./run-compatibility-tests.sh TEST"
else
    echo "================================================================================"
    echo " SKIPPING COMPATIBILITY SUITE: 'figlet' (C version) not found"
    echo "================================================================================"
    echo ""
fi

echo "################################################################################"
echo " FINAL TEST REPORT"
echo "################################################################################"
echo " Total Tests:  $TOTAL_TESTS"
echo " Passed:       $TOTAL_PASSED ✅"
if [ $TOTAL_FAILED -gt 0 ]; then EMOJI="❌"; else EMOJI="✅"; fi
echo " Failed:       $TOTAL_FAILED $EMOJI"
echo "################################################################################"

if [ $TOTAL_FAILED -gt 0 ]; then
    exit 1
fi
exit 0
