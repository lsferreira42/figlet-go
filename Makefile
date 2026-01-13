# FIGlet - Go Implementation
# Makefile for building, running, and testing the Go version

BINARY := figlet
CHKFONT := chkfont-go
GOSRC := figlet.go
CHKFONT_SRC := chkfont.go
FONTDIR := fonts
GO := go

.PHONY: all build build-chkfont clean test test-chkfont run install help

# Default target
all: build build-chkfont

# Build the figlet binary
build:
	@echo "Building figlet..."
	$(GO) build -o $(BINARY) $(GOSRC)
	@echo "Build complete: $(BINARY)"

# Build the chkfont binary
build-chkfont:
	@echo "Building chkfont..."
	$(GO) build -o $(CHKFONT) $(CHKFONT_SRC)
	@echo "Build complete: $(CHKFONT)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) figlet-go $(CHKFONT)
	rm -f tests.log compatibility-test.log
	@echo "Clean complete."

# Run the test suite
test: build
	@echo "Running test suite..."
	./run-tests.sh $(FONTDIR)

# Run chkfont test suite
test-chkfont: build-chkfont
	@echo "Running chkfont test suite..."
	./run-chkfont-tests.sh

# Run compatibility tests against C version (requires figlet in PATH)
test-compat: build
	@echo "Running compatibility tests..."
	./test-compatibility.sh TEST

# Run all tests
test-all: test test-chkfont
	@echo "All tests complete."

# Run figlet with example text
run: build
	@echo "Hello World" | ./$(BINARY)

# Run figlet with custom text
# Usage: make run-text TEXT="your message"
run-text: build
	@echo "$(TEXT)" | ./$(BINARY)

# Install to /usr/local/bin (requires sudo)
install: build
	@echo "Installing figlet to /usr/local/bin..."
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Install complete."

# Show help
help:
	@echo "FIGlet Go - Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all           - Build figlet and chkfont (default)"
	@echo "  build         - Build the figlet binary"
	@echo "  build-chkfont - Build the chkfont binary"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run the figlet test suite"
	@echo "  test-chkfont  - Run the chkfont test suite"
	@echo "  test-all      - Run all test suites"
	@echo "  test-compat   - Run compatibility tests (requires C figlet in PATH)"
	@echo "  run           - Build and run with 'Hello World'"
	@echo "  run-text      - Run with custom text (TEXT=\"message\")"
	@echo "  install       - Install to /usr/local/bin (requires sudo)"
	@echo "  help          - Show this help message"
