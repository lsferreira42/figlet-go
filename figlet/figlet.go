// Package figlet provides FIGlet text rendering functionality.
// It can be used as a library to render ASCII art text.
package figlet

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

//go:embed fonts/*.flf fonts/*.flc
var embeddedFonts embed.FS

const (
	DATE        = "31 May 2012"
	VERSION     = "2.2.5"
	VERSION_INT = 20205

	FONTFILESUFFIX         = ".flf"
	FONTFILEMAGICNUMBER    = "flf2"
	CONTROLFILESUFFIX      = ".flc"
	CONTROLFILEMAGICNUMBER = "flc2"
	TOILETFILESUFFIX       = ".tlf"
	TOILETFILEMAGICNUMBER  = "tlf2"
	DEFAULTCOLUMNS         = 80
	MAXLEN                 = 255

	SM_SMUSH     = 128
	SM_KERN      = 64
	SM_EQUAL     = 1
	SM_LOWLINE   = 2
	SM_HIERARCHY = 4
	SM_PAIR      = 8
	SM_BIGX      = 16
	SM_HARDBLANK = 32

	SMO_NO    = 0
	SMO_YES   = 1
	SMO_FORCE = 2
)

var (
	deutsch = []rune{196, 214, 220, 228, 246, 252, 223}
)

// FCharNode represents a character in the font
type FCharNode struct {
	ord     rune
	thechar [][]rune
	next    *FCharNode
}

// CFNameNode represents a control file name node
type CFNameNode struct {
	thename string
	next    *CFNameNode
}

// ComNode represents a command node for character mapping
type ComNode struct {
	thecommand int
	rangelo    rune
	rangehi    rune
	offset     rune
	next       *ComNode
}

// Config holds the FIGlet configuration and state
type Config struct {
	Deutschflag       bool
	Justification     int // -1 = auto, 0 = left, 1 = center, 2 = right
	Paragraphflag     bool
	Right2left        int // -1 = auto, 0 = left, 1 = right
	Multibyte         int // 0 = ISO 2022, 1 = DBCS, 2 = UTF-8, 3 = HZ, 4 = Shift-JIS
	Cmdinput          bool
	Smushmode         int
	Smushoverride     int
	Outputwidth       int
	Fontdirname       string
	Fontname          string
	cfilelist         *CFNameNode
	cfilelistend      **CFNameNode
	commandlist       *ComNode
	commandlistend    **ComNode
	hardblank         rune
	charheight        int
	fcharlist         *FCharNode
	outputline        [][]rune
	outlinelen        int
	outlinelenlimit   int
	inchrline         []rune
	inchrlinelen      int
	inchrlinelenlimit int
	currchar          [][]rune
	currcharwidth     int
	previouscharwidth int
	hzmode            bool
	gndbl             [4]bool
	gn                [4]rune
	gl                int
	gr                int
	toiletfont        bool
	getinchr_buffer   rune
	getinchr_flag     bool
	Optind            int
	Argv              []string
	agetmode          int // >= 0 for displacement into argv[n], <0 EOF
	output            *strings.Builder
	// Color support
	Colors       []Color
	OutputParser *OutputParser
	// Track current character index for color cycling
	currentCharIndex int
	// Track which input character is at each output position for each line
	// Maps line index -> column index -> input character index
	charPositionMap [][]int
	// Current line being built (for charPositionMap)
	currentLineIndex int
}

// New creates a new Config with default values
func New() *Config {
	cfg := &Config{
		Justification: -1,
		Right2left:    -1,
		Outputwidth:   DEFAULTCOLUMNS,
		gr:            1,
		gn:            [4]rune{0, 0x80, 0, 0},
		Fontdirname:   "fonts",
		Fontname:      "standard",
		Smushoverride: SMO_NO,
	}
	cfg.cfilelistend = &cfg.cfilelist
	cfg.commandlistend = &cfg.commandlist
	// Default parser is terminal (no colors)
	parser, _ := GetParser("terminal")
	cfg.OutputParser = parser
	return cfg
}

// Option is a function type for configuring the FIGlet instance
type Option func(*Config)

// WithFont sets the font name
func WithFont(name string) Option {
	return func(cfg *Config) {
		cfg.Fontname = name
		if suffixcmp(cfg.Fontname, FONTFILESUFFIX) {
			cfg.Fontname = cfg.Fontname[:len(cfg.Fontname)-len(FONTFILESUFFIX)]
		} else if suffixcmp(cfg.Fontname, TOILETFILESUFFIX) {
			cfg.Fontname = cfg.Fontname[:len(cfg.Fontname)-len(TOILETFILESUFFIX)]
		}
	}
}

// WithFontDir sets the font directory
func WithFontDir(dir string) Option {
	return func(cfg *Config) {
		cfg.Fontdirname = dir
	}
}

// WithWidth sets the output width
func WithWidth(width int) Option {
	return func(cfg *Config) {
		if width > 0 {
			cfg.Outputwidth = width
		}
	}
}

// WithJustification sets the text justification (-1=auto, 0=left, 1=center, 2=right)
func WithJustification(j int) Option {
	return func(cfg *Config) {
		cfg.Justification = j
	}
}

// WithRightToLeft sets the right-to-left mode (-1=auto, 0=left, 1=right)
func WithRightToLeft(r int) Option {
	return func(cfg *Config) {
		cfg.Right2left = r
	}
}

// WithSmushMode sets the smush mode
func WithSmushMode(mode int) Option {
	return func(cfg *Config) {
		if mode < -1 {
			cfg.Smushoverride = SMO_NO
			return
		}
		if mode == 0 {
			cfg.Smushmode = SM_KERN
		} else if mode == -1 {
			cfg.Smushmode = 0
		} else {
			cfg.Smushmode = (mode & 63) | SM_SMUSH
		}
		cfg.Smushoverride = SMO_YES
	}
}

// WithKerning enables kerning mode
func WithKerning() Option {
	return func(cfg *Config) {
		cfg.Smushmode = SM_KERN
		cfg.Smushoverride = SMO_YES
	}
}

// WithFullWidth disables smushing
func WithFullWidth() Option {
	return func(cfg *Config) {
		cfg.Smushmode = 0
		cfg.Smushoverride = SMO_YES
	}
}

// WithSmushing enables smushing
func WithSmushing() Option {
	return func(cfg *Config) {
		cfg.Smushmode = SM_SMUSH
		cfg.Smushoverride = SMO_FORCE
	}
}

// WithOverlapping enables overlapping mode
func WithOverlapping() Option {
	return func(cfg *Config) {
		cfg.Smushmode = SM_SMUSH
		cfg.Smushoverride = SMO_YES
	}
}

// WithColors sets the colors to use for rendering
func WithColors(colors ...Color) Option {
	return func(cfg *Config) {
		cfg.Colors = colors
		// If colors are set and parser is still default terminal, switch to terminal-color
		// But don't override if user explicitly set a parser (like HTML)
		if len(colors) > 0 && cfg.OutputParser != nil && cfg.OutputParser.Name == "terminal" {
			parser, _ := GetParser("terminal-color")
			cfg.OutputParser = parser
		}
	}
}

// WithParser sets the output parser
func WithParser(parserName string) Option {
	return func(cfg *Config) {
		parser, err := GetParser(parserName)
		if err == nil {
			cfg.OutputParser = parser
		}
	}
}

// WithOutputParser sets the output parser directly
func WithOutputParser(parser *OutputParser) Option {
	return func(cfg *Config) {
		cfg.OutputParser = parser
	}
}

// Render renders the given text using FIGlet and returns the result as a string
func Render(text string, options ...Option) (string, error) {
	cfg := New()
	for _, opt := range options {
		opt(cfg)
	}

	if err := cfg.LoadFont(); err != nil {
		return "", err
	}

	return cfg.RenderString(text), nil
}

// RenderWithFont is a convenience function to render text with a specific font
func RenderWithFont(text, fontName string) (string, error) {
	return Render(text, WithFont(fontName))
}

// LoadFont loads the font specified in the config
func (cfg *Config) LoadFont() error {
	cfg.outlinelenlimit = cfg.Outputwidth - 1
	readcontrolfiles(cfg)
	if err := readfont(cfg); err != nil {
		return err
	}
	linealloc(cfg)
	return nil
}

// RenderString renders the given text and returns the result as a string
func (cfg *Config) RenderString(text string) string {
	cfg.output = &strings.Builder{}
	cfg.Cmdinput = true
	cfg.Argv = []string{"figlet", text}
	cfg.Optind = 1
	cfg.agetmode = 0
	cfg.currentCharIndex = 0
	cfg.currentLineIndex = 0
	cfg.charPositionMap = make([][]int, cfg.charheight)
	for i := range cfg.charPositionMap {
		cfg.charPositionMap[i] = make([]int, 0, 100)
	}

	// Write parser prefix if any
	if cfg.OutputParser != nil && cfg.OutputParser.Prefix != "" {
		cfg.output.WriteString(cfg.OutputParser.Prefix)
	}

	wordbreakmode := 0
	last_was_eol_flag := false

	for {
		c := getinchr(cfg)
		if c == -1 { // EOF
			break
		}

		if c == '\n' && cfg.Paragraphflag && !last_was_eol_flag {
			c2 := getinchr(cfg)
			ungetinchr(cfg, c2)
			if isASCII(c2) && unicode.IsSpace(c2) {
				c = '\n'
			} else {
				c = ' '
			}
		}
		last_was_eol_flag = isASCII(c) && unicode.IsSpace(c) && c != '\t' && c != ' '

		if cfg.Deutschflag {
			if c >= '[' && c <= ']' {
				c = deutsch[c-'[']
			} else if c >= '{' && c <= '~' {
				c = deutsch[c-'{'+3]
			}
		}

		c = handlemapping(cfg, c)

		if isASCII(c) && unicode.IsSpace(c) {
			if c == '\t' || c == ' ' {
				c = ' '
			} else {
				c = '\n'
			}
		}

		if (c > 0 && c < ' ' && c != '\n') || c == 127 {
			continue
		}

		for {
			char_not_added := false

			if wordbreakmode == -1 {
				if c == ' ' {
					break
				} else if c == '\n' {
					wordbreakmode = 0
					break
				}
				wordbreakmode = 0
			}

			if c == '\n' {
				cfg.printline()
				wordbreakmode = 0
			} else if cfg.addchar(c) {
				if c != ' ' {
					if wordbreakmode >= 2 {
						wordbreakmode = 3
					} else {
						wordbreakmode = 1
					}
				} else {
					if wordbreakmode > 0 {
						wordbreakmode = 2
					} else {
						wordbreakmode = 0
					}
				}
			} else if cfg.outlinelen == 0 {
				for i := 0; i < cfg.charheight; i++ {
					if cfg.Right2left == 1 && cfg.Outputwidth > 1 {
						start := len(cfg.currchar[i]) - cfg.outlinelenlimit
						if start < 0 {
							start = 0
						}
						cfg.putstring(cfg.currchar[i][start:])
					} else {
						cfg.putstring(cfg.currchar[i])
					}
				}
				wordbreakmode = -1
			} else if c == ' ' {
				if wordbreakmode == 2 {
					cfg.splitline()
				} else {
					cfg.printline()
				}
				wordbreakmode = -1
			} else {
				if wordbreakmode >= 2 {
					cfg.splitline()
				} else {
					cfg.printline()
				}
				if wordbreakmode == 3 {
					wordbreakmode = 1
				} else {
					wordbreakmode = 0
				}
				char_not_added = true
			}

			if !char_not_added {
				break
			}
		}
	}

	if cfg.outlinelen != 0 {
		cfg.printline()
	}

	// Write parser suffix if any
	if cfg.OutputParser != nil && cfg.OutputParser.Suffix != "" {
		cfg.output.WriteString(cfg.OutputParser.Suffix)
	}

	return cfg.output.String()
}

// ListFonts returns a list of available fonts from the embedded fonts
func ListFonts() []string {
	entries, err := embeddedFonts.ReadDir("fonts")
	if err != nil {
		return nil
	}
	var fonts []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, FONTFILESUFFIX) {
			fonts = append(fonts, strings.TrimSuffix(name, FONTFILESUFFIX))
		} else if strings.HasSuffix(name, TOILETFILESUFFIX) {
			fonts = append(fonts, strings.TrimSuffix(name, TOILETFILESUFFIX))
		}
	}
	return fonts
}

// GetVersion returns the FIGlet version string
func GetVersion() string {
	return VERSION
}

// GetVersionInt returns the FIGlet version as an integer
func GetVersionInt() int {
	return VERSION_INT
}

func isASCII(r rune) bool {
	return r >= 0 && r <= 127
}

func suffixcmp(s1, s2 string) bool {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)
	return strings.HasSuffix(s1, s2)
}

func hasdirsep(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, "\\")
}

func (cfg *Config) clearcfilelist() {
	cfg.cfilelist = nil
	cfg.cfilelistend = &cfg.cfilelist
}

// ZFILE emulation for reading compressed files
type ZFILE struct {
	reader    io.Reader
	buffer    []byte
	pos       int
	isZip     bool
	zipFile   *zip.File
	zipReader io.ReadCloser
	file      *os.File // For filesystem files that need to be closed
}

func Zopen(path string, mode string) (*ZFILE, error) {
	// Try embedded fonts first
	if strings.HasPrefix(path, "fonts/") || !strings.Contains(path, "/") {
		// Try embedded
		data, err := embeddedFonts.ReadFile(path)
		if err == nil {
			// Check if it's a zip file
			if len(data) >= 4 && string(data[0:4]) == "PK\x03\x04" {
				// It's a zip file
				zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
				if err != nil {
					return nil, err
				}
				if len(zipReader.File) > 0 {
					zf := zipReader.File[0]
					rc, err := zf.Open()
					if err != nil {
						return nil, err
					}
					return &ZFILE{
						reader:    rc,
						isZip:     true,
						zipFile:   zf,
						zipReader: rc,
					}, nil
				}
			}
			return &ZFILE{
				reader: bytes.NewReader(data),
			}, nil
		}
	}

	// Try filesystem
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Don't defer close here - we need to keep the file open for reading

	// Check if it's a zip file
	header := make([]byte, 4)
	n, _ := file.Read(header)
	file.Seek(0, 0)
	if n == 4 && string(header) == "PK\x03\x04" {
		// It's a zip file
		fi, _ := file.Stat()
		zipReader, err := zip.NewReader(file, fi.Size())
		if err != nil {
			file.Close()
			return nil, err
		}
		if len(zipReader.File) > 0 {
			zf := zipReader.File[0]
			rc, err := zf.Open()
			if err != nil {
				file.Close()
				return nil, err
			}
			return &ZFILE{
				reader:    rc,
				isZip:     true,
				zipFile:   zf,
				zipReader: rc,
				file:      file, // Keep file open for zip reader
			}, nil
		}
		file.Close()
	}

	file.Seek(0, 0)
	return &ZFILE{
		reader: file,
		file:   file,
	}, nil
}

func Zgetc(zf *ZFILE) int {
	if zf.buffer == nil || zf.pos >= len(zf.buffer) {
		buf := make([]byte, 4096)
		n, err := zf.reader.Read(buf)
		if err != nil && n == 0 {
			return -1
		}
		zf.buffer = buf[:n]
		zf.pos = 0
	}
	if zf.pos >= len(zf.buffer) {
		return -1
	}
	b := zf.buffer[zf.pos]
	zf.pos++
	return int(b)
}

func Zungetc(c int, zf *ZFILE) {
	if zf.pos > 0 {
		zf.pos--
	}
}

func Zclose(zf *ZFILE) error {
	var err error
	if zf.zipReader != nil {
		err = zf.zipReader.Close()
	}
	if zf.file != nil {
		if closeErr := zf.file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func myfgets(line []byte, maxlen int, zf *ZFILE) []byte {
	p := 0
	for p < maxlen-1 {
		c := Zgetc(zf)
		if c == -1 {
			if p == 0 {
				return nil
			}
			break
		}
		line[p] = byte(c)
		p++
		if c == '\n' {
			break
		}
		if c == '\r' {
			c2 := Zgetc(zf)
			if c2 != -1 && c2 != '\n' {
				Zungetc(c2, zf)
			}
			line[p-1] = '\n'
			break
		}
	}
	if p > 0 {
		return line[:p]
	}
	return nil
}

func skiptoeol(zf *ZFILE) {
	for {
		c := Zgetc(zf)
		if c == -1 || c == '\n' {
			return
		}
		if c == '\r' {
			c2 := Zgetc(zf)
			if c2 != -1 && c2 != '\n' {
				Zungetc(c2, zf)
			}
			return
		}
	}
}

func readmagic(zf *ZFILE) string {
	magic := make([]byte, 4)
	for i := 0; i < 4; i++ {
		c := Zgetc(zf)
		if c == -1 {
			return ""
		}
		magic[i] = byte(c)
	}
	return string(magic)
}

func skipws(zf *ZFILE) {
	for {
		c := Zgetc(zf)
		if c == -1 {
			return
		}
		if !(c >= 0 && c <= 127 && (c == ' ' || c == '\t' || c == '\n' || c == '\r')) {
			Zungetc(c, zf)
			return
		}
	}
}

func readnum(zf *ZFILE) (rune, error) {
	skipws(zf)
	sign := 1
	c := Zgetc(zf)
	if c == '-' {
		sign = -1
		c = Zgetc(zf)
	}
	if c == -1 {
		return 0, io.EOF
	}

	base := 10
	if c == '0' {
		c2 := Zgetc(zf)
		if c2 == 'x' || c2 == 'X' {
			base = 16
		} else {
			base = 8
			Zungetc(c2, zf)
		}
	} else {
		Zungetc(c, zf)
	}

	acc := 0
	for {
		c := Zgetc(zf)
		if c == -1 {
			break
		}
		digit := -1
		if c >= '0' && c <= '9' {
			digit = c - '0'
		} else if base == 16 {
			if c >= 'a' && c <= 'f' {
				digit = c - 'a' + 10
			} else if c >= 'A' && c <= 'F' {
				digit = c - 'A' + 10
			}
		}
		if digit < 0 || digit >= base {
			Zungetc(c, zf)
			break
		}
		acc = acc*base + digit
	}
	return rune(acc * sign), nil
}

func readTchar(zf *ZFILE) rune {
	thechar := Zgetc(zf)
	if thechar == -1 || thechar == '\n' || thechar == '\r' {
		if thechar != -1 {
			Zungetc(thechar, zf)
		}
		return 0
	}
	if thechar != '\\' {
		return rune(thechar)
	}
	next := Zgetc(zf)
	if next == -1 {
		return '\\'
	}
	switch next {
	case 'a':
		return 7
	case 'b':
		return 8
	case 'e':
		return 27
	case 'f':
		return 12
	case 'n':
		return 10
	case 'r':
		return 13
	case 't':
		return 9
	case 'v':
		return 11
	default:
		if next == '-' || next == 'x' || (next >= '0' && next <= '9') {
			Zungetc(next, zf)
			val, err := readnum(zf)
			if err == nil {
				return val
			}
		}
		return rune(next)
	}
}

func FIGopen(cfg *Config, name string, suffix string) (*ZFILE, error) {
	// Try with fontdirname
	if !hasdirsep(name) {
		path := filepath.Join(cfg.Fontdirname, name+suffix)
		zf, err := Zopen(path, "rb")
		if err == nil {
			return zf, nil
		}
		// Try embedded
		embeddedPath := filepath.Join("fonts", name+suffix)
		zf, err = Zopen(embeddedPath, "rb")
		if err == nil {
			return zf, nil
		}
	}
	// Try as full path
	path := name + suffix
	zf, err := Zopen(path, "rb")
	if err == nil {
		return zf, nil
	}
	// Try embedded
	embeddedPath := filepath.Join("fonts", filepath.Base(name)+suffix)
	return Zopen(embeddedPath, "rb")
}

func charsetname(zf *ZFILE) rune {
	result := readTchar(zf)
	if result == '\n' || result == '\r' {
		Zungetc(int(result), zf)
		return 0
	}
	return result
}

func charset(cfg *Config, n int, controlfile *ZFILE) {
	skipws(controlfile)
	if Zgetc(controlfile) != '9' {
		skiptoeol(controlfile)
		return
	}
	ch := Zgetc(controlfile)
	if ch == '6' {
		cfg.gn[n] = rune(65536)*charsetname(controlfile) + 0x80
		cfg.gndbl[n] = false
		skiptoeol(controlfile)
		return
	}
	if ch != '4' {
		skiptoeol(controlfile)
		return
	}
	ch = Zgetc(controlfile)
	if ch == 'x' {
		if Zgetc(controlfile) != '9' {
			skiptoeol(controlfile)
			return
		}
		if Zgetc(controlfile) != '4' {
			skiptoeol(controlfile)
			return
		}
		skipws(controlfile)
		cfg.gn[n] = rune(65536) * charsetname(controlfile)
		cfg.gndbl[n] = true
		skiptoeol(controlfile)
		return
	}
	Zungetc(ch, controlfile)
	skipws(controlfile)
	cfg.gn[n] = rune(65536) * charsetname(controlfile)
	cfg.gndbl[n] = false
}

func readcontrol(cfg *Config, controlname string) error {
	controlfile, err := FIGopen(cfg, controlname, CONTROLFILESUFFIX)
	if err != nil {
		return fmt.Errorf("unable to open control file: %s", controlname)
	}
	defer Zclose(controlfile)

	// Begin with a freeze command
	node := &ComNode{thecommand: 0}
	*cfg.commandlistend = node
	cfg.commandlistend = &node.next

	for {
		command := Zgetc(controlfile)
		if command == -1 {
			break
		}
		switch command {
		case 't':
			skipws(controlfile)
			firstch := readTchar(controlfile)
			dashcheck := Zgetc(controlfile)
			var lastch rune
			if dashcheck == '-' {
				lastch = readTchar(controlfile)
			} else {
				Zungetc(dashcheck, controlfile)
				lastch = firstch
			}
			skipws(controlfile)
			offset := readTchar(controlfile) - firstch
			skiptoeol(controlfile)
			node := &ComNode{
				thecommand: 1,
				rangelo:    firstch,
				rangehi:    lastch,
				offset:     offset,
			}
			*cfg.commandlistend = node
			cfg.commandlistend = &node.next
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			Zungetc(command, controlfile)
			firstch, _ := readnum(controlfile)
			skipws(controlfile)
			lastch, _ := readnum(controlfile)
			offset := lastch - firstch
			lastch = firstch
			skiptoeol(controlfile)
			node := &ComNode{
				thecommand: 1,
				rangelo:    firstch,
				rangehi:    lastch,
				offset:     offset,
			}
			*cfg.commandlistend = node
			cfg.commandlistend = &node.next
		case 'f':
			skiptoeol(controlfile)
			node := &ComNode{thecommand: 0}
			*cfg.commandlistend = node
			cfg.commandlistend = &node.next
		case 'b':
			cfg.Multibyte = 1
		case 'u':
			cfg.Multibyte = 2
		case 'h':
			cfg.Multibyte = 3
		case 'j':
			cfg.Multibyte = 4
		case 'g':
			cfg.Multibyte = 0
			skipws(controlfile)
			command := Zgetc(controlfile)
			switch command {
			case '0':
				charset(cfg, 0, controlfile)
			case '1':
				charset(cfg, 1, controlfile)
			case '2':
				charset(cfg, 2, controlfile)
			case '3':
				charset(cfg, 3, controlfile)
			case 'l', 'L':
				skipws(controlfile)
				cfg.gl = Zgetc(controlfile) - '0'
				skiptoeol(controlfile)
			case 'r', 'R':
				skipws(controlfile)
				cfg.gr = Zgetc(controlfile) - '0'
				skiptoeol(controlfile)
			default:
				skiptoeol(controlfile)
			}
		case '\r', '\n':
			// blank line
		default:
			skiptoeol(controlfile)
		}
	}
	return nil
}

func readcontrolfiles(cfg *Config) {
	for cfnptr := cfg.cfilelist; cfnptr != nil; cfnptr = cfnptr.next {
		readcontrol(cfg, cfnptr.thename)
	}
}

func (cfg *Config) clearline() {
	for i := 0; i < cfg.charheight; i++ {
		cfg.outputline[i] = cfg.outputline[i][:0]
		if cfg.charPositionMap != nil && i < len(cfg.charPositionMap) {
			cfg.charPositionMap[i] = cfg.charPositionMap[i][:0]
		}
	}
	cfg.outlinelen = 0
	cfg.inchrlinelen = 0
}

func readfontchar(cfg *Config, file *ZFILE, theord rune) {
	fclsave := cfg.fcharlist
	cfg.fcharlist = &FCharNode{
		ord:     theord,
		thechar: make([][]rune, cfg.charheight),
		next:    fclsave,
	}

	templine := make([]byte, MAXLEN+1)
	for row := 0; row < cfg.charheight; row++ {
		line := myfgets(templine, MAXLEN+1, file)
		if line == nil {
			cfg.fcharlist.thechar[row] = []rune{}
			continue
		}
		// Remove newline if present
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		// Also remove \r if present
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		var outline []rune
		if cfg.toiletfont {
			outline = []rune(string(line))
		} else {
			outline = []rune(string(line))
		}
		// Remove trailing spaces
		k := len(outline) - 1
		for k >= 0 && k < len(outline) && unicode.IsSpace(outline[k]) {
			k--
		}
		// Remove endmarks
		if k >= 0 && k < len(outline) {
			endchar := outline[k]
			for k >= 0 && k < len(outline) && outline[k] == endchar {
				k--
			}
		}
		// k+1 is the new length (like outline[k+1] = '\0' in C)
		if k+1 >= 0 {
			if k+1 <= len(outline) {
				outline = outline[:k+1]
			} else {
				outline = []rune{}
			}
		} else {
			outline = []rune{}
		}
		cfg.fcharlist.thechar[row] = outline
	}
}

func readfont(cfg *Config) error {
	fontfile, err := FIGopen(cfg, cfg.Fontname, FONTFILESUFFIX)
	if err != nil {
		fontfile, err = FIGopen(cfg, cfg.Fontname, TOILETFILESUFFIX)
		if err == nil {
			cfg.toiletfont = true
		}
	}
	if err != nil {
		return fmt.Errorf("unable to open font file: %s", cfg.Fontname)
	}
	defer Zclose(fontfile)

	magicnum := readmagic(fontfile)
	fileline := make([]byte, MAXLEN+1)
	headerLine := myfgets(fileline, MAXLEN+1, fontfile)
	if len(headerLine) > 0 && headerLine[len(headerLine)-1] != '\n' {
		skiptoeol(fontfile)
	}

	var hardblank byte
	var charheight, upheight, maxlen, smush, cmtlines, ffright2left, smush2 int
	line := strings.TrimSpace(string(fileline))
	// Format: a$ 6 5 16 15 11 0 24463 229
	// magicnum is "flf2", then line has "a$ 6 5 16 15 11 0 24463 229"
	// %*c skips the 'a', then reads hardblank '$'
	var dummy byte
	numsread, _ := fmt.Sscanf(line, "%c%c %d %d %d %d %d %d %d",
		&dummy, &hardblank, &charheight, &upheight, &maxlen, &smush, &cmtlines,
		&ffright2left, &smush2)

	if maxlen > MAXLEN {
		return fmt.Errorf("font %s: character is too wide", cfg.Fontname)
	}

	// Check magic number
	if (!cfg.toiletfont && magicnum != FONTFILEMAGICNUMBER) ||
		(cfg.toiletfont && magicnum != TOILETFILEMAGICNUMBER) {
		return fmt.Errorf("font %s: not a FIGlet 2 font file (magic: %s, expected: %s)", cfg.Fontname, magicnum, FONTFILEMAGICNUMBER)
	}
	if numsread < 7 {
		return fmt.Errorf("font %s: not a FIGlet 2 font file (numsread: %d)", cfg.Fontname, numsread)
	}

	for i := 1; i <= cmtlines; i++ {
		skiptoeol(fontfile)
	}

	if numsread < 8 {
		ffright2left = 0
	}

	if numsread < 9 {
		if smush == 0 {
			smush2 = SM_KERN
		} else if smush < 0 {
			smush2 = 0
		} else {
			smush2 = (smush & 31) | SM_SMUSH
		}
	}

	if charheight < 1 {
		charheight = 1
	}

	if maxlen < 1 {
		maxlen = 1
	}

	maxlen += 100

	if cfg.Smushoverride == SMO_NO {
		cfg.Smushmode = smush2
	} else if cfg.Smushoverride == SMO_FORCE {
		cfg.Smushmode |= smush2
	}

	if cfg.Right2left < 0 {
		if ffright2left != 0 {
			cfg.Right2left = 1
		} else {
			cfg.Right2left = 0
		}
	}

	if cfg.Justification < 0 {
		cfg.Justification = 2 * cfg.Right2left
	}

	cfg.hardblank = rune(hardblank)
	cfg.charheight = charheight

	// Allocate "missing" character
	cfg.fcharlist = &FCharNode{
		ord:     0,
		thechar: make([][]rune, charheight),
		next:    nil,
	}
	for row := 0; row < charheight; row++ {
		cfg.fcharlist.thechar[row] = []rune{}
	}

	for theord := ' '; theord <= '~'; theord++ {
		readfontchar(cfg, fontfile, theord)
	}
	for i := 0; i <= 6; i++ {
		readfontchar(cfg, fontfile, deutsch[i])
	}

	fileline = make([]byte, maxlen+1)
	for {
		line := myfgets(fileline, maxlen+1, fontfile)
		if line == nil {
			break
		}
		lineStr := strings.TrimSpace(string(line))
		var theord int64
		var err error
		// Try to parse as hex (0x...) or octal (0...) or decimal
		if strings.HasPrefix(lineStr, "0x") || strings.HasPrefix(lineStr, "0X") {
			_, err = fmt.Sscanf(lineStr, "0x%x", &theord)
			if err != nil {
				_, err = fmt.Sscanf(lineStr, "0X%x", &theord)
			}
		} else if strings.HasPrefix(lineStr, "-0x") || strings.HasPrefix(lineStr, "-0X") {
			_, err = fmt.Sscanf(lineStr, "-0x%x", &theord)
			if err != nil {
				_, err = fmt.Sscanf(lineStr, "-0X%x", &theord)
			}
			theord = -theord
		} else {
			theord, err = strconv.ParseInt(lineStr, 0, 64)
			if err != nil {
				// Try just reading first number
				_, err = fmt.Sscanf(lineStr, "%d", &theord)
			}
		}
		if err != nil {
			break
		}
		readfontchar(cfg, fontfile, rune(theord))
	}
	return nil
}

func linealloc(cfg *Config) {
	cfg.outputline = make([][]rune, cfg.charheight)
	for row := 0; row < cfg.charheight; row++ {
		cfg.outputline[row] = make([]rune, cfg.outlinelenlimit+1)
	}
	cfg.inchrlinelenlimit = cfg.Outputwidth*4 + 100
	cfg.inchrline = make([]rune, cfg.inchrlinelenlimit+1)
	cfg.clearline()
}

func (cfg *Config) getletter(c rune) {
	var charptr *FCharNode
	for charptr = cfg.fcharlist; charptr != nil && charptr.ord != c; charptr = charptr.next {
	}
	if charptr != nil {
		cfg.currchar = charptr.thechar
	} else {
		for charptr = cfg.fcharlist; charptr != nil && charptr.ord != 0; charptr = charptr.next {
		}
		cfg.currchar = charptr.thechar
	}
	cfg.previouscharwidth = cfg.currcharwidth
	if len(cfg.currchar) > 0 && len(cfg.currchar[0]) > 0 {
		cfg.currcharwidth = len(cfg.currchar[0])
	} else {
		cfg.currcharwidth = 0
	}
}

func (cfg *Config) smushem(lch, rch rune) rune {
	if lch == ' ' {
		return rch
	}
	if rch == ' ' {
		return lch
	}

	if cfg.previouscharwidth < 2 || cfg.currcharwidth < 2 {
		return 0
	}

	if (cfg.Smushmode & SM_SMUSH) == 0 {
		return 0
	}

	if (cfg.Smushmode & 63) == 0 {
		if lch == ' ' {
			return rch
		}
		if rch == ' ' {
			return lch
		}
		if lch == cfg.hardblank {
			return rch
		}
		if rch == cfg.hardblank {
			return lch
		}
		if cfg.Right2left == 1 {
			return lch
		}
		return rch
	}

	if (cfg.Smushmode & SM_HARDBLANK) != 0 {
		if lch == cfg.hardblank && rch == cfg.hardblank {
			return lch
		}
	}

	if lch == cfg.hardblank || rch == cfg.hardblank {
		return 0
	}

	if (cfg.Smushmode & SM_EQUAL) != 0 {
		if lch == rch {
			return lch
		}
	}

	if (cfg.Smushmode & SM_LOWLINE) != 0 {
		if lch == '_' && strings.ContainsRune("|/\\[]{}()<>", rch) {
			return rch
		}
		if rch == '_' && strings.ContainsRune("|/\\[]{}()<>", lch) {
			return lch
		}
	}

	if (cfg.Smushmode & SM_HIERARCHY) != 0 {
		if lch == '|' && strings.ContainsRune("/\\[]{}()<>", rch) {
			return rch
		}
		if rch == '|' && strings.ContainsRune("/\\[]{}()<>", lch) {
			return lch
		}
		if strings.ContainsRune("/\\", lch) && strings.ContainsRune("[]{}()<>", rch) {
			return rch
		}
		if strings.ContainsRune("/\\", rch) && strings.ContainsRune("[]{}()<>", lch) {
			return lch
		}
		if strings.ContainsRune("[]", lch) && strings.ContainsRune("{}()<>", rch) {
			return rch
		}
		if strings.ContainsRune("[]", rch) && strings.ContainsRune("{}()<>", lch) {
			return lch
		}
		if strings.ContainsRune("{}", lch) && strings.ContainsRune("()<>", rch) {
			return rch
		}
		if strings.ContainsRune("{}", rch) && strings.ContainsRune("()<>", lch) {
			return lch
		}
		if strings.ContainsRune("()", lch) && strings.ContainsRune("<>", rch) {
			return rch
		}
		if strings.ContainsRune("()", rch) && strings.ContainsRune("<>", lch) {
			return lch
		}
	}

	if (cfg.Smushmode & SM_PAIR) != 0 {
		if lch == '[' && rch == ']' {
			return '|'
		}
		if rch == '[' && lch == ']' {
			return '|'
		}
		if lch == '{' && rch == '}' {
			return '|'
		}
		if rch == '{' && lch == '}' {
			return '|'
		}
		if lch == '(' && rch == ')' {
			return '|'
		}
		if rch == '(' && lch == ')' {
			return '|'
		}
	}

	if (cfg.Smushmode & SM_BIGX) != 0 {
		if lch == '/' && rch == '\\' {
			return '|'
		}
		if rch == '/' && lch == '\\' {
			return 'Y'
		}
		if lch == '>' && rch == '<' {
			return 'X'
		}
	}

	return 0
}

func (cfg *Config) smushamt() int {
	if (cfg.Smushmode & (SM_SMUSH | SM_KERN)) == 0 {
		return 0
	}
	maxsmush := cfg.currcharwidth
	for row := 0; row < cfg.charheight; row++ {
		var linebd, charbd int
		var ch1, ch2 rune

		if cfg.Right2left == 1 {
			// C: for (charbd=STRLEN(currchar[row]);
			//      ch1=currchar[row][charbd],(charbd>0&&(!ch1||ch1==' '));charbd--) ;
			charbd = len(cfg.currchar[row])
			for {
				// Get ch1 at current position (null terminator if out of bounds)
				if charbd < len(cfg.currchar[row]) {
					ch1 = cfg.currchar[row][charbd]
				} else {
					ch1 = 0
				}
				// Check condition
				if !(charbd > 0 && (ch1 == 0 || ch1 == ' ')) {
					break
				}
				charbd--
			}

			// C: for (linebd=0;ch2=outputline[row][linebd],ch2==' ';linebd++) ;
			linebd = 0
			for {
				if linebd < len(cfg.outputline[row]) {
					ch2 = cfg.outputline[row][linebd]
				} else {
					ch2 = 0
				}
				if ch2 != ' ' {
					break
				}
				linebd++
			}
			amt := linebd + cfg.currcharwidth - 1 - charbd

			// C: if (!ch1||ch1==' ') { amt++; }
			if ch1 == 0 || ch1 == ' ' {
				amt++
			} else if ch2 != 0 {
				if cfg.smushem(ch1, ch2) != 0 {
					amt++
				}
			}

			if amt < maxsmush {
				maxsmush = amt
			}
		} else {
			// C: for (linebd=STRLEN(outputline[row]);
			//      ch1 = outputline[row][linebd],(linebd>0&&(!ch1||ch1==' '));linebd--) ;
			linebd = len(cfg.outputline[row])
			for {
				// Get ch1 at current position (null terminator if out of bounds)
				if linebd < len(cfg.outputline[row]) {
					ch1 = cfg.outputline[row][linebd]
				} else {
					ch1 = 0
				}
				// Check condition
				if !(linebd > 0 && (ch1 == 0 || ch1 == ' ')) {
					break
				}
				linebd--
			}

			// C: for (charbd=0;ch2=currchar[row][charbd],ch2==' ';charbd++) ;
			charbd = 0
			for {
				if charbd < len(cfg.currchar[row]) {
					ch2 = cfg.currchar[row][charbd]
				} else {
					ch2 = 0
				}
				if ch2 != ' ' {
					break
				}
				charbd++
			}
			amt := charbd + cfg.outlinelen - 1 - linebd

			// C: if (!ch1||ch1==' ') { amt++; }
			if ch1 == 0 || ch1 == ' ' {
				amt++
			} else if ch2 != 0 {
				if cfg.smushem(ch1, ch2) != 0 {
					amt++
				}
			}

			if amt < maxsmush {
				maxsmush = amt
			}
		}
	}
	return maxsmush
}

func (cfg *Config) addchar(c rune) bool {
	cfg.getletter(c)
	smushamount := cfg.smushamt()
	if smushamount < 0 {
		smushamount = 0
	}
	if smushamount > cfg.currcharwidth {
		smushamount = cfg.currcharwidth
	}
	if cfg.outlinelen+cfg.currcharwidth-smushamount > cfg.outlinelenlimit ||
		cfg.inchrlinelen+1 > cfg.inchrlinelenlimit {
		return false
	}

	// Track character position for color mapping (only for non-space characters)
	trackChar := c != ' ' && c != '\n' && c != '\t'
	if trackChar {
		cfg.currentCharIndex++
	}

	for row := 0; row < cfg.charheight; row++ {
		if cfg.Right2left == 1 {
			templine := make([]rune, len(cfg.currchar[row]))
			copy(templine, cfg.currchar[row])
			for k := 0; k < smushamount && k < len(cfg.outputline[row]); k++ {
				idx := cfg.currcharwidth - smushamount + k
				if idx >= 0 && idx < len(templine) {
					smushed := cfg.smushem(templine[idx], cfg.outputline[row][k])
					if smushed != 0 {
						templine[idx] = smushed
					}
				}
			}
			remaining := len(cfg.outputline[row])
			if smushamount < remaining {
				cfg.outputline[row] = append(templine, cfg.outputline[row][smushamount:]...)
				// Track character positions for Right2left
				if trackChar && row < len(cfg.charPositionMap) {
					charWidth := len(templine)
					// Insert at the beginning for Right2left
					newMap := make([]int, charWidth)
					charIdx := cfg.currentCharIndex - 1
					for i := range newMap {
						newMap[i] = charIdx
					}
					// Only slice if we have enough elements
					if smushamount < len(cfg.charPositionMap[row]) {
						cfg.charPositionMap[row] = append(newMap, cfg.charPositionMap[row][smushamount:]...)
					} else {
						cfg.charPositionMap[row] = newMap
					}
				}
			} else {
				cfg.outputline[row] = templine
				// Track character positions for Right2left
				if trackChar && row < len(cfg.charPositionMap) {
					charWidth := len(templine)
					newMap := make([]int, charWidth)
					charIdx := cfg.currentCharIndex - 1
					for i := range newMap {
						newMap[i] = charIdx
					}
					cfg.charPositionMap[row] = newMap
				}
			}
		} else {
			// Track character positions for color mapping
			startCol := cfg.outlinelen - smushamount
			if startCol < 0 {
				startCol = 0
			}

			for k := 0; k < smushamount; k++ {
				column := cfg.outlinelen - smushamount + k
				if column < 0 {
					column = 0
				}
				if column < len(cfg.outputline[row]) && k < len(cfg.currchar[row]) {
					cfg.outputline[row][column] = cfg.smushem(cfg.outputline[row][column], cfg.currchar[row][k])
					// Update character position map for smushed positions
					if trackChar && row < len(cfg.charPositionMap) && column < len(cfg.charPositionMap[row]) {
						// Keep the existing character index for smushed positions
					}
				}
			}
			if smushamount < len(cfg.currchar[row]) {
				cfg.outputline[row] = append(cfg.outputline[row], cfg.currchar[row][smushamount:]...)
				// Track character positions for new columns
				if trackChar && row < len(cfg.charPositionMap) {
					charWidth := len(cfg.currchar[row]) - smushamount
					for i := 0; i < charWidth; i++ {
						cfg.charPositionMap[row] = append(cfg.charPositionMap[row], cfg.currentCharIndex-1)
					}
				}
			}
		}
	}
	if len(cfg.outputline[0]) > 0 {
		cfg.outlinelen = len(cfg.outputline[0])
	}
	cfg.inchrline[cfg.inchrlinelen] = c
	cfg.inchrlinelen++
	return true
}

func (cfg *Config) putstring(str []rune) {
	length := len(str)
	if cfg.Outputwidth > 1 {
		if length > cfg.Outputwidth-1 {
			length = cfg.Outputwidth - 1
		}
		if cfg.Justification > 0 {
			for i := 1; (3-cfg.Justification)*i+length+cfg.Justification-2 < cfg.Outputwidth; i++ {
				cfg.output.WriteString(" ")
			}
		}
	}

	// Apply colors if enabled
	hasColors := len(cfg.Colors) > 0 && cfg.OutputParser != nil && cfg.OutputParser.Name != "terminal"

	for i := 0; i < length; i++ {
		if i < len(str) {
			var charStr string
			if str[i] == cfg.hardblank {
				charStr = " "
			} else {
				charStr = string(str[i])
			}

			// Apply color if enabled
			if hasColors {
				charStr = cfg.applyColorToChar(charStr, i)
			} else {
				// Apply parser replacements even without colors
				if cfg.OutputParser != nil {
					charStr = handleReplaces(charStr, cfg.OutputParser)
				}
			}

			cfg.output.WriteString(charStr)
		}
	}

	// Use parser's newline representation
	newline := "\n"
	if cfg.OutputParser != nil && cfg.OutputParser.NewLine != "" {
		newline = cfg.OutputParser.NewLine
	}
	cfg.output.WriteString(newline)

	// Move to next line for character position tracking
	cfg.currentLineIndex++
	if cfg.currentLineIndex >= cfg.charheight {
		cfg.currentLineIndex = 0
	}
}

// applyColorToChar applies color to a character based on its position in the line
func (cfg *Config) applyColorToChar(charStr string, position int) string {
	if len(cfg.Colors) == 0 {
		return handleReplaces(charStr, cfg.OutputParser)
	}

	// Get the input character index for this position
	charIndex := -1
	if cfg.charPositionMap != nil && cfg.currentLineIndex < len(cfg.charPositionMap) {
		if position < len(cfg.charPositionMap[cfg.currentLineIndex]) {
			charIndex = cfg.charPositionMap[cfg.currentLineIndex][position]
		}
	}

	// If we couldn't map to an input character, use position-based cycling
	if charIndex < 0 {
		charIndex = position
	}

	// Cycle through colors based on character index
	colorIndex := charIndex % len(cfg.Colors)
	if colorIndex < 0 {
		colorIndex = 0
	}
	color := cfg.Colors[colorIndex]

	prefix := color.getPrefix(cfg.OutputParser)
	suffix := color.getSuffix(cfg.OutputParser)

	// Apply parser replacements
	replaced := handleReplaces(charStr, cfg.OutputParser)

	return prefix + replaced + suffix
}

func (cfg *Config) printline() {
	cfg.currentLineIndex = 0
	for i := 0; i < cfg.charheight; i++ {
		cfg.putstring(cfg.outputline[i])
	}
	cfg.clearline()
}

func (cfg *Config) splitline() {
	part1 := make([]rune, cfg.inchrlinelen+1)
	part2 := make([]rune, cfg.inchrlinelen+1)
	gotspace := false
	lastspace := cfg.inchrlinelen - 1
	i := cfg.inchrlinelen - 1
	for i >= 0 {
		if !gotspace && cfg.inchrline[i] == ' ' {
			gotspace = true
			lastspace = i
		}
		if gotspace && cfg.inchrline[i] != ' ' {
			break
		}
		i--
	}
	len1 := i + 1
	len2 := cfg.inchrlinelen - lastspace - 1
	for i := 0; i < len1; i++ {
		part1[i] = cfg.inchrline[i]
	}
	for i := 0; i < len2; i++ {
		part2[i] = cfg.inchrline[lastspace+1+i]
	}
	cfg.clearline()
	for i := 0; i < len1; i++ {
		cfg.addchar(part1[i])
	}
	cfg.printline()
	for i := 0; i < len2; i++ {
		cfg.addchar(part2[i])
	}
}

func handlemapping(cfg *Config, c rune) rune {
	if cfg.commandlist == nil {
		return c
	}
	for cmptr := cfg.commandlist; cmptr != nil; {
		if cmptr.thecommand != 0 {
			if c >= cmptr.rangelo && c <= cmptr.rangehi {
				c += cmptr.offset
				for cmptr != nil && cmptr.thecommand != 0 {
					cmptr = cmptr.next
				}
			} else {
				cmptr = cmptr.next
			}
		} else {
			cmptr = cmptr.next
		}
	}
	return c
}

func ungetinchr(cfg *Config, c rune) {
	cfg.getinchr_buffer = c
	cfg.getinchr_flag = true
}

func Agetchar(cfg *Config) int {
	if !cfg.Cmdinput {
		var b [1]byte
		n, _ := os.Stdin.Read(b[:])
		if n == 0 {
			return -1
		}
		return int(b[0])
	}

	if cfg.getinchr_flag {
		cfg.getinchr_flag = false
		return int(cfg.getinchr_buffer)
	}

	// EOF is sticky: ensure it now and forever more
	if cfg.agetmode < 0 || cfg.Optind >= len(cfg.Argv) {
		return -1
	}

	// find next character
	arg := cfg.Argv[cfg.Optind]
	var c int
	if cfg.agetmode < len(arg) {
		c = int(arg[cfg.agetmode]) & 0xFF
	} else {
		c = 0 // reached end of string (null terminator)
	}
	cfg.agetmode++

	if c == 0 {
		// at end of word: return ' ' if normal word, '\n' if empty
		c = ' '                // suppose normal word and return blank
		if cfg.agetmode == 1 { // if ran out in very 1st char, force \n
			c = '\n' // (allows "hello '' world" to do \n at '')
		}
		cfg.agetmode = 0                 // return to char 0 in NEXT word
		cfg.Optind++                     // run up word count
		if cfg.Optind >= len(cfg.Argv) { // check if at "EOF"
			// just ran out of arguments
			c = -1            // return EOF
			cfg.agetmode = -1 // ensure all future returns return EOF
		}
	}

	return c
}

func iso2022(cfg *Config) rune {
	ch := rune(Agetchar(cfg))
	if ch == -1 {
		return ch
	}
	if ch == 27 {
		ch = rune(Agetchar(cfg)) + 0x100
	}
	if ch == 0x100+'$' {
		ch = rune(Agetchar(cfg)) + 0x200
	}
	switch ch {
	case 14:
		cfg.gl = 1
		return iso2022(cfg)
	case 15:
		cfg.gl = 0
		return iso2022(cfg)
	case 142, 'N' + 0x100:
		save_gl := cfg.gl
		save_gr := cfg.gr
		cfg.gl = 2
		cfg.gr = 2
		ch = iso2022(cfg)
		cfg.gl = save_gl
		cfg.gr = save_gr
		return ch
	case 143, 'O' + 0x100:
		save_gl := cfg.gl
		save_gr := cfg.gr
		cfg.gl = 3
		cfg.gr = 3
		ch = iso2022(cfg)
		cfg.gl = save_gl
		cfg.gr = save_gr
		return ch
	case 'n' + 0x100:
		cfg.gl = 2
		return iso2022(cfg)
	case 'o' + 0x100:
		cfg.gl = 3
		return iso2022(cfg)
	case '~' + 0x100:
		cfg.gr = 1
		return iso2022(cfg)
	case '}' + 0x100:
		cfg.gr = 2
		return iso2022(cfg)
	case '|' + 0x100:
		cfg.gr = 3
		return iso2022(cfg)
	case '(' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'B' {
			ch = 0
		}
		cfg.gn[0] = ch << 16
		cfg.gndbl[0] = false
		return iso2022(cfg)
	case ')' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'B' {
			ch = 0
		}
		cfg.gn[1] = ch << 16
		cfg.gndbl[1] = false
		return iso2022(cfg)
	case '*' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'B' {
			ch = 0
		}
		cfg.gn[2] = ch << 16
		cfg.gndbl[2] = false
		return iso2022(cfg)
	case '+' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'B' {
			ch = 0
		}
		cfg.gn[3] = ch << 16
		cfg.gndbl[3] = false
		return iso2022(cfg)
	case '-' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'A' {
			ch = 0
		}
		cfg.gn[1] = (ch << 16) | 0x80
		cfg.gndbl[1] = false
		return iso2022(cfg)
	case '.' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'A' {
			ch = 0
		}
		cfg.gn[2] = (ch << 16) | 0x80
		cfg.gndbl[2] = false
		return iso2022(cfg)
	case '/' + 0x100:
		ch = rune(Agetchar(cfg))
		if ch == 'A' {
			ch = 0
		}
		cfg.gn[3] = (ch << 16) | 0x80
		cfg.gndbl[3] = false
		return iso2022(cfg)
	case '(' + 0x200:
		ch = rune(Agetchar(cfg))
		cfg.gn[0] = ch << 16
		cfg.gndbl[0] = true
		return iso2022(cfg)
	case ')' + 0x200:
		ch = rune(Agetchar(cfg))
		cfg.gn[1] = ch << 16
		cfg.gndbl[1] = true
		return iso2022(cfg)
	case '*' + 0x200:
		ch = rune(Agetchar(cfg))
		cfg.gn[2] = ch << 16
		cfg.gndbl[2] = true
		return iso2022(cfg)
	case '+' + 0x200:
		ch = rune(Agetchar(cfg))
		cfg.gn[3] = ch << 16
		cfg.gndbl[3] = true
		return iso2022(cfg)
	}

	if ch >= 0x21 && ch <= 0x7E {
		if cfg.gndbl[cfg.gl] {
			ch2 := rune(Agetchar(cfg))
			return cfg.gn[cfg.gl] | (ch << 8) | ch2
		}
		return cfg.gn[cfg.gl] | ch
	} else if ch >= 0xA0 && ch <= 0xFF {
		if cfg.gndbl[cfg.gr] {
			ch2 := rune(Agetchar(cfg))
			return cfg.gn[cfg.gr] | (ch << 8) | ch2
		}
		return cfg.gn[cfg.gr] | (ch &^ 0x80)
	}
	return ch
}

func getinchr(cfg *Config) rune {
	if cfg.getinchr_flag {
		cfg.getinchr_flag = false
		return cfg.getinchr_buffer
	}

	switch cfg.Multibyte {
	case 0:
		return iso2022(cfg)
	case 1:
		ch := Agetchar(cfg)
		if (ch >= 0x80 && ch <= 0x9F) || (ch >= 0xE0 && ch <= 0xEF) {
			ch = (ch << 8) + Agetchar(cfg)
		}
		return rune(ch)
	case 2:
		ch := Agetchar(cfg)
		if ch < 0x80 {
			return rune(ch)
		}
		if ch < 0xC0 || ch > 0xFD {
			return 0x0080
		}
		ch2 := Agetchar(cfg) & 0x3F
		if ch < 0xE0 {
			return rune(((ch & 0x1F) << 6) + ch2)
		}
		ch3 := Agetchar(cfg) & 0x3F
		if ch < 0xF0 {
			return rune(((ch & 0x0F) << 12) + (ch2 << 6) + ch3)
		}
		ch4 := Agetchar(cfg) & 0x3F
		if ch < 0xF8 {
			return rune(((ch & 0x07) << 18) + (ch2 << 12) + (ch3 << 6) + ch4)
		}
		ch5 := Agetchar(cfg) & 0x3F
		if ch < 0xFC {
			return rune(((ch & 0x03) << 24) + (ch2 << 18) + (ch3 << 12) + (ch4 << 6) + ch5)
		}
		ch6 := Agetchar(cfg) & 0x3F
		return rune(((ch & 0x01) << 30) + (ch2 << 24) + (ch3 << 18) + (ch4 << 12) + (ch5 << 6) + ch6)
	case 3:
		ch := Agetchar(cfg)
		if ch == -1 {
			return -1
		}
		if cfg.hzmode {
			ch = (ch << 8) + Agetchar(cfg)
			if ch == (int('}')<<8)+int('~') {
				cfg.hzmode = false
				return getinchr(cfg)
			}
			return rune(ch)
		} else if ch == '~' {
			ch2 := Agetchar(cfg)
			if ch2 == '{' {
				cfg.hzmode = true
				return getinchr(cfg)
			} else if ch2 == '~' {
				return rune(ch)
			} else {
				return getinchr(cfg)
			}
		}
		return rune(ch)
	case 4:
		ch := Agetchar(cfg)
		if (ch >= 0x80 && ch <= 0x9F) || (ch >= 0xE0 && ch <= 0xEF) {
			ch = (ch << 8) + Agetchar(cfg)
		}
		return rune(ch)
	default:
		return 0x80
	}
}

// AddControlFile adds a control file to the configuration
func (cfg *Config) AddControlFile(name string) {
	controlname := name
	if suffixcmp(controlname, CONTROLFILESUFFIX) {
		controlname = controlname[:len(controlname)-len(CONTROLFILESUFFIX)]
	}
	node := &CFNameNode{thename: controlname}
	*cfg.cfilelistend = node
	cfg.cfilelistend = &node.next
}

// ClearControlFiles clears all control files
func (cfg *Config) ClearControlFiles() {
	cfg.clearcfilelist()
	cfg.Multibyte = 0
	cfg.gn[0] = 0
	cfg.gn[1] = 0x80
	cfg.gn[2] = 0
	cfg.gn[3] = 0
	cfg.gndbl[0] = false
	cfg.gndbl[1] = false
	cfg.gndbl[2] = false
	cfg.gndbl[3] = false
	cfg.gl = 0
	cfg.gr = 1
}
