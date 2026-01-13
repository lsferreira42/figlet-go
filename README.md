<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/github/actions/workflow/status/lsferreira42/figlet-go/ci.yml?branch=main&style=for-the-badge&logo=github&label=Build" alt="Build Status">
  <img src="https://img.shields.io/badge/License-BSD--3--Clause-green?style=for-the-badge" alt="License">
  <img src="https://goreportcard.com/badge/github.com/lsferreira42/figlet-go?style=for-the-badge" alt="Go Report Card">
  <img src="https://img.shields.io/badge/FIGlet-2.2.5-orange?style=for-the-badge" alt="FIGlet Version">
</p>

<h1 align="center">
  <br>
  🎨 FIGlet-Go
  <br>
</h1>

<h4 align="center">A complete, 100% compatible rewrite of FIGlet in Go</h4>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#fonts">Fonts</a> •
  <a href="#compatibility">Compatibility</a> •
  <a href="#building">Building</a> •
  <a href="#testing">Testing</a> •
  <a href="#license">License</a>
</p>

---

```
 _____ ___ ____ _      _         ____       
|  ___|_ _/ ___| | ___| |_      / ___| ___  
| |_   | | |  _| |/ _ \ __|____| |  _ / _ \ 
|  _|  | | |_| | |  __/ ||_____| |_| | (_) |
|_|   |___\____|_|\___|\__|     \____|\___/ 
```

**FIGlet-Go** is a complete rewrite of the classic [FIGlet](http://www.figlet.org/) program in Go. It generates text banners in various typefaces composed of ASCII art characters. This implementation is **100% compatible** with the original C version, passing all compatibility tests.

## ✨ Features

- 🚀 **Pure Go implementation** - Single binary, no dependencies
- 📦 **Embedded fonts** - All standard FIGlet fonts bundled in the binary
- 🔄 **100% Compatible** - Passes all FIGlet 2.2.5 compatibility tests
- 🎨 **20+ Built-in fonts** - Including standard, big, small, slant, banner, and more
- 📝 **Control files support** - Full support for `.flc` control files
- 🌐 **Unicode support** - UTF-8, ISO 2022, DBCS, HZ, and Shift-JIS encodings
- ↔️ **Right-to-left text** - Support for RTL languages (Hebrew, etc.)
- 🔧 **TOIlet compatibility** - Support for TOIlet font format (`.tlf`)

## 📥 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/lsferreira42/figlet-go.git
cd figlet-go

# Build
make build

# Or install directly to /usr/local/bin
sudo make install
```

### Using Go Install

```bash
go install github.com/lsferreira42/figlet-go@latest
```

## 🚀 Usage

### Basic Usage

```bash
# Simple text banner
echo "Hello World" | ./figlet

# Or provide text as argument
./figlet "Hello World"
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
| `-c` | Center justify output |
| `-l` | Left justify output |
| `-r` | Right justify output |
| `-k` | Use kerning (letters touch) |
| `-o` | Use overlapping (letters overlap) |
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
| `-I code` | Display info (0-5) |

### Examples

```bash
# Use a different font
echo "Hello" | ./figlet -f banner

# Centered output with slant font
echo "Centered" | ./figlet -c -f slant

# Right-to-left text
echo "RTL" | ./figlet -R -f ivrit

# Wide output
echo "Wide Text" | ./figlet -w 120

# Full width (no letter overlapping)
echo "FULL" | ./figlet -W

# Show version
./figlet -v
```

### Sample Output

**Standard font:**
```
 _   _      _ _        __        __         _     _ 
| | | | ___| | | ___   \ \      / /__  _ __| | __| |
| |_| |/ _ \ | |/ _ \   \ \ /\ / / _ \| '__| |/ _` |
|  _  |  __/ | | (_) |   \ V  V / (_) | |  | | (_| |
|_| |_|\___|_|_|\___/     \_/\_/ \___/|_|  |_|\__,_|
```

**Banner font:**
```
#     # ####### #       #       ####### 
#     # #       #       #       #     # 
#     # #       #       #       #     # 
####### #####   #       #       #     # 
#     # #       #       #       #     # 
#     # #       #       #       #     # 
#     # ####### ####### ####### ####### 
```

**Slant font:**
```
    __  __     ____         _       __           __    __
   / / / /__  / / /___     | |     / /___  _____/ /___/ /
  / /_/ / _ \/ / / __ \    | | /| / / __ \/ ___/ / __  / 
 / __  /  __/ / / /_/ /    | |/ |/ / /_/ / /  / / /_/ /  
/_/ /_/\___/_/_/\____/     |__/|__/\____/_/  /_/\__,_/   
```

## 🎨 Fonts

### Included Fonts (.flf)

| Font | Description |
|------|-------------|
| `standard` | Default FIGlet font |
| `banner` | Large banner style |
| `big` | Large font |
| `block` | Block letters |
| `bubble` | Bubble letters |
| `digital` | Digital display style |
| `ivrit` | Hebrew (right-to-left) |
| `lean` | Lean letters |
| `mini` | Minimal/compact |
| `mnemonic` | Mnemonic style |
| `script` | Script/cursive style |
| `shadow` | Letters with shadow |
| `slant` | Italic/slanted |
| `small` | Compact font |
| `smscript` | Small script |
| `smshadow` | Small shadow |
| `smslant` | Small slant |
| `term` | Terminal-friendly |

### Control Files (.flc)

Control files provide character mapping and encoding support:

- `646-*` - ISO 646 national variants
- `8859-*` - ISO 8859 character sets
- `utf8` - UTF-8 encoding
- `jis0201` - JIS X 0201 (Japanese katakana)
- `uskata` - US to Katakana mapping
- `koi8r` - KOI8-R (Russian)
- And more...

### Using Custom Fonts

```bash
# Use a font from a specific directory
./figlet -d /path/to/fonts -f myfont "Hello"

# Set default font directory via environment variable
export FIGLET_FONTDIR=/path/to/fonts
./figlet -f myfont "Hello"
```

## 🔄 Compatibility

This implementation is **100% compatible** with the original FIGlet 2.2.5. It:

- ✅ Passes all 26 official test cases
- ✅ Produces identical output to the C version
- ✅ Supports all command-line options
- ✅ Handles all font files (.flf) and control files (.flc)
- ✅ Supports TOIlet fonts (.tlf)
- ✅ Handles all encoding modes (ISO 2022, UTF-8, DBCS, HZ, Shift-JIS)

### Included Utilities

| Tool | Description |
|------|-------------|
| `figlet` | Main text banner generator |
| `chkfont-go` | Font file validator (checks .flf files for errors) |
| `figlist` | Lists available fonts and control files |
| `showfigfonts` | Shows samples of all available fonts |

## 🔨 Building

### Prerequisites

- Go 1.21 or later

### Build Commands

```bash
# Build everything
make all

# Build only figlet
make build

# Build only chkfont
make build-chkfont

# Clean build artifacts
make clean

# Show all available targets
make help
```

## 🧪 Testing

### Run All Tests

```bash
# Run the main test suite
make test

# Run chkfont tests
make test-chkfont

# Run all tests
make test-all
```

### Compatibility Testing

To run compatibility tests against the original C version (requires `figlet` in PATH):

```bash
make test-compat
```

### Test Coverage

The test suite includes:

- ✅ Text rendering in all fonts
- ✅ All justification modes (left, center, right)
- ✅ All smushing modes (kerning, overlap, full width)
- ✅ Right-to-left text rendering
- ✅ Long text wrapping
- ✅ Paragraph mode
- ✅ Control file processing
- ✅ TOIlet font support
- ✅ Various output widths

## 📁 Project Structure

```
figlet-go/
├── figlet.go          # Main FIGlet implementation
├── chkfont.go         # Font checker implementation
├── Makefile           # Build system
├── fonts/             # Font files (.flf) and control files (.flc)
│   ├── standard.flf
│   ├── banner.flf
│   ├── ...
│   └── utf8.flc
├── tests/             # Test files and expected results
│   ├── input.txt
│   ├── res001.txt
│   └── ...
├── run-tests.sh       # Main test runner
├── run-chkfont-tests.sh
├── test-compatibility.sh
├── showfigfonts       # Font showcase script
├── figlist            # Font listing script
└── LICENSE
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.

```
Copyright (C) 1991, 1993, 1994 Glenn Chappell and Ian Chai
Copyright (C) 1996-2001 John Cowan
Copyright (C) 2002 Christiaan Keet
Copyright (C) 2011 Claudio Matsuoka
All rights reserved.
```

## 🙏 Acknowledgments

- Original [FIGlet](http://www.figlet.org/) authors
- Glenn Chappell, Ian Chai, John Cowan, Christiaan Keet, and Claudio Matsuoka
- The FIGlet font designers community

## 📚 References

- [FIGlet Official Website](http://www.figlet.org/)
- [FIGlet Font Database](http://www.figlet.org/fontdb.cgi)
- [FIGfont Documentation](http://www.jave.de/figlet/figfont.html)

---

<p align="center">
  Made with ❤️ in Go
</p>
