#!/bin/sh

# Test script for chkfont-go compatibility with chkfont (C version)

LC_ALL=POSIX
export LC_ALL

CHKFONT_GO=./chkfont-go
CHKFONT_C=chkfont
OUTPUT_GO=$(mktemp)
OUTPUT_C=$(mktemp)
FAILED=0
PASSED=0

cleanup() {
    rm -f "$OUTPUT_GO" "$OUTPUT_C"
}

trap cleanup EXIT

run_test() {
    test_name="$1"
    shift
    args="$@"
    
    printf "Test: %s... " "$test_name"
    
    $CHKFONT_GO $args > "$OUTPUT_GO" 2>&1
    $CHKFONT_C $args > "$OUTPUT_C" 2>&1
    
    # Normalize program names in output for comparison
    sed -i 's/chkfont-go:/chkfont:/g' "$OUTPUT_GO"
    
    if diff -q "$OUTPUT_GO" "$OUTPUT_C" > /dev/null 2>&1; then
        echo "pass"
        PASSED=$((PASSED + 1))
    else
        echo "**FAIL**"
        echo "  Go output:"
        head -5 "$OUTPUT_GO" | sed 's/^/    /'
        echo "  C output:"
        head -5 "$OUTPUT_C" | sed 's/^/    /'
        FAILED=$((FAILED + 1))
    fi
}

echo "======================================"
echo "chkfont-go compatibility tests"
echo "======================================"
echo

# Build chkfont-go if needed
if [ ! -f "$CHKFONT_GO" ]; then
    echo "Building chkfont-go..."
    go build -o chkfont-go chkfont.go
    if [ $? -ne 0 ]; then
        echo "Failed to build chkfont-go"
        exit 1
    fi
fi

# Check if chkfont (C) is available
if ! command -v $CHKFONT_C > /dev/null 2>&1; then
    echo "Error: chkfont (C version) not found in PATH"
    exit 1
fi

# Test 1: Standard font (no errors expected)
run_test "standard.flf (no errors)" "fonts/standard.flf"

# Test 2: Big font
run_test "big.flf" "fonts/big.flf"

# Test 3: Small font
run_test "small.flf" "fonts/small.flf"

# Test 4: Slant font  
run_test "slant.flf" "fonts/slant.flf"

# Test 5: Banner font (has code-tagged characters)
run_test "banner.flf (with code tags)" "fonts/banner.flf"

# Test 6: Block font
run_test "block.flf" "fonts/block.flf"

# Test 7: Script font
run_test "script.flf" "fonts/script.flf"

# Test 8: Multiple fonts at once
run_test "multiple fonts" "fonts/big.flf" "fonts/small.flf" "fonts/slant.flf"

# Test 9: TLF file (should report errors)
run_test "emboss.tlf (errors expected)" "tests/emboss.tlf"

# Test 10: All standard fonts
run_test "all .flf fonts" fonts/*.flf

# Test 11: Non-existent file
run_test "non-existent file" "nonexistent.flf"

# Test 12: Wrong extension
echo "test" > /tmp/test_font.txt
run_test "wrong extension file" "/tmp/test_font.txt"
rm -f /tmp/test_font.txt

# Test 13: Ivrit (right-to-left) font
run_test "ivrit.flf (rtl font)" "fonts/ivrit.flf"

# Test 14: Shadow font
run_test "shadow.flf" "fonts/shadow.flf"

# Test 15: Term font (minimal)
run_test "term.flf" "fonts/term.flf"

# Test 16: Digital font
run_test "digital.flf" "fonts/digital.flf"

# Test 17: Mini font
run_test "mini.flf" "fonts/mini.flf"

# Test 18: Bubble font
run_test "bubble.flf" "fonts/bubble.flf"

# Test 19: Lean font
run_test "lean.flf" "fonts/lean.flf"

# Test 20: Mnemonic font
run_test "mnemonic.flf" "fonts/mnemonic.flf"

echo
echo "======================================"
echo "Results: $PASSED passed, $FAILED failed"
echo "SUMMARY: PASSED=$PASSED, FAILED=$FAILED, TOTAL=$((PASSED + FAILED))"
echo "======================================"

if [ $FAILED -gt 0 ]; then
    exit 1
fi

exit 0
