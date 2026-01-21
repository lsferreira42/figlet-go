#!/bin/sh

# Script to test output parsers (terminal, terminal-color, html)
# Tests all three output formats

LC_ALL=POSIX
export LC_ALL

TESTDIR=tests
OUTPUT=`mktemp`
LOGFILE=output-tests.log
CMD=${FIGLET_BIN:-./figlet-bin}

run_test() {
	test_num=$1
	test_dsc=$2
	test_cmd=$3
	expected_file=$4

	echo >> $LOGFILE
	printf "Run test $test_num: ${test_dsc}... " | tee -a $LOGFILE
	echo >> $LOGFILE
	echo "Command: $test_cmd" >> $LOGFILE
	# Use sh -c to properly handle commands with semicolons
	sh -c "$test_cmd" > "$OUTPUT" 2>> $LOGFILE
	cmd_exit=$?
	
	if [ -n "$expected_file" ] && [ -f "$expected_file" ]; then
		cmp "$OUTPUT" "$expected_file" >> $LOGFILE 2>&1
		if [ $? -eq 0 ]; then
			echo "pass" | tee -a $LOGFILE
		else
			echo "**fail**" | tee -a $LOGFILE
			result=1
			fail=`expr $fail + 1`
		fi
	else
		# For tests without expected output, check if output contains expected patterns
		if [ $cmd_exit -eq 0 ]; then
			# Check for parser-specific patterns
			case "$test_num" in
				001|002|003|004|005)
					# Terminal parser - should not have HTML tags or ANSI codes
					if grep -q "<code>" "$OUTPUT" 2>/dev/null || grep -q $'\x1b\[' "$OUTPUT" 2>/dev/null; then
						echo "**fail** (unexpected parser output)" | tee -a $LOGFILE
						result=1
						fail=`expr $fail + 1`
					else
						echo "pass" | tee -a $LOGFILE
					fi
					;;
				006|007|008|009|010)
					# Terminal-color parser - should have ANSI codes (if colors are set)
					# Check if colors were provided in the command
					if echo "$test_cmd" | grep -q "colors"; then
						# Check for ANSI escape sequence (ESC[)
						if od -An -tx1 "$OUTPUT" 2>/dev/null | grep -q "1b 5b" || grep -q $'\033\[' "$OUTPUT" 2>/dev/null; then
							echo "pass" | tee -a $LOGFILE
						else
							echo "**fail** (missing ANSI codes)" | tee -a $LOGFILE
							result=1
							fail=`expr $fail + 1`
						fi
					else
						# No colors, should work like terminal parser
						echo "pass" | tee -a $LOGFILE
					fi
					;;
				011|012|013|014|015)
					# HTML parser - should have HTML tags
					if grep -q "<code>" "$OUTPUT" 2>/dev/null; then
						echo "pass" | tee -a $LOGFILE
					else
						echo "**fail** (missing HTML tags)" | tee -a $LOGFILE
						result=1
						fail=`expr $fail + 1`
					fi
					;;
				*)
					# For other tests, just check if command succeeded and output exists
					if [ $cmd_exit -eq 0 ] && [ -s "$OUTPUT" ]; then
						echo "pass" | tee -a $LOGFILE
					else
						echo "**fail**" | tee -a $LOGFILE
						result=1
						fail=`expr $fail + 1`
					fi
					;;
			esac
		else
			echo "**fail**" | tee -a $LOGFILE
			result=1
			fail=`expr $fail + 1`
		fi
	fi
	total=`expr $total + 1`
}

result=0
fail=0
total=0
$CMD -v > $LOGFILE 2>&1

echo "=========================================" | tee -a $LOGFILE
echo "FIGlet Output Parser Tests" | tee -a $LOGFILE
echo "=========================================" | tee -a $LOGFILE
echo "" | tee -a $LOGFILE

# Test terminal parser (default, no colors)
echo "Testing terminal parser (default)..." | tee -a $LOGFILE
run_test 001 "Terminal parser - default" \
	"echo 'Test' | $CMD" \
	""

run_test 002 "Terminal parser - explicit" \
	"echo 'Test' | $CMD --parser terminal" \
	""

run_test 003 "Terminal parser - with font" \
	"echo 'Test' | $CMD -f banner --parser terminal" \
	""

run_test 004 "Terminal parser - long text" \
	"echo 'Hello World' | $CMD --parser terminal" \
	""

run_test 005 "Terminal parser - multiple lines" \
	"printf 'Line1\nLine2' | $CMD --parser terminal" \
	""

# Test terminal-color parser
echo "" | tee -a $LOGFILE
echo "Testing terminal-color parser..." | tee -a $LOGFILE
run_test 006 "Terminal-color parser - with ANSI colors" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'red;green;blue'" \
	""

run_test 007 "Terminal-color parser - with TrueColor" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'FF0000;00FF00;0000FF'" \
	""

run_test 008 "Terminal-color parser - without colors (should still work)" \
	"echo 'Test' | $CMD --parser terminal-color" \
	""

run_test 009 "Terminal-color parser - with font" \
	"echo 'Test' | $CMD -f slant --parser terminal-color --colors red" \
	""

run_test 010 "Terminal-color parser - long text with colors" \
	"echo 'Hello World' | $CMD --parser terminal-color --colors 'red;green;blue'" \
	""

# Test HTML parser
echo "" | tee -a $LOGFILE
echo "Testing HTML parser..." | tee -a $LOGFILE
run_test 011 "HTML parser - basic" \
	"echo 'Test' | $CMD --parser html" \
	""

run_test 012 "HTML parser - with colors" \
	"echo 'Test' | $CMD --parser html --colors 'red;green;blue'" \
	""

run_test 013 "HTML parser - with TrueColor" \
	"echo 'Test' | $CMD --parser html --colors 'FF0000;00FF00;0000FF'" \
	""

run_test 014 "HTML parser - with font" \
	"echo 'Test' | $CMD -f banner --parser html --colors red" \
	""

run_test 015 "HTML parser - long text" \
	"echo 'Hello World' | $CMD --parser html --colors 'red;green;blue'" \
	""

# Test parser switching
echo "" | tee -a $LOGFILE
echo "Testing parser switching..." | tee -a $LOGFILE
run_test 016 "Parser switching - terminal to html" \
	"echo 'Test' | $CMD --parser terminal && echo 'Test' | $CMD --parser html" \
	""

run_test 017 "Parser switching - html to terminal-color" \
	"echo 'Test' | $CMD --parser html --colors red && echo 'Test' | $CMD --parser terminal-color --colors red" \
	""

# Test invalid parser (should handle gracefully)
echo "" | tee -a $LOGFILE
echo "Testing invalid parser handling..." | tee -a $LOGFILE
run_test 018 "Invalid parser (should handle gracefully)" \
	"echo 'Test' | $CMD --parser invalid 2>&1 || true" \
	""

# Test parser with various options
echo "" | tee -a $LOGFILE
echo "Testing parsers with various options..." | tee -a $LOGFILE
run_test 019 "HTML parser with justification" \
	"echo 'Test' | $CMD -c --parser html --colors red" \
	""

run_test 020 "Terminal-color parser with width" \
	"echo 'Hello World' | $CMD -w 60 --parser terminal-color --colors 'red;green;blue'" \
	""

rm -f "$OUTPUT"

echo "" | tee -a $LOGFILE
echo "=========================================" | tee -a $LOGFILE
if [ $result -ne 0 ]; then
	echo " $fail tests failed. See $LOGFILE for result details" | tee -a $LOGFILE
else
	echo " All output parser tests passed." | tee -a $LOGFILE
fi
passed=`expr $total - $fail`
echo "SUMMARY: PASSED=$passed, FAILED=$fail, TOTAL=$total"
echo "=========================================" | tee -a $LOGFILE

exit $result
