# FIGlet - Go Implementation
# Makefile for building, running, and testing the Go version

BINARY := figlet-bin
CHKFONT := chkfont-go
GOSRC := figlet.go
CHKFONT_SRC := chkfont.go
WASM_SRC := wasm/main.go
WASM_OUT := website/figlet.wasm
FONTDIR := fonts
GO := go
NPM := npm

.PHONY: all build build-chkfont build-wasm clean test test-lib test-chkfont run install help
.PHONY: website serve-website npm-build npm-publish

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

# Build WebAssembly module
build-wasm:
	@echo "Building WebAssembly module..."
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) $(WASM_SRC)
	@echo "Build complete: $(WASM_OUT)"

# Build everything for website
website: build-wasm
	@echo "Website ready in website/ folder"
	@echo "Run 'make serve-website' to start a local server"

# Serve website locally (requires Python 3)
serve-website: build-wasm
	@echo "Starting local server at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	cd website && python3 -m http.server 8080

# Build npm package
npm-build: build-wasm
	@echo "Building npm package..."
	cd npm && $(NPM) run build
	@echo "npm package built in npm/dist/"

# Publish to npm (requires npm login)
npm-publish: npm-build
	@echo "Publishing to npm..."
	cd npm && $(NPM) publish
	@echo "Published to npm!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) figlet-go $(CHKFONT)
	rm -f $(WASM_OUT)
	rm -rf npm/dist
	rm -f tests.log compatibility-test.log lib-tests.log coverage.out
	@echo "Clean complete."

# Run the test suite
test: build
	@echo "Running test suite..."
	FIGLET_BIN=./$(BINARY) ./run-tests.sh $(FONTDIR)

# Run library tests
test-lib:
	@echo "Running library test suite..."
	./run-lib-tests.sh -v

# Run library tests with coverage
test-lib-cover:
	@echo "Running library test suite with coverage..."
	./run-lib-tests.sh -v -c

# Run chkfont test suite
test-chkfont: build-chkfont
	@echo "Running chkfont test suite..."
	./run-chkfont-tests.sh

# Run compatibility tests against C version (requires figlet in PATH)
test-compat: build
	@echo "Running compatibility tests..."
	FIGLET_BIN=./$(BINARY) ./test-compatibility.sh TEST

# Run all tests
test-all: test test-lib test-chkfont
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
	install -m 755 $(BINARY) /usr/local/bin/figlet
	@echo "Install complete."

# Show help
help:
	@echo "FIGlet Go - Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Build figlet and chkfont (default)"
	@echo "  build          - Build the figlet binary"
	@echo "  build-chkfont  - Build the chkfont binary"
	@echo "  build-wasm     - Build WebAssembly module"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run the figlet test suite"
	@echo "  test-lib       - Run the library test suite"
	@echo "  test-lib-cover - Run library tests with coverage"
	@echo "  test-chkfont   - Run the chkfont test suite"
	@echo "  test-all       - Run all test suites"
	@echo "  test-compat    - Run compatibility tests (requires C figlet in PATH)"
	@echo "  run            - Build and run with 'Hello World'"
	@echo "  run-text       - Run with custom text (TEXT=\"message\")"
	@echo "  install        - Install to /usr/local/bin (requires sudo)"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "WebAssembly/Website:"
	@echo "  build-wasm     - Build the WASM module (website/figlet.wasm)"
	@echo "  website        - Build WASM and prepare website"
	@echo "  serve-website  - Start local server at http://localhost:8080"
	@echo ""
	@echo "npm Package:"
	@echo "  npm-build      - Build the npm package"
	@echo "  npm-publish    - Publish to npm (requires npm login)"
	@echo ""
	@echo "Library Usage:"
	@echo "  The figlet package can be imported as a library:"
	@echo "    import \"github.com/lsferreira42/figlet-go/figlet\""
	@echo ""
	@echo "  Example:"
	@echo "    result, err := figlet.Render(\"Hello\")"
	@echo "    result, err := figlet.RenderWithFont(\"Hello\", \"slant\")"