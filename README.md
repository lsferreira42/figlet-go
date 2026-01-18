<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/github/actions/workflow/status/lsferreira42/figlet-go/ci.yml?branch=main&style=for-the-badge&logo=github&label=Build" alt="Build Status">
  <img src="https://img.shields.io/badge/License-BSD--3--Clause-green?style=for-the-badge" alt="License">
  <img src="https://goreportcard.com/badge/github.com/lsferreira42/figlet-go?style=for-the-badge" alt="Go Report Card">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Linux-FCC624?style=flat-square&logo=linux&logoColor=black" alt="Linux">
  <img src="https://img.shields.io/badge/macOS-000000?style=flat-square&logo=apple&logoColor=white" alt="macOS">
  <img src="https://img.shields.io/badge/Windows-0078D6?style=flat-square&logo=windows&logoColor=white" alt="Windows">
</p>

<h1 align="center">FIGlet-Go</h1>

<p align="center">A complete rewrite of <a href="http://www.figlet.org/">FIGlet</a> in Go</p>

```
 _____ ___ ____ _      _         ____       
|  ___|_ _/ ___| | ___| |_      / ___| ___  
| |_   | | |  _| |/ _ \ __|____| |  _ / _ \ 
|  _|  | | |_| | |  __/ ||_____| |_| | (_) |
|_|   |___\____|_|\___|\__|     \____|\___/ 
```

This is a **100% compatible** implementation - it passes all the original FIGlet 2.2.5 tests and produces identical output to the C version.

## Why?

I wanted a FIGlet that:
- Compiles to a single binary with all fonts embedded
- Works the same way on any platform
- Can be used as a Go library in other projects ✓

## Installation

```bash
git clone https://github.com/lsferreira42/figlet-go.git
cd figlet-go
make build
```

Or with go install:

```bash
go install github.com/lsferreira42/figlet-go@latest
```

## Usage

### figlet

```bash
# pipe text
echo "Hello World" | ./figlet-bin

# pass as argument
./figlet-bin "Hello World"

# use a different font
./figlet-bin -f banner "Hello"

# centered with slant font
./figlet-bin -c -f slant "Centered"

# right-to-left (Hebrew font)
./figlet-bin -R -f ivrit "Hello"

# custom width
./figlet-bin -w 120 "Wide output"

# full width (no smushing)
./figlet-bin -W "FULL"

# with colors (ANSI)
./figlet-bin --colors 'red;green;blue' "Colors"

# with TrueColor (hex)
./figlet-bin --colors 'FF0000;00FF00;0000FF' "TrueColor"

# HTML output
./figlet-bin --parser html "HTML Output"

# colored HTML output
./figlet-bin --parser html --colors 'red;green;blue' "Colored HTML"
```

📖 **[Complete Colors and Output Formats Guide →](colors_outputs.md)**

Sample output with different fonts:

**standard (default):**
```
 _   _      _ _        __        __         _     _ 
| | | | ___| | | ___   \ \      / /__  _ __| | __| |
| |_| |/ _ \ | |/ _ \   \ \ /\ / / _ \| '__| |/ _` |
|  _  |  __/ | | (_) |   \ V  V / (_) | |  | | (_| |
|_| |_|\___|_|_|\___/     \_/\_/ \___/|_|  |_|\__,_|
```

**banner:**
```
#     # ####### #       #       ####### 
#     # #       #       #       #     # 
#     # #       #       #       #     # 
####### #####   #       #       #     # 
#     # #       #       #       #     # 
#     # #       #       #       #     # 
#     # ####### ####### ####### ####### 
```

### Command Line Options

```
Usage: figlet [ -cklnoprstvxDELNRSWX ] [ -d fontdirectory ]
              [ -f fontfile ] [ -m smushmode ] [ -w outputwidth ]
              [ -C controlfile ] [ -I infocode ] [ message ]
```

| Option | Description |
|--------|-------------|
| `-f font` | Specify font file |
| `-d dir` | Specify font directory |
| `-w width` | Set output width (default: 80) |
| `-c` | Center justify |
| `-l` | Left justify |
| `-r` | Right justify |
| `-k` | Kerning mode (letters touch) |
| `-o` | Overlap mode (letters overlap) |
| `-W` | Full width (no smushing) |
| `-S` | Force smushing |
| `-s` | Use font's default smushing |
| `-L` | Left-to-right text |
| `-R` | Right-to-left text |
| `-X` | Auto direction based on font |
| `-p` | Paragraph mode |
| `-n` | Normal mode (default) |
| `-D` | German character translation |
| `-E` | Disable German translation |
| `-C file` | Add control file |
| `-N` | Clear control file list |
| `-t` | Use terminal width |
| `-v` | Display version info |
| `-I code` | Display info (0=version, 1=version int, 2=font dir, 3=font name, 4=output width, 5=supported font formats) |
| `--colors colors` | Set colors for output (e.g., `--colors red;green;blue` or `--colors FF0000;00FF00`) - See [Colors Guide](colors_outputs.md) |
| `--parser parser` | Set output parser (`terminal`, `terminal-color`, or `html`) - See [Output Formats Guide](colors_outputs.md) |

### chkfont

Font file validator. Checks FIGlet 2.0/2.1 font files (`.flf`) for format errors without modifying them.

```bash
# build chkfont
make build-chkfont

# check a single font
./chkfont-go fonts/standard.flf

# check multiple fonts
./chkfont-go fonts/*.flf

# check from stdin
./chkfont-go -
```

```
Usage: chkfont-go fontfile ...
```

**What it checks:**

Errors (fatal):
- Invalid magic number (must be `flf2`)
- First line improperly formatted
- charheight/maxlen not positive
- Unexpected end of file

Errors:
- Filename doesn't end with `.flf`
- Line length exceeds maxlen
- Inconsistent character width within a character
- Too many endmarks (more than 2)
- Invalid layout values
- Invalid old_layout values
- up_height out of bounds
- Code tag -1 (unusable)
- Inconsistent Codetag_Cnt

Warnings:
- Sub-version character is not 'a'
- Unusual hardblank character
- Blank endmark
- Inconsistent endmark between lines
- Endchar count convention violated
- Code tag > 65535
- Code tag in ASCII range (32-126)
- Code tag in old Deutsch area (-255 to -249)
- Non-increasing code tags
- Extra characters after font data

**Output example (valid font):**
```
fonts/standard.flf: Errors: 0, Warnings: 0
-------------------------------------------------------------------------------
```

**Output example (font with issues):**
```
tests/emboss.tlf: ERROR- Filename does not end with '.flf'.
tests/emboss.tlf: ERROR- Incorrect magic number.
tests/emboss.tlf: ERROR- Inconsistent character width in line 27.
tests/emboss.tlf: ERROR- Line length > maxlen in line 38.
*******************************************************************************
tests/emboss.tlf: Too many errors/warnings.
tests/emboss.tlf: Errors: 21, Warnings: 0
tests/emboss.tlf: maxlen: 8, actual max line length: 13
-------------------------------------------------------------------------------
```

### Helper Scripts

```bash
# list all available fonts and control files
./figlist

# show a sample of each font
./showfigfonts

# show a specific word in all fonts
./showfigfonts "Test"

# use fonts from a different directory
./showfigfonts -d /path/to/fonts
```

## Using as a Library

FIGlet-Go can be used as a library in your Go projects. See the **[complete library documentation](lib.md)** for a full tutorial and API reference.

### Quick Start

```bash
go get github.com/lsferreira42/figlet-go/figlet
```

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    // Simple usage
    result, err := figlet.Render("Hello!")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(result)

    // With a specific font
    result, err = figlet.RenderWithFont("Go!", "slant")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(result)

    // With options
    result, err = figlet.Render("Centered",
        figlet.WithFont("big"),
        figlet.WithWidth(60),
        figlet.WithJustification(1), // center
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(result)
}
```

### Library Features

- **Simple API**: `Render()` and `RenderWithFont()` for quick usage
- **Functional Options**: Configure with `WithFont()`, `WithWidth()`, `WithJustification()`, etc.
- **Full Control**: Use `Config` struct directly for advanced usage
- **146 Embedded Fonts**: All fonts from [figlet.org](http://www.figlet.org/fontdb.cgi) are embedded
- **Font Discovery**: `ListFonts()` returns all available fonts
- **Color Support**: ANSI colors and TrueColor (24-bit RGB) support
- **Output Parsers**: Terminal (normal), terminal with colors, and HTML output formats

📖 **[Colors and Output Formats Documentation →](colors_outputs.md)** - Complete guide for command-line usage
- **Color Support**: ANSI colors and TrueColor (24-bit RGB) support
- **Output Parsers**: Terminal (normal), terminal with colors, and HTML output formats

📖 **[Full Library Documentation →](lib.md)**

## Web/Browser Usage

FIGlet-Go can run in the browser via WebAssembly!

### Online Playground

Try it now at: **[FIGlet-Go Playground](https://lsferreira42.github.io/figlet-go/)**

### npm Package

Install the npm package for Node.js or browser use:

```bash
npm install figlet-go
```

```javascript
const figlet = require('figlet-go');

// Simple rendering
const art = await figlet.render('Hello!');
console.log(art);

// With a specific font
const slantArt = await figlet.renderWithFont('Go!', 'slant');
console.log(slantArt);

// List available fonts
const fonts = await figlet.listFonts();
console.log(fonts);
```

### Building WebAssembly

```bash
# Build the WASM module
make build-wasm

# Build and serve the playground locally
make serve-website
# Opens at http://localhost:8080

# Build the npm package
make npm-build

# Publish to npm (requires npm login)
make npm-publish
```

## Compatibility

This implementation is **100% compatible** with the original FIGlet 2.2.5:

- Passes all 26 official test cases
- Produces identical output to the C version
- Supports all command-line options
- Handles all font files (.flf) and control files (.flc)
- Supports TOIlet fonts (.tlf)
- Handles all encoding modes:
  - ISO 2022 (with G0/G1/G2/G3 character sets)
  - UTF-8
  - DBCS (Double-Byte Character Sets)
  - HZ encoding
  - Shift-JIS

You can run compatibility tests against the original C version:

```bash
# requires figlet (C version) installed
make test-compat
```

## Fonts

**146 fonts** are embedded in the binary, downloaded from the [FIGlet font database](http://www.figlet.org/fontdb.cgi). Popular fonts include:

`standard`, `banner`, `big`, `block`, `slant`, `shadow`, `script`, `small`, `doom`, `graffiti`, `starwars`, `larry3d`, `colossal`, `gothic`, `epic`, `poison`, `roman`, `rounded`, `speed`, `stellar`, and many more!

Run `figlist` to see all available fonts, or use `figlet.ListFonts()` in Go.

There are also control files (`.flc`) for different encodings: UTF-8, ISO 646 variants, ISO 8859, JIS, KOI8-R, etc.

You can use fonts from other directories:

```bash
./figlet-bin -d /path/to/fonts -f myfont "Hello"
# or
export FIGLET_FONTDIR=/path/to/fonts
```

## Building

```bash
make build          # build figlet
make build-chkfont  # build the font checker
make test           # run tests
make test-compat    # test against C version (needs figlet installed)
```

Requires Go 1.21+.

## Project Structure

```
figlet-go/
├── figlet.go              # main executable entry point
├── chkfont.go             # font file validator
├── go.mod                 # Go module
├── Makefile               # build commands
├── LICENSE                # BSD 3-Clause
├── lib.md                 # library documentation
│
├── figlet/                # FIGlet library package
│   ├── figlet.go          # core FIGlet implementation
│   ├── figlet_test.go     # library tests
│   ├── terminal_unix.go   # terminal width detection (Linux/macOS)
│   ├── terminal_windows.go # terminal width detection (Windows)
│   └── fonts/             # 146 embedded .flf fonts + .flc control files
│
├── wasm/                  # WebAssembly build source
│   └── main.go            # WASM entry point
│
├── website/               # Online playground
│   ├── index.html         # playground UI
│   ├── styles.css         # styles
│   ├── main.js            # JavaScript
│   ├── wasm_exec.js       # Go WASM support
│   └── figlet.wasm        # compiled WASM (generated)
│
├── npm/                   # npm package
│   ├── package.json       # npm configuration
│   ├── src/               # package source
│   └── README.md          # npm documentation
│
├── example/               # library usage examples
│   └── main.go
│
├── figlet.6               # man page for figlet
├── chkfont.6              # man page for chkfont
├── showfigfonts.6         # man page for showfigfonts
│
├── figlist                # lists available fonts (shell script)
├── showfigfonts           # shows samples of all fonts (shell script)
├── run-tests.sh           # main test runner
├── run-lib-tests.sh       # library test runner
├── run-chkfont-tests.sh   # chkfont test runner
├── run-compatibility-tests.sh  # tests against C version
├── fonts/                 # fonts for CLI (also embedded in library)
└── tests/                 # 26 test cases + input files
```

## Roadmap

Done:
- [x] Full FIGlet 2.2.5 compatibility
- [x] Cross-platform (Linux, macOS, Windows)
- [x] Embedded fonts
- [x] All encoding modes (UTF-8, ISO 2022, DBCS, HZ, Shift-JIS)
- [x] TOIlet font support (.tlf)
- [x] CI/CD
- [x] **Go library for use in other projects** ([documentation](lib.md))
- [x] **WASM build for browser usage** ([playground](https://lsferreira42.github.io/figlet-go/))
- [x] **JavaScript/npm package** ([npm](https://www.npmjs.com/package/figlet-go))

Next:
- [x] **Color support (ANSI and TrueColor)** ✓ - See [Colors and Output Formats Guide](colors_outputs.md)
- [x] **Output parsers (terminal with colors, HTML)** ✓ - See [Colors and Output Formats Guide](colors_outputs.md)

The color and parser ideas come from [figlet4go](https://github.com/mbndr/figlet4go).

## License

BSD 3-Clause. See [LICENSE](LICENSE).

Original FIGlet by Glenn Chappell, Ian Chai, John Cowan, Christiaan Keet and Claudio Matsuoka.

## Links

- [Library Documentation](lib.md) - Complete API reference and tutorial
- [FIGlet Official](http://www.figlet.org/)
- [FIGlet Font Database](http://www.figlet.org/fontdb.cgi)
