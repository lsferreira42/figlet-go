// Package figlet provides output parsers for different formats.
package figlet

import (
	"errors"
	"strings"
)

// OutputParser defines how to format the output
type OutputParser struct {
	// Name of the parser (used for switching in color methods)
	Name string
	// Prefix to add before the output
	Prefix string
	// Suffix to add after the output
	Suffix string
	// Newline representation
	NewLine string
	// Character replacements (e.g., " " to "&nbsp;" for HTML)
	Replaces map[string]string
}

var parsers = map[string]OutputParser{
	// Default terminal parser (no colors)
	"terminal": {
		Name:    "terminal",
		Prefix:  "",
		Suffix:  "",
		NewLine: "\n",
		Replaces: nil,
	},
	// Terminal parser with ANSI color support
	"terminal-color": {
		Name:    "terminal-color",
		Prefix:  "",
		Suffix:  "",
		NewLine: "\n",
		Replaces: nil,
	},
	// HTML parser
	"html": {
		Name:    "html",
		Prefix:  "<code>",
		Suffix:  "</code>",
		NewLine: "<br>",
		Replaces: map[string]string{
			" ": "&nbsp;",
		},
	},
}

// GetParser returns a parser by its key
func GetParser(key string) (*OutputParser, error) {
	parser, ok := parsers[key]
	if !ok {
		return nil, errors.New("invalid parser key: " + key + " (valid: terminal, terminal-color, html)")
	}
	return &parser, nil
}

// handleReplaces applies character replacements based on parser configuration
func handleReplaces(str string, parser *OutputParser) string {
	if parser.Replaces == nil {
		return str
	}
	result := str
	for old, new := range parser.Replaces {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}
