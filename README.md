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
- Can eventually be used as a Go library

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
echo "Hello World" | ./figlet

# pass as argument
./figlet "Hello World"

# use a different font
./figlet -f banner "Hello"

# centered with slant font
./figlet -c -f slant "Centered"

# right-to-left (Hebrew font)
./figlet -R -f ivrit "Hello"

# custom width
./figlet -w 120 "Wide output"

# full width (no smushing)
./figlet -W "FULL"
```

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

### chkfont

Font file validator. Checks `.flf` files for format errors without modifying them.

```bash
# build chkfont
make build-chkfont

# check a single font
./chkfont-go fonts/standard.flf

# check multiple fonts
./chkfont-go fonts/*.flf
```

Output example:
```
fonts/standard.flf: Errors: 0, Warnings: 0
fonts/standard.flf: maxlen: 22, actual max line length: 22
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

## Fonts

18 fonts are embedded in the binary: `standard`, `banner`, `big`, `block`, `bubble`, `digital`, `ivrit`, `lean`, `mini`, `mnemonic`, `script`, `shadow`, `slant`, `small`, `smscript`, `smshadow`, `smslant`, `term`.

There are also control files (`.flc`) for different encodings: UTF-8, ISO 646 variants, ISO 8859, JIS, KOI8-R, etc.

You can use fonts from other directories:

```bash
./figlet -d /path/to/fonts -f myfont "Hello"
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
├── figlet.go              # main FIGlet implementation
├── terminal_unix.go       # terminal width detection (Linux/macOS)
├── terminal_windows.go    # terminal width detection (Windows)
├── chkfont.go             # font file validator
├── go.mod                 # Go module
├── Makefile               # build commands
├── LICENSE                # BSD 3-Clause
│
├── figlet.6               # man page for figlet
├── chkfont.6              # man page for chkfont
├── showfigfonts.6         # man page for showfigfonts
│
├── figlist                # lists available fonts (shell script)
├── showfigfonts           # shows samples of all fonts (shell script)
├── run-tests.sh           # main test runner
├── run-chkfont-tests.sh   # chkfont test runner
├── test-compatibility.sh  # tests against C version
├── fonts/                 # 18 .flf fonts + strconv .flc control files
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

Next:
- [ ] Refactor as a Go library for use in other projects
- [ ] WASM build for browser usage
- [ ] JavaScript/npm package
- [ ] Color support (ANSI and TrueColor)
- [ ] Output parsers (terminal with colors, HTML)

The color and parser ideas come from [figlet4go](https://github.com/mbndr/figlet4go).

## License

BSD 3-Clause. See [LICENSE](LICENSE).

Original FIGlet by Glenn Chappell, Ian Chai, John Cowan, Christiaan Keet and Claudio Matsuoka.

## Links

- [FIGlet Official](http://www.figlet.org/)
- [FIGlet Font Database](http://www.figlet.org/fontdb.cgi)
