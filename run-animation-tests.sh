#!/bin/bash

# Build the project
go build -o figlet-go figlet.go

# Function to run a test
run_test() {
    local name=$1
    local cmd=$2
    echo "Running test: $name..."
    eval $cmd > /dev/null
    if [ $? -eq 0 ]; then
        echo "✅ $name passed"
    else
        echo "❌ $name failed"
        exit 1
    fi
}

# Test each animation type
run_test "Reveal animation" "./figlet-go --animation reveal 'Test' --animation-delay 1"
run_test "Scroll animation" "./figlet-go --animation scroll 'Test' --animation-delay 1"
run_test "Rain animation" "./figlet-go --animation rain 'Test' --animation-delay 1"
run_test "Wave animation" "./figlet-go --animation wave 'Test' --animation-delay 1"
run_test "Explosion animation" "./figlet-go --animation explosion 'Test' --animation-delay 1"

# Test export and file playback
run_test "Export reveal animation" "./figlet-go --animation reveal 'Export' --export test.ani --animation-delay 1"
run_test "Play animation from file" "./figlet-go --animation-file test.ani"

# Cleanup
rm figlet-go test.ani

echo "All animation tests passed!"
