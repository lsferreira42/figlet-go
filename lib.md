# FIGlet-Go Library Documentation

This document provides a complete tutorial and API reference for using FIGlet-Go as a library in your Go projects.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Tutorial](#tutorial)
  - [Basic Usage](#basic-usage)
  - [Using Different Fonts](#using-different-fonts)
  - [Configuration Options](#configuration-options)
  - [Advanced Usage with Config](#advanced-usage-with-config)
  - [Listing Available Fonts](#listing-available-fonts)
- [API Reference](#api-reference)
  - [Functions](#functions)
  - [Types](#types)
  - [Option Functions](#option-functions)
  - [Constants](#constants)
- [Examples](#examples)
- [Best Practices](#best-practices)

---

## Installation

Add the library to your Go project:

```bash
go get github.com/lsferreira42/figlet-go/figlet
```

Then import it in your code:

```go
import "github.com/lsferreira42/figlet-go/figlet"
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    result, err := figlet.Render("Hello!")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(result)
}
```

Output:
```
 _   _      _ _       _ 
| | | | ___| | | ___ | |
| |_| |/ _ \ | |/ _ \| |
|  _  |  __/ | | (_) |_|
|_| |_|\___|_|_|\___/(_)
```

---

## Tutorial

### Basic Usage

The simplest way to use the library is with the `Render` function:

```go
result, err := figlet.Render("Hello World")
if err != nil {
    log.Fatal(err)
}
fmt.Print(result)
```

### Using Different Fonts

Use `RenderWithFont` to specify a font:

```go
result, err := figlet.RenderWithFont("Go!", "slant")
if err != nil {
    log.Fatal(err)
}
fmt.Print(result)
```

Output:
```
   ______      __
  / ____/___  / /
 / / __/ __ \/ / 
/ /_/ / /_/ /_/  
\____/\____(_)   
```

Or use the `WithFont` option:

```go
result, err := figlet.Render("Go!", figlet.WithFont("slant"))
```

#### Available Fonts

The library includes 18 embedded fonts:

| Font | Description |
|------|-------------|
| `standard` | Default FIGlet font |
| `banner` | Large block letters using `#` |
| `big` | Large rounded letters |
| `block` | Solid block letters |
| `bubble` | Rounded bubble letters |
| `digital` | Digital display style |
| `ivrit` | Hebrew (right-to-left) |
| `lean` | Thin slanted letters |
| `mini` | Very small font |
| `mnemonic` | ASCII representation |
| `script` | Cursive script style |
| `shadow` | Letters with shadow |
| `slant` | Italic/slanted style |
| `small` | Compact font |
| `smscript` | Small script |
| `smshadow` | Small with shadow |
| `smslant` | Small slanted |
| `term` | Terminal-style |

### Configuration Options

Use option functions to customize the output:

```go
result, err := figlet.Render("Centered",
    figlet.WithFont("big"),
    figlet.WithWidth(100),
    figlet.WithJustification(1), // center
)
```

#### Available Options

| Option | Description |
|--------|-------------|
| `WithFont(name)` | Set the font to use |
| `WithFontDir(dir)` | Set custom font directory |
| `WithWidth(width)` | Set output width (default: 80) |
| `WithJustification(j)` | Set justification: -1=auto, 0=left, 1=center, 2=right |
| `WithRightToLeft(r)` | Set direction: -1=auto, 0=left-to-right, 1=right-to-left |
| `WithSmushMode(mode)` | Set smush mode (advanced) |
| `WithKerning()` | Enable kerning (letters touch) |
| `WithFullWidth()` | Disable smushing (full width) |
| `WithSmushing()` | Force smushing |
| `WithOverlapping()` | Enable overlapping mode |

#### Justification Examples

```go
// Left justified (default)
result, _ := figlet.Render("Left", figlet.WithJustification(0))

// Center justified
result, _ := figlet.Render("Center", figlet.WithJustification(1))

// Right justified
result, _ := figlet.Render("Right", figlet.WithJustification(2))
```

#### Smushing Modes

```go
// Full width - no smushing, characters don't touch
result, _ := figlet.Render("FULL", figlet.WithFullWidth())

// Kerning - characters touch but don't overlap
result, _ := figlet.Render("Kern", figlet.WithKerning())

// Smushing - characters overlap (default for most fonts)
result, _ := figlet.Render("Smush", figlet.WithSmushing())
```

### Advanced Usage with Config

For more control, use the `Config` struct directly:

```go
cfg := figlet.New()
cfg.Fontname = "banner"
cfg.Outputwidth = 120
cfg.Justification = 1 // center

if err := cfg.LoadFont(); err != nil {
    log.Fatal(err)
}

result := cfg.RenderString("Hello")
fmt.Print(result)
```

#### Config Fields

| Field | Type | Description |
|-------|------|-------------|
| `Fontname` | `string` | Name of the font to use |
| `Fontdirname` | `string` | Directory to search for fonts |
| `Outputwidth` | `int` | Maximum output width |
| `Justification` | `int` | -1=auto, 0=left, 1=center, 2=right |
| `Right2left` | `int` | -1=auto, 0=LTR, 1=RTL |
| `Smushmode` | `int` | Smushing mode flags |
| `Smushoverride` | `int` | Override font's smush mode |
| `Paragraphflag` | `bool` | Enable paragraph mode |
| `Deutschflag` | `bool` | Enable German character translation |

#### Config Methods

```go
// Create a new config with defaults
cfg := figlet.New()

// Load the font (required before rendering)
err := cfg.LoadFont()

// Render a string
result := cfg.RenderString("Hello")

// Add a control file for character translation
cfg.AddControlFile("utf8")

// Clear all control files
cfg.ClearControlFiles()
```

### Listing Available Fonts

```go
fonts := figlet.ListFonts()
for _, font := range fonts {
    fmt.Println(font)
}
```

### Getting Version Information

```go
fmt.Println("Version:", figlet.GetVersion())      // "2.2.5"
fmt.Println("Version Int:", figlet.GetVersionInt()) // 20205
```

---

## API Reference

### Functions

#### `Render`

```go
func Render(text string, options ...Option) (string, error)
```

Renders text using FIGlet with the specified options. Uses the default font ("standard") if no font is specified.

**Parameters:**
- `text` - The text to render
- `options` - Optional configuration functions

**Returns:**
- The rendered ASCII art as a string
- An error if the font cannot be loaded

**Example:**
```go
result, err := figlet.Render("Hello", figlet.WithFont("slant"))
```

---

#### `RenderWithFont`

```go
func RenderWithFont(text, fontName string) (string, error)
```

Convenience function to render text with a specific font.

**Parameters:**
- `text` - The text to render
- `fontName` - Name of the font to use

**Returns:**
- The rendered ASCII art as a string
- An error if the font cannot be loaded

**Example:**
```go
result, err := figlet.RenderWithFont("Hello", "banner")
```

---

#### `ListFonts`

```go
func ListFonts() []string
```

Returns a list of all available embedded fonts.

**Returns:**
- A slice of font names

**Example:**
```go
fonts := figlet.ListFonts()
// ["banner", "big", "block", ...]
```

---

#### `GetVersion`

```go
func GetVersion() string
```

Returns the FIGlet version string.

**Returns:**
- Version string (e.g., "2.2.5")

---

#### `GetVersionInt`

```go
func GetVersionInt() int
```

Returns the FIGlet version as an integer.

**Returns:**
- Version integer (e.g., 20205)

---

#### `GetColumns`

```go
func GetColumns() int
```

Returns the current terminal width. Returns -1 if it cannot be determined.

**Returns:**
- Terminal width in columns, or -1

---

#### `New`

```go
func New() *Config
```

Creates a new Config with default values.

**Returns:**
- A pointer to a new Config

**Example:**
```go
cfg := figlet.New()
cfg.Fontname = "slant"
```

---

### Types

#### `Config`

The main configuration structure for FIGlet rendering.

```go
type Config struct {
    Deutschflag   bool   // German character translation
    Justification int    // -1=auto, 0=left, 1=center, 2=right
    Paragraphflag bool   // Paragraph mode
    Right2left    int    // -1=auto, 0=LTR, 1=RTL
    Multibyte     int    // Encoding mode
    Cmdinput      bool   // Command input mode
    Smushmode     int    // Smushing mode
    Smushoverride int    // Smush override
    Outputwidth   int    // Output width
    Fontdirname   string // Font directory
    Fontname      string // Font name
    // ... internal fields
}
```

**Methods:**

| Method | Description |
|--------|-------------|
| `LoadFont() error` | Load the specified font |
| `RenderString(text string) string` | Render text to ASCII art |
| `AddControlFile(name string)` | Add a control file |
| `ClearControlFiles()` | Clear all control files |

---

#### `Option`

```go
type Option func(*Config)
```

Function type for configuring a FIGlet instance.

---

### Option Functions

#### `WithFont`

```go
func WithFont(name string) Option
```

Sets the font name.

---

#### `WithFontDir`

```go
func WithFontDir(dir string) Option
```

Sets the font directory for loading fonts from the filesystem.

---

#### `WithWidth`

```go
func WithWidth(width int) Option
```

Sets the output width. Default is 80.

---

#### `WithJustification`

```go
func WithJustification(j int) Option
```

Sets text justification:
- `-1` - Auto (based on font)
- `0` - Left
- `1` - Center
- `2` - Right

---

#### `WithRightToLeft`

```go
func WithRightToLeft(r int) Option
```

Sets text direction:
- `-1` - Auto (based on font)
- `0` - Left-to-right
- `1` - Right-to-left

---

#### `WithSmushMode`

```go
func WithSmushMode(mode int) Option
```

Sets the smushing mode. This is an advanced option.

---

#### `WithKerning`

```go
func WithKerning() Option
```

Enables kerning mode (letters touch but don't overlap).

---

#### `WithFullWidth`

```go
func WithFullWidth() Option
```

Disables all smushing (full width output).

---

#### `WithSmushing`

```go
func WithSmushing() Option
```

Forces smushing mode.

---

#### `WithOverlapping`

```go
func WithOverlapping() Option
```

Enables overlapping mode.

---

### Constants

```go
const (
    VERSION        = "2.2.5"
    VERSION_INT    = 20205
    DEFAULTCOLUMNS = 80
    
    // File suffixes
    FONTFILESUFFIX    = ".flf"
    CONTROLFILESUFFIX = ".flc"
    TOILETFILESUFFIX  = ".tlf"
    
    // Magic numbers
    FONTFILEMAGICNUMBER    = "flf2"
    CONTROLFILEMAGICNUMBER = "flc2"
    TOILETFILEMAGICNUMBER  = "tlf2"
    
    // Smush modes
    SM_SMUSH     = 128
    SM_KERN      = 64
    SM_EQUAL     = 1
    SM_LOWLINE   = 2
    SM_HIERARCHY = 4
    SM_PAIR      = 8
    SM_BIGX      = 16
    SM_HARDBLANK = 32
    
    // Smush override modes
    SMO_NO    = 0
    SMO_YES   = 1
    SMO_FORCE = 2
)
```

---

## Examples

### Example 1: Simple Banner

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    banner, err := figlet.RenderWithFont("Welcome!", "banner")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(banner)
}
```

### Example 2: CLI Tool with FIGlet Header

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    // Print styled header
    header, err := figlet.Render("My App",
        figlet.WithFont("slant"),
        figlet.WithWidth(60),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Print(header)
    fmt.Println("Version 1.0.0")
    fmt.Println("=".repeat(40))
    // ... rest of your app
}
```

### Example 3: All Fonts Preview

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    text := "Go"
    fonts := figlet.ListFonts()
    
    for _, font := range fonts {
        fmt.Printf("\n=== %s ===\n", font)
        result, err := figlet.RenderWithFont(text, font)
        if err != nil {
            log.Printf("Error with font %s: %v", font, err)
            continue
        }
        fmt.Print(result)
    }
}
```

### Example 4: Centered Box

```go
package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    width := 60
    
    art, err := figlet.Render("Title",
        figlet.WithFont("big"),
        figlet.WithWidth(width),
        figlet.WithJustification(1), // center
    )
    if err != nil {
        log.Fatal(err)
    }
    
    border := strings.Repeat("=", width)
    fmt.Println(border)
    fmt.Print(art)
    fmt.Println(border)
}
```

### Example 5: Using Config for Multiple Renders

```go
package main

import (
    "fmt"
    "log"

    "github.com/lsferreira42/figlet-go/figlet"
)

func main() {
    // Create and configure once
    cfg := figlet.New()
    cfg.Fontname = "small"
    cfg.Outputwidth = 40
    cfg.Justification = 1 // center
    
    if err := cfg.LoadFont(); err != nil {
        log.Fatal(err)
    }
    
    // Render multiple times with same config
    messages := []string{"Hello", "World", "FIGlet"}
    
    for _, msg := range messages {
        fmt.Print(cfg.RenderString(msg))
        fmt.Println()
    }
}
```

---

## Best Practices

1. **Reuse Config for multiple renders** - If rendering multiple strings with the same settings, create a `Config` once and reuse it.

2. **Check for errors** - Always check the error return from `Render` and `RenderWithFont`.

3. **Use embedded fonts** - The library includes 18 fonts. Using external fonts requires setting `Fontdirname`.

4. **Consider output width** - FIGlet text can be wide. Set `Outputwidth` appropriately for your use case.

5. **Test with different fonts** - Some fonts work better for certain purposes. `banner` is good for headers, `small` for compact output.

---

## See Also

- [FIGlet Official Website](http://www.figlet.org/)
- [FIGlet Font Database](http://www.figlet.org/fontdb.cgi)
- [FIGlet-Go Repository](https://github.com/lsferreira42/figlet-go)
