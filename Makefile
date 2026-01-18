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

# Tool Check Macro
CHECK_TOOL = @command -v $(1) >/dev/null 2>&1 || { echo >&2 "Error: $(1) is not installed. Please install $(1) to continue."; exit 1; }

.PHONY: all build build-chkfont build-wasm clean test test-lib test-chkfont test-colors test-output run install help
.PHONY: website serve-website npm-build npm-publish
.PHONY: packages package-deb package-rpm package-apk package-arch package-appimage

# Default target
all: build build-chkfont

# Build the figlet binary
build:
	$(call CHECK_TOOL,$(GO))
	@echo "Building figlet..."
	$(GO) build -o $(BINARY) $(GOSRC)
	@echo "Build complete: $(BINARY)"

# Build the chkfont binary
build-chkfont:
	$(call CHECK_TOOL,$(GO))
	@echo "Building chkfont..."
	$(GO) build -o $(CHKFONT) $(CHKFONT_SRC)
	@echo "Build complete: $(CHKFONT)"

# Build WebAssembly module
build-wasm:
	$(call CHECK_TOOL,$(GO))
	@echo "Building WebAssembly module..."
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) $(WASM_SRC)
	@echo "Build complete: $(WASM_OUT)"

# Build everything for website
website: build-wasm
	@echo "Website ready in website/ folder"
	@echo "Run 'make serve-website' to start a local server"

# Serve website locally (requires Python 3)
serve-website: build-wasm
	$(call CHECK_TOOL,python3)
	@echo "Starting local server at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	cd website && python3 -m http.server 8080

# Build npm package
# Packaging Targets
packages:
	$(call CHECK_TOOL,goreleaser)
	goreleaser release --snapshot --clean --config .goreleaser.yaml
	chmod +x packages/appimage/build-appimage.sh
	./packages/appimage/build-appimage.sh

package-deb:
	$(call CHECK_TOOL,goreleaser)
	@grep -vE " - rpm| - apk| - archlinux" .goreleaser.yaml > .goreleaser-deb.tmp.yaml
	goreleaser release --snapshot --clean --config .goreleaser-deb.tmp.yaml
	@rm .goreleaser-deb.tmp.yaml

package-rpm:
	$(call CHECK_TOOL,goreleaser)
	@grep -vE " - deb| - apk| - archlinux" .goreleaser.yaml > .goreleaser-rpm.tmp.yaml
	goreleaser release --snapshot --clean --config .goreleaser-rpm.tmp.yaml
	@rm .goreleaser-rpm.tmp.yaml

package-apk:
	$(call CHECK_TOOL,goreleaser)
	@grep -vE " - deb| - rpm| - archlinux" .goreleaser.yaml > .goreleaser-apk.tmp.yaml
	goreleaser release --snapshot --clean --config .goreleaser-apk.tmp.yaml
	@rm .goreleaser-apk.tmp.yaml

package-arch:
	$(call CHECK_TOOL,goreleaser)
	@grep -vE " - deb| - rpm| - apk" .goreleaser.yaml > .goreleaser-arch.tmp.yaml
	goreleaser release --snapshot --clean --config .goreleaser-arch.tmp.yaml
	@rm .goreleaser-arch.tmp.yaml

package-appimage:
	$(call CHECK_TOOL,$(GO))
	$(call CHECK_TOOL,curl)
	chmod +x packages/appimage/build-appimage.sh
	./packages/appimage/build-appimage.sh

package-flatpak:
	$(call CHECK_TOOL,flatpak-builder)
	@echo "Building Flatpak..."
	@# Implementation details would go here if automated, currently manual notice
	@echo "Flatpak build requires flatpak-builder. Run manually from packages/flatpak/"

npm-build: build-wasm
	$(call CHECK_TOOL,$(NPM))
	@echo "Building npm package..."
	cd npm && $(NPM) run build
	@echo "npm package built in npm/dist/"

# Publish to npm (requires npm login)
npm-publish: npm-build
	$(call CHECK_TOOL,$(NPM))
	@echo "Publishing to npm..."
	cd npm && $(NPM) publish
	@echo "Published to npm!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) figlet-go $(CHKFONT)
	rm -f $(WASM_OUT)
	rm -rf npm/dist
	rm -f tests.log compatibility-test.log lib-tests.log colors-tests.log output-tests.log coverage.out
	rm -rf dist/ packages/appimage/work/ packages/appimage/*.AppImage
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
	FIGLET_BIN=./$(BINARY) ./run-compatibility-tests.sh TEST

# Run all tests
test-colors: build
	@echo "Running color tests..."
	./run-colors-tests.sh

test-output: build
	@echo "Running output parser tests..."
	./run-output-tests.sh

test-all: test test-lib test-chkfont test-colors test-output
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
	@echo "General Targets:"
	@echo "  all            - Build figlet and chkfont (default)"
	@echo "  build          - Build the figlet binary"
	@echo "  build-chkfont  - Build the chkfont binary"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install to /usr/local/bin (requires sudo)"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Testing Targets:"
	@echo "  test           - Run the figlet functional test suite"
	@echo "  test-lib       - Run the library test suite"
	@echo "  test-lib-cover - Run library tests with coverage"
	@echo "  test-chkfont   - Run the chkfont test suite"
	@echo "  test-colors    - Run color support tests (ANSI and TrueColor)"
	@echo "  test-output    - Run output parser tests (terminal, html)"
	@echo "  test-all       - Run all test suites"
	@echo "  test-compat    - Run compatibility tests (requires C figlet in PATH)"
	@echo ""
	@echo "Packaging Targets:"
	@echo "  packages       - Build all standard Linux packages"
	@echo "  package-deb    - Build Debian/Ubuntu package"
	@echo "  package-rpm    - Build RedHat/Fedora package"
	@echo "  package-apk    - Build Alpine package"
	@echo "  package-arch   - Build Arch Linux package"
	@echo "  package-appimage - Build AppImage package"
	@echo "  package-flatpak - Show Flatpak instructions"
	@echo ""
	@echo "Web/Other Targets:"
	@echo "  build-wasm     - Build the WASM module"
	@echo "  website        - Build WASM and prepare website"
	@echo "  serve-website  - Start local server at http://localhost:8080"
	@echo "  npm-build      - Build the npm package"
	@echo "  npm-publish    - Publish to npm (requires npm login)"
	@echo "  run            - Build and run with 'Hello World'"
	@echo "  run-text       - Run with custom text (TEXT=\"message\")"
