package main

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
	"syscall"
	"unicode"
	"unsafe"
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

type FCharNode struct {
	ord     rune
	thechar [][]rune
	next    *FCharNode
}

type CFNameNode struct {
	thename string
	next    *CFNameNode
}

type ComNode struct {
	thecommand int
	rangelo    rune
	rangehi    rune
	offset     rune
	next       *ComNode
}

type Config struct {
	deutschflag       bool
	justification     int // -1 = auto, 0 = left, 1 = center, 2 = right
	paragraphflag     bool
	right2left        int // -1 = auto, 0 = left, 1 = right
	multibyte         int // 0 = ISO 2022, 1 = DBCS, 2 = UTF-8, 3 = HZ, 4 = Shift-JIS
	cmdinput          bool
	smushmode         int
	smushoverride     int
	outputwidth       int
	fontdirname       string
	fontname          string
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
	optind            int
	argv              []string
	agetmode          int // >= 0 for displacement into argv[n], <0 EOF
}

func main() {
	cfg := &Config{
		justification: -1,
		right2left:    -1,
		outputwidth:   DEFAULTCOLUMNS,
		gr:            1,
		gn:            [4]rune{0, 0x80, 0, 0},
		argv:          os.Args,
	}
	cfg.cfilelistend = &cfg.cfilelist
	cfg.commandlistend = &cfg.commandlist

	getparams(cfg)
	readcontrolfiles(cfg)
	readfont(cfg)
	linealloc(cfg)

	wordbreakmode := 0
	last_was_eol_flag := false

	for {
		c := getinchr(cfg)
		if c == -1 { // EOF
			break
		}

		if c == '\n' && cfg.paragraphflag && !last_was_eol_flag {
			c2 := getinchr(cfg)
			ungetinchr(cfg, c2)
			if isASCII(c2) && unicode.IsSpace(c2) {
				c = '\n'
			} else {
				c = ' '
			}
		}
		last_was_eol_flag = isASCII(c) && unicode.IsSpace(c) && c != '\t' && c != ' '

		if cfg.deutschflag {
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
				printline(cfg)
				wordbreakmode = 0
			} else if addchar(cfg, c) {
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
					if cfg.right2left == 1 && cfg.outputwidth > 1 {
						start := len(cfg.currchar[i]) - cfg.outlinelenlimit
						if start < 0 {
							start = 0
						}
						putstring(cfg, cfg.currchar[i][start:])
					} else {
						putstring(cfg, cfg.currchar[i])
					}
				}
				wordbreakmode = -1
			} else if c == ' ' {
				if wordbreakmode == 2 {
					splitline(cfg)
				} else {
					printline(cfg)
				}
				wordbreakmode = -1
			} else {
				if wordbreakmode >= 2 {
					splitline(cfg)
				} else {
					printline(cfg)
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
		printline(cfg)
	}
}

func isASCII(r rune) bool {
	return r >= 0 && r <= 127
}

func getmyname(argv []string) string {
	if len(argv) == 0 {
		return "figlet"
	}
	name := filepath.Base(argv[0])
	return name
}

func printusage(cfg *Config, out io.Writer) {
	myname := getmyname(cfg.argv)
	fmt.Fprintf(out, "Usage: %s [ -cklnoprstvxDELNRSWX ] [ -d fontdirectory ]\n", myname)
	fmt.Fprintf(out, "              [ -f fontfile ] [ -m smushmode ] [ -w outputwidth ]\n")
	fmt.Fprintf(out, "              [ -C controlfile ] [ -I infocode ] [ message ]\n")
}

func printinfo(cfg *Config, infonum int) {
	switch infonum {
	case 0:
		fmt.Printf("FIGlet Copyright (C) 1991-2012 Glenn Chappell, Ian Chai, ")
		fmt.Printf("John Cowan,\nChristiaan Keet and Claudio Matsuoka\n")
		fmt.Printf("Internet: <info@figlet.org> ")
		fmt.Printf("Version: %s, date: %s\n\n", VERSION, DATE)
		fmt.Printf("FIGlet, along with the various FIGlet fonts")
		fmt.Printf(" and documentation, may be\n")
		fmt.Printf("freely copied and distributed.\n\n")
		fmt.Printf("If you use FIGlet, please send an")
		fmt.Printf(" e-mail message to <info@figlet.org>.\n\n")
		fmt.Printf("The latest version of FIGlet is available from the")
		fmt.Printf(" web site,\n\thttp://www.figlet.org/\n\n")
		printusage(cfg, os.Stdout)
	case 1:
		fmt.Printf("%d\n", VERSION_INT)
	case 2:
		fmt.Printf("%s\n", cfg.fontdirname)
	case 3:
		fmt.Printf("%s\n", cfg.fontname)
	case 4:
		fmt.Printf("%d\n", cfg.outputwidth)
	case 5:
		fmt.Printf("%s", FONTFILEMAGICNUMBER)
		fmt.Printf(" %s", TOILETFILEMAGICNUMBER)
		fmt.Printf("\n")
	}
}

func hasdirsep(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, "\\")
}

func suffixcmp(s1, s2 string) bool {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)
	return strings.HasSuffix(s1, s2)
}

func get_columns() int {
	fd, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return -1
	}
	defer fd.Close()

	var ws struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return -1
	}
	return int(ws.Col)
}

func getparams(cfg *Config) {
	myname := getmyname(cfg.argv)
	cfg.fontdirname = "fonts"
	if env := os.Getenv("FIGLET_FONTDIR"); env != "" {
		cfg.fontdirname = env
	}
	cfg.fontname = "standard"
	cfg.smushoverride = SMO_NO
	cfg.deutschflag = false
	cfg.justification = -1
	cfg.right2left = -1
	cfg.paragraphflag = false
	infoprint := -1
	cfg.cmdinput = false
	cfg.outputwidth = DEFAULTCOLUMNS
	cfg.gn[1] = 0x80
	cfg.gr = 1

	// Simple getopt implementation
	optind := 1
	for optind < len(cfg.argv) {
		arg := cfg.argv[optind]
		if len(arg) == 0 || arg[0] != '-' {
			cfg.cmdinput = true
			cfg.optind = optind
			break
		}
		if arg == "--" {
			optind++
			cfg.cmdinput = true
			cfg.optind = optind
			break
		}

		for i := 1; i < len(arg); i++ {
			c := arg[i]
			switch c {
			case 'A':
				cfg.cmdinput = true
			case 'D':
				cfg.deutschflag = true
			case 'E':
				cfg.deutschflag = false
			case 'X':
				cfg.right2left = -1
			case 'L':
				cfg.right2left = 0
			case 'R':
				cfg.right2left = 1
			case 'x':
				cfg.justification = -1
			case 'l':
				cfg.justification = 0
			case 'c':
				cfg.justification = 1
			case 'r':
				cfg.justification = 2
			case 'p':
				cfg.paragraphflag = true
			case 'n':
				cfg.paragraphflag = false
			case 's':
				cfg.smushoverride = SMO_NO
			case 'k':
				cfg.smushmode = SM_KERN
				cfg.smushoverride = SMO_YES
			case 'S':
				cfg.smushmode = SM_SMUSH
				cfg.smushoverride = SMO_FORCE
			case 'o':
				cfg.smushmode = SM_SMUSH
				cfg.smushoverride = SMO_YES
			case 'W':
				cfg.smushmode = 0
				cfg.smushoverride = SMO_YES
			case 't':
				columns := get_columns()
				if columns > 0 {
					cfg.outputwidth = columns
				}
			case 'v':
				infoprint = 0
			case 'I':
				if i+1 < len(arg) {
					val, _ := strconv.Atoi(arg[i+1:])
					infoprint = val
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					val, _ := strconv.Atoi(cfg.argv[optind+1])
					infoprint = val
					optind++
				}
			case 'm':
				var val int
				if i+1 < len(arg) {
					val, _ = strconv.Atoi(arg[i+1:])
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					val, _ = strconv.Atoi(cfg.argv[optind+1])
					optind++
				}
				if val < -1 {
					cfg.smushoverride = SMO_NO
					break
				}
				if val == 0 {
					cfg.smushmode = SM_KERN
				} else if val == -1 {
					cfg.smushmode = 0
				} else {
					cfg.smushmode = (val & 63) | SM_SMUSH
				}
				cfg.smushoverride = SMO_YES
			case 'w':
				var val int
				if i+1 < len(arg) {
					val, _ = strconv.Atoi(arg[i+1:])
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					val, _ = strconv.Atoi(cfg.argv[optind+1])
					optind++
				}
				if val > 0 {
					cfg.outputwidth = val
				}
			case 'd':
				if i+1 < len(arg) {
					cfg.fontdirname = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					cfg.fontdirname = cfg.argv[optind+1]
					optind++
				}
			case 'f':
				var name string
				if i+1 < len(arg) {
					name = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					name = cfg.argv[optind+1]
					optind++
				}
				cfg.fontname = name
				if suffixcmp(cfg.fontname, FONTFILESUFFIX) {
					cfg.fontname = cfg.fontname[:len(cfg.fontname)-len(FONTFILESUFFIX)]
				} else if suffixcmp(cfg.fontname, TOILETFILESUFFIX) {
					cfg.fontname = cfg.fontname[:len(cfg.fontname)-len(TOILETFILESUFFIX)]
				}
			case 'C':
				var name string
				if i+1 < len(arg) {
					name = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.argv) {
					name = cfg.argv[optind+1]
					optind++
				}
				controlname := name
				if suffixcmp(controlname, CONTROLFILESUFFIX) {
					controlname = controlname[:len(controlname)-len(CONTROLFILESUFFIX)]
				}
				node := &CFNameNode{thename: controlname}
				*cfg.cfilelistend = node
				cfg.cfilelistend = &node.next
			case 'N':
				clearcfilelist(cfg)
				cfg.multibyte = 0
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
			case 'F':
				fmt.Fprintf(os.Stderr, "%s: illegal option -- F\n", myname)
				printusage(cfg, os.Stderr)
				fmt.Fprintf(os.Stderr, "\nBecause of numerous incompatibilities, the")
				fmt.Fprintf(os.Stderr, " \"-F\" option has been\n")
				fmt.Fprintf(os.Stderr, "removed.  It has been replaced by the \"figlist\"")
				fmt.Fprintf(os.Stderr, " program, which is now\n")
				fmt.Fprintf(os.Stderr, "included in the basic FIGlet package.  \"figlist\"")
				fmt.Fprintf(os.Stderr, " is also available\n")
				fmt.Fprintf(os.Stderr, "from  http://www.figlet.org/")
				fmt.Fprintf(os.Stderr, "under UNIX utilities.\n")
				os.Exit(1)
			default:
				printusage(cfg, os.Stderr)
				os.Exit(1)
			}
		}
		optind++
	}

	if optind < len(cfg.argv) {
		cfg.cmdinput = true
		cfg.optind = optind
	}

	cfg.outlinelenlimit = cfg.outputwidth - 1
	if infoprint >= 0 {
		printinfo(cfg, infoprint)
		os.Exit(0)
	}
}

func clearcfilelist(cfg *Config) {
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
		path := filepath.Join(cfg.fontdirname, name+suffix)
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

func readcontrol(cfg *Config, controlname string) {
	controlfile, err := FIGopen(cfg, controlname, CONTROLFILESUFFIX)
	if err != nil {
		myname := getmyname(cfg.argv)
		fmt.Fprintf(os.Stderr, "%s: %s: Unable to open control file\n", myname, controlname)
		os.Exit(1)
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
			cfg.multibyte = 1
		case 'u':
			cfg.multibyte = 2
		case 'h':
			cfg.multibyte = 3
		case 'j':
			cfg.multibyte = 4
		case 'g':
			cfg.multibyte = 0
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
}

func readcontrolfiles(cfg *Config) {
	for cfnptr := cfg.cfilelist; cfnptr != nil; cfnptr = cfnptr.next {
		readcontrol(cfg, cfnptr.thename)
	}
}

func clearline(cfg *Config) {
	for i := 0; i < cfg.charheight; i++ {
		cfg.outputline[i] = cfg.outputline[i][:0]
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

func readfont(cfg *Config) {
	fontfile, err := FIGopen(cfg, cfg.fontname, FONTFILESUFFIX)
	if err != nil {
		fontfile, err = FIGopen(cfg, cfg.fontname, TOILETFILESUFFIX)
		if err == nil {
			cfg.toiletfont = true
		}
	}
	if err != nil {
		myname := getmyname(cfg.argv)
		fmt.Fprintf(os.Stderr, "%s: %s: Unable to open font file\n", myname, cfg.fontname)
		os.Exit(1)
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
		myname := getmyname(cfg.argv)
		fmt.Fprintf(os.Stderr, "%s: %s: character is too wide\n", myname, cfg.fontname)
		os.Exit(1)
	}

	// Check magic number
	if (!cfg.toiletfont && magicnum != FONTFILEMAGICNUMBER) ||
		(cfg.toiletfont && magicnum != TOILETFILEMAGICNUMBER) {
		myname := getmyname(cfg.argv)
		fmt.Fprintf(os.Stderr, "%s: %s: Not a FIGlet 2 font file (magic: %s, expected: %s)\n", myname, cfg.fontname, magicnum, FONTFILEMAGICNUMBER)
		os.Exit(1)
	}
	if numsread < 5 {
		myname := getmyname(cfg.argv)
		fmt.Fprintf(os.Stderr, "%s: %s: Not a FIGlet 2 font file (numsread: %d)\n", myname, cfg.fontname, numsread)
		os.Exit(1)
	}

	for i := 1; i <= cmtlines; i++ {
		skiptoeol(fontfile)
	}

	if numsread < 6 {
		ffright2left = 0
	}

	if numsread < 7 {
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

	if cfg.smushoverride == SMO_NO {
		cfg.smushmode = smush2
	} else if cfg.smushoverride == SMO_FORCE {
		cfg.smushmode |= smush2
	}

	if cfg.right2left < 0 {
		if ffright2left != 0 {
			cfg.right2left = 1
		} else {
			cfg.right2left = 0
		}
	}

	if cfg.justification < 0 {
		cfg.justification = 2 * cfg.right2left
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
}

func linealloc(cfg *Config) {
	cfg.outputline = make([][]rune, cfg.charheight)
	for row := 0; row < cfg.charheight; row++ {
		cfg.outputline[row] = make([]rune, cfg.outlinelenlimit+1)
	}
	cfg.inchrlinelenlimit = cfg.outputwidth*4 + 100
	cfg.inchrline = make([]rune, cfg.inchrlinelenlimit+1)
	clearline(cfg)
}

func getletter(cfg *Config, c rune) {
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

func smushem(cfg *Config, lch, rch rune) rune {
	if lch == ' ' {
		return rch
	}
	if rch == ' ' {
		return lch
	}

	if cfg.previouscharwidth < 2 || cfg.currcharwidth < 2 {
		return 0
	}

	if (cfg.smushmode & SM_SMUSH) == 0 {
		return 0
	}

	if (cfg.smushmode & 63) == 0 {
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
		if cfg.right2left == 1 {
			return lch
		}
		return rch
	}

	if (cfg.smushmode & SM_HARDBLANK) != 0 {
		if lch == cfg.hardblank && rch == cfg.hardblank {
			return lch
		}
	}

	if lch == cfg.hardblank || rch == cfg.hardblank {
		return 0
	}

	if (cfg.smushmode & SM_EQUAL) != 0 {
		if lch == rch {
			return lch
		}
	}

	if (cfg.smushmode & SM_LOWLINE) != 0 {
		if lch == '_' && strings.ContainsRune("|/\\[]{}()<>", rch) {
			return rch
		}
		if rch == '_' && strings.ContainsRune("|/\\[]{}()<>", lch) {
			return lch
		}
	}

	if (cfg.smushmode & SM_HIERARCHY) != 0 {
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

	if (cfg.smushmode & SM_PAIR) != 0 {
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

	if (cfg.smushmode & SM_BIGX) != 0 {
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

func smushamt(cfg *Config) int {
	if (cfg.smushmode & (SM_SMUSH | SM_KERN)) == 0 {
		return 0
	}
	maxsmush := cfg.currcharwidth
	for row := 0; row < cfg.charheight; row++ {
		var amt int
		var ch1, ch2 rune

		if cfg.right2left == 1 {
			// C: for (charbd=STRLEN(currchar[row]);
			//      ch1=currchar[row][charbd],(charbd>0&&(!ch1||ch1==' '));charbd--) ;
			charbd := len(cfg.currchar[row])
			// First evaluation of condition (sets ch1)
			if charbd < len(cfg.currchar[row]) {
				ch1 = cfg.currchar[row][charbd]
			} else {
				ch1 = 0 // null terminator equivalent
			}
			for charbd > 0 && (ch1 == 0 || ch1 == ' ') {
				charbd--
				if charbd < len(cfg.currchar[row]) {
					ch1 = cfg.currchar[row][charbd]
				} else {
					ch1 = 0
				}
			}

			// C: for (linebd=0;ch2=outputline[row][linebd],ch2==' ';linebd++) ;
			linebd := 0
			if linebd < len(cfg.outputline[row]) {
				ch2 = cfg.outputline[row][linebd]
			} else {
				ch2 = 0
			}
			for ch2 == ' ' {
				linebd++
				if linebd < len(cfg.outputline[row]) {
					ch2 = cfg.outputline[row][linebd]
				} else {
					ch2 = 0
					break
				}
			}
			amt = linebd + cfg.currcharwidth - 1 - charbd
		} else {
			// C: for (linebd=STRLEN(outputline[row]);
			//      ch1 = outputline[row][linebd],(linebd>0&&(!ch1||ch1==' '));linebd--) ;
			linebd := len(cfg.outputline[row])
			// First evaluation of condition (sets ch1)
			if linebd < len(cfg.outputline[row]) {
				ch1 = cfg.outputline[row][linebd]
			} else {
				ch1 = 0 // null terminator equivalent
			}
			for linebd > 0 && (ch1 == 0 || ch1 == ' ') {
				linebd--
				if linebd < len(cfg.outputline[row]) {
					ch1 = cfg.outputline[row][linebd]
				} else {
					ch1 = 0
				}
			}

			// C: for (charbd=0;ch2=currchar[row][charbd],ch2==' ';charbd++) ;
			charbd := 0
			if charbd < len(cfg.currchar[row]) {
				ch2 = cfg.currchar[row][charbd]
			} else {
				ch2 = 0
			}
			for ch2 == ' ' {
				charbd++
				if charbd < len(cfg.currchar[row]) {
					ch2 = cfg.currchar[row][charbd]
				} else {
					ch2 = 0
					break
				}
			}
			amt = charbd + cfg.outlinelen - 1 - linebd
		}

		// C: if (!ch1||ch1==' ') { amt++; }
		if ch1 == 0 || ch1 == ' ' {
			amt++
		} else if ch2 != 0 {
			// C: else if (ch2) { if (smushem(ch1,ch2)!='\0') { amt++; } }
			if smushem(cfg, ch1, ch2) != 0 {
				amt++
			}
		}

		if amt < maxsmush {
			maxsmush = amt
		}
	}
	return maxsmush
}

func addchar(cfg *Config, c rune) bool {
	getletter(cfg, c)
	smushamount := smushamt(cfg)
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

	for row := 0; row < cfg.charheight; row++ {
		if cfg.right2left == 1 {
			templine := make([]rune, len(cfg.currchar[row]))
			copy(templine, cfg.currchar[row])
			for k := 0; k < smushamount && k < len(cfg.outputline[row]); k++ {
				idx := cfg.currcharwidth - smushamount + k
				if idx >= 0 && idx < len(templine) {
					smushed := smushem(cfg, templine[idx], cfg.outputline[row][k])
					if smushed != 0 {
						templine[idx] = smushed
					}
				}
			}
			remaining := len(cfg.outputline[row])
			if smushamount < remaining {
				cfg.outputline[row] = append(templine, cfg.outputline[row][smushamount:]...)
			} else {
				cfg.outputline[row] = templine
			}
		} else {
			for k := 0; k < smushamount; k++ {
				column := cfg.outlinelen - smushamount + k
				if column < 0 {
					column = 0
				}
				if column < len(cfg.outputline[row]) && k < len(cfg.currchar[row]) {
					cfg.outputline[row][column] = smushem(cfg, cfg.outputline[row][column], cfg.currchar[row][k])
				}
			}
			if smushamount < len(cfg.currchar[row]) {
				cfg.outputline[row] = append(cfg.outputline[row], cfg.currchar[row][smushamount:]...)
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

func putstring(cfg *Config, str []rune) {
	length := len(str)
	if cfg.outputwidth > 1 {
		if length > cfg.outputwidth-1 {
			length = cfg.outputwidth - 1
		}
		if cfg.justification > 0 {
			for i := 1; (3-cfg.justification)*i+length+cfg.justification-2 < cfg.outputwidth; i++ {
				fmt.Print(" ")
			}
		}
	}
	for i := 0; i < length; i++ {
		if i < len(str) {
			if str[i] == cfg.hardblank {
				fmt.Print(" ")
			} else {
				fmt.Print(string(str[i]))
			}
		}
	}
	fmt.Println()
}

func printline(cfg *Config) {
	for i := 0; i < cfg.charheight; i++ {
		putstring(cfg, cfg.outputline[i])
	}
	clearline(cfg)
}

func splitline(cfg *Config) {
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
	clearline(cfg)
	for i := 0; i < len1; i++ {
		addchar(cfg, part1[i])
	}
	printline(cfg)
	for i := 0; i < len2; i++ {
		addchar(cfg, part2[i])
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
	if !cfg.cmdinput {
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
	if cfg.agetmode < 0 || cfg.optind >= len(cfg.argv) {
		return -1
	}

	// find next character
	arg := cfg.argv[cfg.optind]
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
		cfg.optind++                     // run up word count
		if cfg.optind >= len(cfg.argv) { // check if at "EOF"
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

	switch cfg.multibyte {
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
