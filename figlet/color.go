// Package figlet provides color support for FIGlet rendering.
package figlet

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// Escape character for ANSI codes
const escape = "\x1b"

// Color interface defines methods for color formatting
type Color interface {
	getPrefix(parser *OutputParser) string
	getSuffix(parser *OutputParser) string
}

// AnsiColor represents an ANSI color code
type AnsiColor struct {
	code int
}

// Predefined ANSI colors
var (
	ColorBlack   = AnsiColor{30}
	ColorRed     = AnsiColor{31}
	ColorGreen   = AnsiColor{32}
	ColorYellow  = AnsiColor{33}
	ColorBlue    = AnsiColor{34}
	ColorMagenta = AnsiColor{35}
	ColorCyan    = AnsiColor{36}
	ColorWhite   = AnsiColor{37}
)

// TrueColor represents a 24-bit RGB color
type TrueColor struct {
	R int
	G int
	B int
}

// TrueColor lookalikes for displaying AnsiColor (e.g., with HTML parser)
// Colors based on http://clrs.cc/
var tcfac = map[AnsiColor]TrueColor{
	ColorBlack:   {0, 0, 0},
	ColorRed:     {255, 65, 54},
	ColorGreen:   {149, 189, 64},
	ColorYellow:  {255, 220, 0},
	ColorBlue:    {0, 116, 217},
	ColorMagenta: {177, 13, 201},
	ColorCyan:    {105, 206, 245},
	ColorWhite:   {255, 255, 255},
}

// getPrefix returns the prefix for TrueColor based on parser type
func (tc TrueColor) getPrefix(parser *OutputParser) string {
	switch parser.Name {
	case "terminal-color":
		return fmt.Sprintf("%s[38;2;%d;%d;%dm", escape, tc.R, tc.G, tc.B)
	case "html":
		return fmt.Sprintf("<span style='color: rgb(%d,%d,%d);'>", tc.R, tc.G, tc.B)
	}
	return ""
}

// getSuffix returns the suffix for TrueColor based on parser type
func (tc TrueColor) getSuffix(parser *OutputParser) string {
	switch parser.Name {
	case "terminal-color":
		return fmt.Sprintf("%s[0m", escape)
	case "html":
		return "</span>"
	}
	return ""
}

// NewTrueColorFromHexString creates a TrueColor from a hexadecimal string (e.g., "FF0000" or "#FF0000")
func NewTrueColorFromHexString(hexStr string) (*TrueColor, error) {
	// Remove # if present
	if len(hexStr) > 0 && hexStr[0] == '#' {
		hexStr = hexStr[1:]
	}
	
	// Must be 6 characters for RGB
	if len(hexStr) != 6 {
		return nil, errors.New("hex color must be 6 characters (e.g., 'FF0000' or '#FF0000')")
	}
	
	rgb, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hexStr)
	}
	
	if len(rgb) < 3 {
		return nil, errors.New("invalid hex color format")
	}
	
	return &TrueColor{
		R: int(rgb[0]),
		G: int(rgb[1]),
		B: int(rgb[2]),
	}, nil
}

// getPrefix returns the prefix for AnsiColor based on parser type
func (ac AnsiColor) getPrefix(parser *OutputParser) string {
	switch parser.Name {
	case "terminal-color":
		return fmt.Sprintf("%s[0;%dm", escape, ac.code)
	case "html":
		// Get the TrueColor for the AnsiColor
		tc := tcfac[ac]
		return tc.getPrefix(parser)
	}
	return ""
}

// getSuffix returns the suffix for AnsiColor based on parser type
func (ac AnsiColor) getSuffix(parser *OutputParser) string {
	switch parser.Name {
	case "terminal-color":
		return fmt.Sprintf("%s[0m", escape)
	case "html":
		// Get the TrueColor for the AnsiColor
		tc := tcfac[ac]
		return tc.getSuffix(parser)
	}
	return ""
}
