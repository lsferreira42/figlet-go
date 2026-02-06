package main

import (
	"sync"
	"syscall/js"
	"time"

	"github.com/lsferreira42/figlet-go/figlet"
)

var (
	configs = make(map[int]*figlet.Config)
	nextID  = 1
	mu      sync.Mutex
)

// loadFont loads the font and keeps config values that might be overwritten
func loadFont(cfg *figlet.Config) error {
	// Preserve settings that might be overwritten by LoadFont
	smushMode := cfg.Smushmode
	smushOverride := cfg.Smushoverride
	right2left := cfg.Right2left
	justification := cfg.Justification
	paragraph := cfg.Paragraphflag
	deutsch := cfg.Deutschflag

	if err := cfg.LoadFont(); err != nil {
		return err
	}

	// Restore settings only if they were explicitly changed from defaults
	if smushOverride != figlet.SMO_NO {
		cfg.Smushmode = smushMode
	}
	if smushOverride != figlet.SMO_NO {
		cfg.Smushoverride = smushOverride
	}
	if right2left != -1 {
		cfg.Right2left = right2left
	}
	if justification != -1 {
		cfg.Justification = justification
	}
	cfg.Paragraphflag = paragraph
	cfg.Deutschflag = deutsch

	return nil
}

func init() {
	mu.Lock()
	defer mu.Unlock()
	cfg := figlet.New()
	configs[0] = cfg
	// Load the default font (standard)
	loadFont(cfg)
}

// getConfig gets a config by handle or return the default if not a number
func getConfig(args []js.Value) (*figlet.Config, []js.Value) {
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		id := args[0].Int()
		mu.Lock()
		defer mu.Unlock()
		if cfg, ok := configs[id]; ok {
			return cfg, args[1:]
		}
	}
	return configs[0], args
}

// createInstance creates a new FIGlet instance and returns its handle
func createInstance(this js.Value, args []js.Value) interface{} {
	mu.Lock()
	defer mu.Unlock()
	id := nextID
	nextID++
	cfg := figlet.New()
	configs[id] = cfg
	if err := loadFont(cfg); err != nil {
		return map[string]interface{}{
			"error":  err.Error(),
			"handle": -1,
		}
	}
	return map[string]interface{}{
		"error":  nil,
		"handle": id,
	}
}

// render renders text
func render(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
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
	_, args = getConfig(args)
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
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "font name required",
			"success": false,
		}
	}

	fontName := args[0].String()
	cfg.Fontname = fontName

	if err := loadFont(cfg); err != nil {
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
	cfg, args := getConfig(args)
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
	if err := loadFont(cfg); err != nil {
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
func setJustification(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
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
func setColors(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
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
func setParser(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
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

// setSmushMode sets the smush mode
func setSmushMode(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "smush mode argument required",
			"success": false,
		}
	}
	mode := args[0].Int()
	if mode < -1 {
		cfg.Smushoverride = figlet.SMO_NO
	} else if mode == 0 {
		cfg.Smushmode = figlet.SM_KERN
		cfg.Smushoverride = figlet.SMO_YES
	} else if mode == -1 {
		cfg.Smushmode = 0
		cfg.Smushoverride = figlet.SMO_YES
	} else {
		cfg.Smushmode = (mode & 63) | figlet.SM_SMUSH
		cfg.Smushoverride = figlet.SMO_YES
	}
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setRightToLeft sets the right-to-left mode
func setRightToLeft(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "right2left argument required",
			"success": false,
		}
	}
	cfg.Right2left = args[0].Int()
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setParagraphMode sets the paragraph mode
func setParagraphMode(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "paragraph flag argument required",
			"success": false,
		}
	}
	cfg.Paragraphflag = args[0].Bool()
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// setDeutschFlag sets the deutsch flag
func setDeutschFlag(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "deutsch flag argument required",
			"success": false,
		}
	}
	cfg.Deutschflag = args[0].Bool()
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// addControlFile adds a control file
func addControlFile(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error":   "control file name required",
			"success": false,
		}
	}
	name := args[0].String()
	cfg.AddControlFile(name)
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// clearControlFiles clears all control files
func clearControlFiles(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	cfg.ClearControlFiles()
	return map[string]interface{}{
		"error":   nil,
		"success": true,
	}
}

// listAnimations returns available animations
func listAnimations(this js.Value, args []js.Value) interface{} {
	animations := figlet.ListAnimations()
	jsAnims := make([]interface{}, len(animations))
	for i, a := range animations {
		jsAnims[i] = a
	}
	return map[string]interface{}{
		"error":      nil,
		"animations": jsAnims,
	}
}

// generateAnimation generates frames for an animation
func generateAnimation(this js.Value, args []js.Value) interface{} {
	cfg, args := getConfig(args)
	if len(args) < 1 {
		return map[string]interface{}{
			"error": "text argument required",
		}
	}

	text := args[0].String()
	animType := ""
	if len(args) > 1 {
		animType = args[1].String()
	}

	delayMs := 50
	if len(args) > 2 {
		delayMs = args[2].Int()
	}

	animator := figlet.NewAnimator(cfg)
	frames, err := animator.GenerateAnimation(text, animType, time.Duration(delayMs)*time.Millisecond)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Convert frames to JS-compatible structure
	jsFrames := make([]interface{}, len(frames))
	for i, f := range frames {
		jsFrames[i] = map[string]interface{}{
			"content":        f.Content,
			"delay":          f.Delay.Milliseconds(),
			"baselineOffset": f.BaselineOffset,
		}
	}

	return map[string]interface{}{
		"error":  nil,
		"frames": jsFrames,
	}
}

func main() {
	// Register functions to be called from JavaScript
	js.Global().Set("figlet", js.ValueOf(map[string]interface{}{
		"createInstance":    js.FuncOf(createInstance),
		"render":            js.FuncOf(render),
		"renderWithFont":    js.FuncOf(renderWithFont),
		"setFont":           js.FuncOf(setFont),
		"listFonts":         js.FuncOf(listFonts),
		"getVersion":        js.FuncOf(getVersion),
		"setWidth":          js.FuncOf(setWidth),
		"setJustification":  js.FuncOf(setJustification),
		"setColors":         js.FuncOf(setColors),
		"setParser":         js.FuncOf(setParser),
		"setSmushMode":      js.FuncOf(setSmushMode),
		"setRightToLeft":    js.FuncOf(setRightToLeft),
		"setParagraph":      js.FuncOf(setParagraphMode),
		"setDeutsch":        js.FuncOf(setDeutschFlag),
		"addControlFile":    js.FuncOf(addControlFile),
		"clearControlFiles": js.FuncOf(clearControlFiles),
		"listAnimations":    js.FuncOf(listAnimations),
		"generateAnimation": js.FuncOf(generateAnimation),
	}))

	// Signal that WASM is ready in browser environment
	doc := js.Global().Get("document")
	if !doc.IsUndefined() {
		doc.Call("dispatchEvent",
			js.Global().Get("CustomEvent").New("figlet-ready"))
	}

	// Keep the program running
	c := make(chan struct{})
	<-c
}
