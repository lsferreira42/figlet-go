//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/lsferreira42/figlet-go/figlet"
)

var cfg *figlet.Config
var initError error

func init() {
	cfg = figlet.New()
	// Load the default font (standard)
	initError = cfg.LoadFont()
}

// render renders text with the current font
func render(this js.Value, args []js.Value) interface{} {
	if initError != nil {
		return map[string]interface{}{
			"error":  "font not loaded: " + initError.Error(),
			"result": "",
		}
	}

	if len(args) < 1 {
		return map[string]interface{}{
			"error":  "text argument required",
			"result": "",
		}
	}

	text := args[0].String()
	result := cfg.RenderString(text)

	return map[string]interface{}{
		"error":  nil,
		"result": result,
	}
}

// renderWithFont renders text with a specific font
func renderWithFont(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return map[string]interface{}{
			"error":  "text and font arguments required",
			"result": "",
		}
	}

	text := args[0].String()
	fontName := args[1].String()

	result, err := figlet.RenderWithFont(text, fontName)
	if err != nil {
		return map[string]interface{}{
			"error":  err.Error(),
			"result": "",
		}
	}

	return map[string]interface{}{
		"error":  nil,
		"result": result,
	}
}

// setFont sets the current font
func setFont(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "font name required",
			"success": false,
		}
	}

	fontName := args[0].String()
	cfg.Fontname = fontName

	if err := cfg.LoadFont(); err != nil {
		return map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
	}

	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// listFonts returns available fonts
func listFonts(this js.Value, args []js.Value) interface{} {
	fonts := figlet.ListFonts()
	if fonts == nil {
		return map[string]interface{}{
			"error": "failed to list fonts",
			"fonts": []interface{}{},
		}
	}

	// Convert to JS-compatible slice
	jsFonts := make([]interface{}, len(fonts))
	for i, f := range fonts {
		jsFonts[i] = f
	}

	return map[string]interface{}{
		"error": nil,
		"fonts": jsFonts,
	}
}

// getVersion returns the FIGlet version
func getVersion(this js.Value, args []js.Value) interface{} {
	return figlet.GetVersion()
}

// setWidth sets the output width
func setWidth(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "width argument required",
			"success": false,
		}
	}
	width := args[0].Int()
	if width < 1 {
		return map[string]interface{}{
			"error":   "width must be positive",
			"success": false,
		}
	}

	cfg.Outputwidth = width
	// Reload font to recalculate internal buffers with new width
	if err := cfg.LoadFont(); err != nil {
		return map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
	}

	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setJustification sets text justification
// 0 = left, 1 = center, 2 = right, -1 = auto
func setJustification(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "justification argument required",
			"success": false,
		}
	}
	cfg.Justification = args[0].Int()
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setColors sets colors for rendering
// Accepts an array of color strings (e.g., ["red", "green", "blue"] or ["FF0000", "00FF00"])
func setColors(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "colors array required",
			"success": false,
		}
	}

	colorsArray := args[0]
	if colorsArray.Type() != js.TypeObject {
		return map[string]interface{}{
			"error":   "colors must be an array",
			"success": false,
		}
	}

	length := colorsArray.Length()
	colors := make([]figlet.Color, 0, length)

	for i := 0; i < length; i++ {
		colorStr := colorsArray.Index(i).String()
		if colorStr == "" {
			continue
		}

		// Try predefined color names first
		var color figlet.Color
		switch colorStr {
		case "black":
			color = figlet.ColorBlack
		case "red":
			color = figlet.ColorRed
		case "green":
			color = figlet.ColorGreen
		case "yellow":
			color = figlet.ColorYellow
		case "blue":
			color = figlet.ColorBlue
		case "magenta":
			color = figlet.ColorMagenta
		case "cyan":
			color = figlet.ColorCyan
		case "white":
			color = figlet.ColorWhite
		default:
			// Try to parse as hex color
			tc, err := figlet.NewTrueColorFromHexString(colorStr)
			if err != nil {
				return map[string]interface{}{
					"error":   "invalid color: " + colorStr,
					"success": false,
				}
			}
			color = tc
		}
		colors = append(colors, color)
	}

	cfg.Colors = colors
	// If colors are set, default to terminal-color parser if not already set
	if len(colors) > 0 && (cfg.OutputParser == nil || cfg.OutputParser.Name == "terminal") {
		parser, _ := figlet.GetParser("terminal-color")
		cfg.OutputParser = parser
	}

	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setParser sets the output parser
// Accepts: "terminal", "terminal-color", or "html"
func setParser(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "parser name required",
			"success": false,
		}
	}

	parserName := args[0].String()
	parser, err := figlet.GetParser(parserName)
	if err != nil {
		return map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
	}

	cfg.OutputParser = parser
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

func main() {
	c := make(chan struct{}, 0)

	// Register functions to be called from JavaScript
	js.Global().Set("figlet", js.ValueOf(map[string]interface{}{
		"render":           js.FuncOf(render),
		"renderWithFont":   js.FuncOf(renderWithFont),
		"setFont":          js.FuncOf(setFont),
		"listFonts":        js.FuncOf(listFonts),
		"getVersion":       js.FuncOf(getVersion),
		"setWidth":         js.FuncOf(setWidth),
		"setJustification": js.FuncOf(setJustification),
		"setColors":         js.FuncOf(setColors),
		"setParser":         js.FuncOf(setParser),
	}))

	// Signal that WASM is ready
	js.Global().Get("document").Call("dispatchEvent",
		js.Global().Get("CustomEvent").New("figlet-ready"))

	// Keep the program running
	<-c
}
