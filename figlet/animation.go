package figlet

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Frame represents a single frame in an animation
type Frame struct {
	Content        string
	Delay          time.Duration
	BaselineOffset int // Number of lines before the FIGlet row 0 in this frame
}

// Animator handles the generation and playback of FIGlet animations
type Animator struct {
	Config *Config
}

// NewAnimator creates a new Animator
func NewAnimator(cfg *Config) *Animator {
	return &Animator{Config: cfg}
}

// ListAnimations returns a list of available animation types
func ListAnimations() []string {
	return []string{"reveal", "scroll", "rain", "wave", "explosion"}
}

// GenerateAnimation generates frames for the specified animation type
func (a *Animator) GenerateAnimation(text string, animType string, delay time.Duration) ([]Frame, error) {
	// First, get the final rendered string to know the dimensions and content
	// We use the terminal parser to get raw geometry.
	rows, maps := a.renderToRowsAndMaps(text)
	if len(rows) == 0 {
		return nil, nil
	}

	switch strings.ToLower(animType) {
	case "reveal":
		return a.generateReveal(rows, maps, delay), nil
	case "scroll":
		return a.generateScroll(rows, maps, delay), nil
	case "rain":
		return a.generateRain(rows, maps, delay), nil
	case "wave":
		return a.generateWave(rows, maps, delay), nil
	case "explosion":
		return a.generateExplosion(rows, maps, delay), nil
	default:
		return nil, fmt.Errorf("unknown animation type: %s", animType)
	}
}

// renderToRowsAndMaps renders the text and returns it as a slice of strings (one per line)
// and a corresponding character position map.
func (a *Animator) renderToRowsAndMaps(text string) ([]string, [][]int) {
	// Remember original parser
	origParser := a.Config.OutputParser
	parser, _ := GetParser("terminal")
	a.Config.OutputParser = parser

	a.Config.PreserveMap = true
	defer func() { a.Config.PreserveMap = false }()

	rendered := a.Config.RenderString(text)

	// Capture character maps
	maps := make([][]int, len(a.Config.charPositionMap))
	for i, row := range a.Config.charPositionMap {
		maps[i] = make([]int, len(row))
		copy(maps[i], row)
	}

	// Restore original parser
	a.Config.OutputParser = origParser

	// Split by newline and remove empty trailing line if present
	lines := strings.Split(rendered, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, maps
}

// createFrame wraps the content with parser prefix/suffix and returns a Frame
func (a *Animator) createFrame(content string, delay time.Duration, baselineOffset int) Frame {
	if a.Config.OutputParser != nil {
		content = a.Config.OutputParser.Prefix + content + a.Config.OutputParser.Suffix
	}
	return Frame{Content: content, Delay: delay, BaselineOffset: baselineOffset}
}

// appendStyledRange appends a range of characters from a row using character mapping for colors
func (a *Animator) appendStyledRange(sb *strings.Builder, row string, rowMap []int, start, end int) {
	runes := []rune(row)
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return
	}

	hasColors := len(a.Config.Colors) > 0 && a.Config.OutputParser != nil && a.Config.OutputParser.Name != "terminal"

	for i := start; i < end; i++ {
		charStr := string(runes[i])
		if hasColors {
			charIndex := -1
			if i < len(rowMap) {
				charIndex = rowMap[i]
			}
			charStr = a.Config.applyColorWithIndex(charStr, charIndex)
		} else if a.Config.OutputParser != nil {
			charStr = handleReplaces(charStr, a.Config.OutputParser)
		}
		sb.WriteString(charStr)
	}
}

func (a *Animator) generateReveal(rows []string, maps [][]int, delay time.Duration) []Frame {
	width := 0
	for _, row := range rows {
		if len([]rune(row)) > width {
			width = len([]rune(row))
		}
	}

	frames := make([]Frame, 0, width+1)

	for i := 0; i <= width; i++ {
		var sb strings.Builder
		for r, row := range rows {
			rowMap := maps[r]
			a.Config.currentLineIndex = r
			runes := []rune(row)
			if i < len(runes) {
				a.appendStyledRange(&sb, row, rowMap, 0, i)
				// Fill rest with spaces (no mapping)
				a.appendStyledRange(&sb, strings.Repeat(" ", len(runes)-i), nil, 0, len(runes)-i)
			} else {
				a.appendStyledRange(&sb, row, rowMap, 0, len(runes))
			}
			sb.WriteString("\n")
		}
		frames = append(frames, a.createFrame(sb.String(), delay, 0))
	}

	return frames
}

func (a *Animator) generateScroll(rows []string, maps [][]int, delay time.Duration) []Frame {
	width := 0
	for _, row := range rows {
		if len([]rune(row)) > width {
			width = len([]rune(row))
		}
	}

	termWidth := a.Config.Outputwidth
	if termWidth <= 0 {
		termWidth = 80
	}

	frames := make([]Frame, 0, termWidth+1)

	for i := termWidth; i >= 0; i-- {
		var sb strings.Builder
		for r, row := range rows {
			rowMap := maps[r]
			a.Config.currentLineIndex = r
			// Leading spaces (no mapping)
			a.appendStyledRange(&sb, strings.Repeat(" ", i), nil, 0, i)

			// Row content (possibly truncated)
			runes := []rune(row)
			available := termWidth - i
			if available > 0 {
				end := len(runes)
				if end > available {
					end = available
				}
				a.appendStyledRange(&sb, row, rowMap, 0, end)
			}
			sb.WriteString("\n")
		}
		frames = append(frames, a.createFrame(sb.String(), delay, 0))
	}

	return frames
}

func (a *Animator) generateRain(rows []string, maps [][]int, delay time.Duration) []Frame {
	height := len(rows)
	width := 0
	for _, row := range rows {
		if len([]rune(row)) > width {
			width = len([]rune(row))
		}
	}

	numFrames := height + 15
	frames := make([]Frame, 0, numFrames)

	for f := 0; f < numFrames; f++ {
		grid := make([][]rune, height)
		gridMap := make([][]int, height)
		for i := range grid {
			grid[i] = make([]rune, width)
			gridMap[i] = make([]int, width)
			for j := range grid[i] {
				grid[i][j] = ' '
				gridMap[i][j] = -1
			}
		}

		for r := 0; r < height; r++ {
			rowRunes := []rune(rows[r])
			rowMap := maps[r]
			for c := 0; c < len(rowRunes); c++ {
				if rowRunes[c] == ' ' {
					continue
				}
				delayColumn := (c / 2) % 10
				reachFrame := r + delayColumn

				if f >= reachFrame {
					grid[r][c] = rowRunes[c]
					if c < len(rowMap) {
						gridMap[r][c] = rowMap[c]
					}
				} else {
					currR := f - delayColumn
					if currR >= 0 && currR < height {
						grid[currR][c] = rowRunes[c]
						if c < len(rowMap) {
							gridMap[currR][c] = rowMap[c]
						}
					}
				}
			}
		}

		var sb strings.Builder
		for r, gridRow := range grid {
			a.Config.currentLineIndex = r
			rowStr := string(gridRow)
			trimmedRow := strings.TrimRight(rowStr, " ")
			runes := []rune(trimmedRow)
			a.appendStyledRange(&sb, trimmedRow, gridMap[r][:len(runes)], 0, len(runes))
			sb.WriteString("\n")
		}
		frames = append(frames, a.createFrame(sb.String(), delay, 0))
	}

	return frames
}

func (a *Animator) generateWave(rows []string, maps [][]int, delay time.Duration) []Frame {
	numFrames := 40
	frames := make([]Frame, 0, numFrames)

	for f := 0; f < numFrames; f++ {
		var sb strings.Builder
		phase := float64(f) * 0.5
		dampening := 1.0 - float64(f)/float64(numFrames-1)

		for r := 0; r < len(rows); r++ {
			row := rows[r]
			rowMap := maps[r]
			a.Config.currentLineIndex = r
			runes := []rune(row)
			shift := int(5.0 * dampening * math.Sin(phase+float64(r)*0.5))

			if shift > 0 {
				a.appendStyledRange(&sb, strings.Repeat(" ", shift), nil, 0, shift)
				a.appendStyledRange(&sb, row, rowMap, 0, len(runes))
			} else if shift < 0 {
				start := -shift
				if start < len(runes) {
					a.appendStyledRange(&sb, row, rowMap, start, len(runes))
				}
			} else {
				a.appendStyledRange(&sb, row, rowMap, 0, len(runes))
			}
			sb.WriteString("\n")
		}
		frames = append(frames, a.createFrame(sb.String(), delay, 0))
	}

	return frames
}

func (a *Animator) generateExplosion(rows []string, maps [][]int, delay time.Duration) []Frame {
	height := len(rows)

	// Capture the initial static content and mappings for pauses
	var staticSb strings.Builder
	for r, row := range rows {
		a.Config.currentLineIndex = r
		a.appendStyledRange(&staticSb, row, maps[r], 0, len([]rune(row)))
		staticSb.WriteString("\n")
	}
	staticContent := staticSb.String()

	numStaticStart := 8
	frames := make([]Frame, 0, 70)
	for i := 0; i < numStaticStart; i++ {
		frames = append(frames, a.createFrame(staticContent, delay, 0))
	}

	numFrames := 40
	type particle struct {
		char      rune
		charIndex int
		row, col  int
		vx, vy    float64
	}

	var particles []particle
	for r, row := range rows {
		runes := []rune(row)
		rowMap := maps[r]
		for c, char := range runes {
			if char != ' ' {
				charIndex := -1
				if c < len(rowMap) {
					charIndex = rowMap[c]
				}
				angle := rand.Float64() * 2 * math.Pi
				speed := rand.Float64() * 3.0
				particles = append(particles, particle{
					char:      char,
					charIndex: charIndex,
					row:       r,
					col:       c,
					vx:        math.Cos(angle) * speed * 2.0,
					vy:        math.Sin(angle) * speed * 0.4,
				})
			}
		}
	}

	explosionPositions := make([]struct{ x, y float64 }, len(particles))
	for i := range particles {
		p := particles[i]
		x, y := float64(p.col), float64(p.row)
		vx, vy := p.vx, p.vy
		for f := 0; f < numFrames/2; f++ {
			x += vx
			y += vy
			vx *= 0.92
			vy *= 0.92
		}
		explosionPositions[i] = struct{ x, y float64 }{x, y}
	}

	for f := 0; f < numFrames; f++ {
		gridHeight := height + 10
		targetWidth := a.Config.Outputwidth
		if targetWidth <= 0 {
			targetWidth = 80
		}

		grid := make([][]rune, gridHeight)
		gridMap := make([][]int, gridHeight)
		for i := range grid {
			grid[i] = make([]rune, targetWidth)
			gridMap[i] = make([]int, targetWidth)
			for j := range grid[i] {
				grid[i][j] = ' '
				gridMap[i][j] = -1
			}
		}

		offsetY := 5
		for i := range particles {
			p := &particles[i]
			var x, y float64
			if f < numFrames/2 {
				x, y = float64(p.col), float64(p.row)
				vx, vy := p.vx, p.vy
				for j := 0; j < f; j++ {
					x += vx
					y += vy
					vx *= 0.92
					vy *= 0.92
				}
			} else {
				startPos := explosionPositions[i]
				targetX, targetY := float64(p.col), float64(p.row)
				t := float64(f-numFrames/2) / float64(numFrames/2-1)
				t = t * t * (3 - 2*t)
				x = startPos.x + (targetX-startPos.x)*t
				y = startPos.y + (targetY-startPos.y)*t
			}

			ix, iy := int(x), int(y+float64(offsetY))
			if iy >= 0 && iy < len(grid) && ix >= 0 && ix < len(grid[iy]) {
				grid[iy][ix] = p.char
				gridMap[iy][ix] = p.charIndex
			}
		}

		var sb strings.Builder
		for r, gridRow := range grid {
			rowStr := string(gridRow)
			trimmedRow := strings.TrimRight(rowStr, " ")
			runes := []rune(trimmedRow)
			a.appendStyledRange(&sb, trimmedRow, gridMap[r][:len(runes)], 0, len(runes))
			sb.WriteString("\n")
		}
		frames = append(frames, a.createFrame(sb.String(), delay, offsetY))
	}

	frames = append(frames, a.createFrame(staticContent, delay, 0))

	return frames
}

// PlayAnimation plays the animation with terminal control codes OR as a standalone HTML player.
func PlayAnimation(cfg *Config, frames []Frame) {
	if len(frames) == 0 {
		return
	}

	// For HTML output, we generate a standalone player
	if cfg.OutputParser != nil && cfg.OutputParser.Name == "html" {
		playHTMLAnimation(frames)
		return
	}

	// Default: Terminal playback with ANSI codes
	fmt.Print("\033[?25l")       // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor

	lastTotalLines := 0
	lastBaselineOffset := 0

	for i, frame := range frames {
		contentLines := strings.Split(strings.TrimSuffix(frame.Content, "\n"), "\n")

		if i > 0 {
			if lastTotalLines > 0 {
				fmt.Printf("\033[%dA", lastTotalLines)
			}
			diff := frame.BaselineOffset - lastBaselineOffset
			if diff > 0 {
				fmt.Printf("\033[%dA", diff)
			} else if diff < 0 {
				fmt.Printf("\033[%dB", -diff)
			}
		} else {
			if frame.BaselineOffset > 0 {
				fmt.Printf("\033[%dA", frame.BaselineOffset)
			}
		}

		for _, line := range contentLines {
			fmt.Print(line)
			fmt.Print("\033[K\n")
		}

		lastTotalLines = len(contentLines)
		lastBaselineOffset = frame.BaselineOffset
		time.Sleep(frame.Delay)
	}
}

// playHTMLAnimation generates a standalone HTML player for the animation.
func playHTMLAnimation(frames []Frame) {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>FIGlet Animation</title>\n")
	sb.WriteString("<style>\n")
	sb.WriteString("  body { background: #0c0c0c; color: #cccccc; font-family: 'Cascadia Code', 'Ubuntu Mono', 'Roboto Mono', 'DejaVu Sans Mono', monospace; margin: 0; padding: 20px; overflow: auto; }\n")
	sb.WriteString("  #terminal { white-space: pre; line-height: 1.25; font-size: 14px; position: relative; }\n")
	sb.WriteString("</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<div id='terminal'></div>\n")
	sb.WriteString("<script>\n")
	sb.WriteString("  const frames = [\n")

	for _, frame := range frames {
		// Escape backticks and backslashes for JS template literal
		content := strings.ReplaceAll(frame.Content, "\\", "\\\\")
		content = strings.ReplaceAll(content, "`", "\\`")
		content = strings.ReplaceAll(content, "${", "\\${")

		sb.WriteString(fmt.Sprintf("    { c: `%s`, d: %d, o: %d },\n",
			content, frame.Delay.Milliseconds(), frame.BaselineOffset))
	}

	sb.WriteString("  ];\n")
	sb.WriteString("  const term = document.getElementById('terminal');\n")
	sb.WriteString("  let idx = 0;\n")
	sb.WriteString("  const LINE_HEIGHT = 17.5;\n")
	sb.WriteString("\n")
	sb.WriteString("  function update() {\n")
	sb.WriteString("    const frame = frames[idx];\n")
	sb.WriteString("    term.innerHTML = frame.c;\n")
	sb.WriteString("    term.style.marginTop = (frame.o * LINE_HEIGHT) + 'px';\n")
	sb.WriteString("    const delay = frame.d || 50;\n")
	sb.WriteString("    idx = (idx + 1) % frames.length;\n")
	sb.WriteString("    setTimeout(update, delay);\n")
	sb.WriteString("  }\n")
	sb.WriteString("  if (frames.length > 0) update();\n")
	sb.WriteString("</script>\n")
	sb.WriteString("</body>\n</html>\n")

	fmt.Print(sb.String())
}
