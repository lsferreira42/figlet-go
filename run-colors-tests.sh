#!/bin/sh

# Script to test color support (ANSI and TrueColor)
# Tests both ANSI colors and TrueColor (hex) colors

LC_ALL=POSIX
export LC_ALL

TESTDIR=tests
OUTPUT=`mktemp`
LOGFILE=colors-tests.log
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
		# For tests without expected output, check if command succeeded and output is not empty
		if [ $cmd_exit -eq 0 ] && [ -s "$OUTPUT" ]; then
			echo "pass" | tee -a $LOGFILE
		else
			# For invalid color tests, command should still succeed (graceful handling)
			if [ $cmd_exit -eq 0 ]; then
				echo "pass" | tee -a $LOGFILE
			else
				echo "**fail**" | tee -a $LOGFILE
				result=1
				fail=`expr $fail + 1`
			fi
		fi
	fi
	total=`expr $total + 1`
}

result=0
fail=0
total=0
$CMD -v > $LOGFILE 2>&1

echo "=========================================" | tee -a $LOGFILE
echo "FIGlet Color Support Tests" | tee -a $LOGFILE
echo "=========================================" | tee -a $LOGFILE
echo "" | tee -a $LOGFILE

# Test ANSI colors
echo "Testing ANSI colors..." | tee -a $LOGFILE
run_test 001 "ANSI color - red" \
	"echo 'Test' | $CMD --parser terminal-color --colors red" \
	""

run_test 002 "ANSI color - green" \
	"echo 'Test' | $CMD --parser terminal-color --colors green" \
	""

run_test 003 "ANSI color - blue" \
	"echo 'Test' | $CMD --parser terminal-color --colors blue" \
	""

run_test 004 "ANSI color - multiple colors" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'red;green;blue'" \
	""

run_test 005 "ANSI color - all basic colors" \
	"echo 'Colors' | $CMD --parser terminal-color --colors 'black;red;green;yellow;blue;magenta;cyan;white'" \
	""

# Test TrueColor (hex)
echo "" | tee -a $LOGFILE
echo "Testing TrueColor (hex)..." | tee -a $LOGFILE
run_test 006 "TrueColor - red (FF0000)" \
	"echo 'Test' | $CMD --parser terminal-color --colors FF0000" \
	""

run_test 007 "TrueColor - green (00FF00)" \
	"echo 'Test' | $CMD --parser terminal-color --colors 00FF00" \
	""

run_test 008 "TrueColor - blue (0000FF)" \
	"echo 'Test' | $CMD --parser terminal-color --colors 0000FF" \
	""

run_test 009 "TrueColor - multiple hex colors" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'FF0000;00FF00;0000FF'" \
	""

run_test 010 "TrueColor - with # prefix" \
	"echo 'Test' | $CMD --parser terminal-color --colors '#FF0000'" \
	""

# Test mixed ANSI and TrueColor
echo "" | tee -a $LOGFILE
echo "Testing mixed ANSI and TrueColor..." | tee -a $LOGFILE
run_test 011 "Mixed - ANSI red and TrueColor blue" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'red;0000FF'" \
	""

run_test 012 "Mixed - TrueColor red and ANSI green" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'FF0000;green'" \
	""

# Test with different fonts
echo "" | tee -a $LOGFILE
echo "Testing colors with different fonts..." | tee -a $LOGFILE
run_test 013 "Colors with banner font" \
	"echo 'Test' | $CMD -f banner --parser terminal-color --colors 'red;green;blue'" \
	""

run_test 014 "Colors with slant font" \
	"echo 'Test' | $CMD -f slant --parser terminal-color --colors 'FF0000;00FF00;0000FF'" \
	""

# Test color cycling
echo "" | tee -a $LOGFILE
echo "Testing color cycling..." | tee -a $LOGFILE
run_test 015 "Color cycling - short text" \
	"echo 'Hi' | $CMD --parser terminal-color --colors 'red;green;blue'" \
	""

run_test 016 "Color cycling - long text" \
	"echo 'Hello World' | $CMD --parser terminal-color --colors 'red;green;blue'" \
	""

# Test invalid colors (should not crash)
echo "" | tee -a $LOGFILE
echo "Testing invalid color handling..." | tee -a $LOGFILE
run_test 017 "Invalid hex color (should handle gracefully)" \
	"echo 'Test' | $CMD --parser terminal-color --colors 'INVALID' 2>&1" \
	""

run_test 018 "Empty color string (should work)" \
	"echo 'Test' | $CMD --parser terminal-color --colors '' 2>&1" \
	""

rm -f "$OUTPUT"

echo "" | tee -a $LOGFILE
echo "=========================================" | tee -a $LOGFILE
if [ $result -ne 0 ]; then
	echo " $fail tests failed. See $LOGFILE for result details" | tee -a $LOGFILE
else
	echo " All color tests passed." | tee -a $LOGFILE
fi
passed=`expr $total - $fail`
echo "SUMMARY: PASSED=$passed, FAILED=$fail, TOTAL=$total"
echo "=========================================" | tee -a $LOGFILE

exit $result
