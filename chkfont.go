package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	CHKFONT_DATE    = "31 May 2012"
	CHKFONT_VERSION = "2.2.5"
)

var possHardblanks = []byte{'!', '@', '#', '$', '%', '&', '*', 0x7f}

type FontChecker struct {
	myname          string
	fontfilename    string
	fontfile        *os.File
	hardblank       byte
	charheight      int
	upheight        int
	maxlen          int
	oldLayout       int
	spectagcnt      int
	fileline        string
	maxlinelength   int
	currline        int
	ec              int // error count
	wc              int // warning count
	inconEndmarkWarn   bool
	endmarkCountWarn   bool
	nonincrWarn        bool
	bigCodetagWarn     bool
	deutschCodetagWarn bool
	asciiCodetagWarn   bool
	codetagcnt      int
	gone            bool
	scanner         *bufio.Scanner
}

func newFontChecker(myname string) *FontChecker {
	return &FontChecker{
		myname:     myname,
		spectagcnt: -1,
	}
}

func (fc *FontChecker) weregone(really bool) {
	if !really && 2*fc.ec+fc.wc <= 40 {
		return
	}
	if fc.ec+fc.wc > 0 {
		fmt.Println("*******************************************************************************")
	}
	if !really {
		fmt.Printf("%s: Too many errors/warnings.\n", fc.fontfilename)
	}
	fmt.Printf("%s: Errors: %d, Warnings: %d\n", fc.fontfilename, fc.ec, fc.wc)
	if fc.currline > 1 && fc.maxlen != fc.maxlinelength {
		fmt.Printf("%s: maxlen: %d, actual max line length: %d\n",
			fc.fontfilename, fc.maxlen, fc.maxlinelength)
		if fc.codetagcnt > 0 && fc.spectagcnt == -1 {
			fmt.Printf("%s: Code-tagged characters: %d\n", fc.fontfilename, fc.codetagcnt)
		}
	}
	fmt.Println("-------------------------------------------------------------------------------")
	fc.gone = true
}

func (fc *FontChecker) badsuffix(path, suffix string) bool {
	ucsuffix := strings.ToUpper(suffix)
	if len(path) < len(suffix) {
		return true
	}
	s := path[len(path)-len(suffix):]
	if s == suffix || s == ucsuffix {
		return false
	}
	return true
}

func (fc *FontChecker) readLine() (string, bool) {
	if fc.scanner.Scan() {
		return fc.scanner.Text(), true
	}
	return "", false
}

func (fc *FontChecker) readchar() {
	var expectedWidth int
	var expectedEndmark byte
	var minLeadblanks, minTrailblanks int

	for i := 0; i < fc.charheight; i++ {
		line, ok := fc.readLine()
		if !ok {
			if fc.scanner.Err() != nil {
				fmt.Printf("%s: ERROR (fatal)- Unexpected read error after line %d.\n",
					fc.fontfilename, fc.currline)
			} else {
				fmt.Printf("%s: ERROR (fatal)- Unexpected end of file after line %d.\n",
					fc.fontfilename, fc.currline)
			}
			fc.ec++
			fc.weregone(true)
			return
		}
		fc.currline++
		lineLen := len(line)
		if lineLen > fc.maxlinelength {
			fc.maxlinelength = lineLen
		}
		if lineLen > fc.maxlen {
			fmt.Printf("%s: ERROR- Line length > maxlen in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.ec++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}

		// Find endmark
		k := lineLen - 1
		var endmark byte
		if k < 0 {
			endmark = 0
		} else {
			endmark = line[k]
		}

		// Remove endmarks from the end
		for k >= 0 && line[k] == endmark {
			k--
		}
		newlen := k + 1
		cleanLine := ""
		if newlen > 0 {
			cleanLine = line[:newlen]
		}

		// Count leading blanks
		leadblanks := 0
		for l := 0; l < len(cleanLine) && cleanLine[l] == ' '; l++ {
			leadblanks++
		}

		// Count trailing blanks
		trailblanks := 0
		for l := len(cleanLine) - 1; l >= 0 && cleanLine[l] == ' '; l-- {
			trailblanks++
		}

		if i == 0 {
			expectedEndmark = endmark
			expectedWidth = newlen
			minLeadblanks = leadblanks
			minTrailblanks = trailblanks
			if endmark == ' ' {
				fmt.Printf("%s: Warning- Blank endmark in line %d.\n",
					fc.fontfilename, fc.currline)
				fc.wc++
				fc.weregone(false)
				if fc.gone {
					return
				}
			}
		} else {
			if leadblanks < minLeadblanks {
				minLeadblanks = leadblanks
			}
			if trailblanks < minTrailblanks {
				minTrailblanks = trailblanks
			}
			if endmark != expectedEndmark && !fc.inconEndmarkWarn {
				fmt.Printf("%s: Warning- Inconsistent endmark in line %d.\n",
					fc.fontfilename, fc.currline)
				fmt.Printf("%s:          (Above warning will only be printed once.)\n",
					fc.fontfilename)
				fc.inconEndmarkWarn = true
				fc.wc++
				fc.weregone(false)
				if fc.gone {
					return
				}
			}
			if newlen != expectedWidth {
				fmt.Printf("%s: ERROR- Inconsistent character width in line %d.\n",
					fc.fontfilename, fc.currline)
				fc.ec++
				fc.weregone(false)
				if fc.gone {
					return
				}
			}
		}

		diff := lineLen - newlen
		if diff > 2 {
			fmt.Printf("%s: ERROR- Too many endmarks in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.ec++
			fc.weregone(false)
			if fc.gone {
				return
			}
		} else if fc.charheight > 1 {
			expectedDiff := 1
			if i == fc.charheight-1 {
				expectedDiff = 2
			}
			if diff != expectedDiff && !fc.endmarkCountWarn {
				fmt.Printf("%s: Warning- Endchar count convention violated in line %d.\n",
					fc.fontfilename, fc.currline)
				fmt.Printf("%s:          (Above warning will only be printed once.)\n",
					fc.fontfilename)
				fc.endmarkCountWarn = true
				fc.wc++
				fc.weregone(false)
				if fc.gone {
					return
				}
			}
		}
	}
	// Suppress unused variable warnings
	_ = minLeadblanks
	_ = minTrailblanks
}

func (fc *FontChecker) checkit() {
	fc.ec = 0
	fc.wc = 0
	fc.inconEndmarkWarn = false
	fc.endmarkCountWarn = false
	fc.nonincrWarn = false
	fc.bigCodetagWarn = false
	fc.deutschCodetagWarn = false
	fc.asciiCodetagWarn = false
	fc.codetagcnt = 0
	fc.gone = false
	fc.maxlinelength = 0

	if fc.fontfilename == "-" {
		fc.fontfilename = "(stdin)"
		fc.fontfile = os.Stdin
	} else {
		var err error
		fc.fontfile, err = os.Open(fc.fontfilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Could not open file '%s'\n", fc.myname, fc.fontfilename)
			os.Exit(1)
		}
		defer fc.fontfile.Close()
	}

	fc.scanner = bufio.NewScanner(fc.fontfile)

	// Check filename suffix
	if fc.fontfile != os.Stdin {
		if fc.badsuffix(fc.fontfilename, ".flf") {
			fmt.Printf("%s: ERROR- Filename does not end with '.flf'.\n", fc.fontfilename)
			fc.ec++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}
	}

	// Read first line
	firstLine, ok := fc.readLine()
	if !ok {
		fmt.Printf("%s: ERROR- can't read magic number.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
		return
	}

	// Check magic number
	if len(firstLine) < 4 {
		fmt.Printf("%s: ERROR- can't read magic number.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
		return
	}

	magicnum := firstLine[:4]
	if magicnum != "flf2" {
		fmt.Printf("%s: ERROR- Incorrect magic number.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check sub-version
	if len(firstLine) < 5 {
		fmt.Printf("%s: Warning- Sub-version character is not 'a'.\n", fc.fontfilename)
		fc.wc++
		fc.weregone(false)
		if fc.gone {
			return
		}
		// Can't parse header, report fatal error
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	} else if firstLine[4] != 'a' {
		fmt.Printf("%s: Warning- Sub-version character is not 'a'.\n", fc.fontfilename)
		fc.wc++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Parse header line
	headerPart := firstLine[5:]
	fields := strings.Fields(headerPart)
	if len(fields) < 6 {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	// Parse hardblank (first character after space)
	if len(fields[0]) < 1 {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}
	fc.hardblank = fields[0][0]

	var err error
	fc.charheight, err = strconv.Atoi(fields[1])
	if err != nil {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	fc.upheight, err = strconv.Atoi(fields[2])
	if err != nil {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	fc.maxlen, err = strconv.Atoi(fields[3])
	if err != nil {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	fc.oldLayout, err = strconv.Atoi(fields[4])
	if err != nil {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	cmtcount, err := strconv.Atoi(fields[5])
	if err != nil {
		fmt.Printf("%s: ERROR (fatal)- First line improperly formatted.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	ffrighttoleft := 0
	if len(fields) >= 7 {
		ffrighttoleft, _ = strconv.Atoi(fields[6])
	}

	var layout int
	haveLayout := false
	if len(fields) >= 8 {
		layout, _ = strconv.Atoi(fields[7])
		haveLayout = true
	}

	if len(fields) >= 9 {
		fc.spectagcnt, _ = strconv.Atoi(fields[8])
	} else {
		fc.spectagcnt = -1
	}

	// Check hardblank
	foundHardblank := false
	for _, hb := range possHardblanks {
		if fc.hardblank == hb {
			foundHardblank = true
			break
		}
	}
	if !foundHardblank {
		fmt.Printf("%s: Warning- Unusual hardblank.\n", fc.fontfilename)
		fc.wc++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check charheight
	if fc.charheight < 1 {
		fmt.Printf("%s: ERROR (fatal)- charheight not positive.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	// Check upheight
	if fc.upheight > fc.charheight || fc.upheight < 1 {
		fmt.Printf("%s: ERROR- up_height out of bounds.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check maxlen
	if fc.maxlen < 1 {
		fmt.Printf("%s: ERROR (fatal)- maxlen not positive.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(true)
		return
	}

	// Check old_layout
	if fc.oldLayout < -1 {
		fmt.Printf("%s: ERROR- old_layout < -1.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}
	if fc.oldLayout > 63 {
		fmt.Printf("%s: ERROR- old_layout > 63.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check layout
	if haveLayout && layout < 0 {
		fmt.Printf("%s: ERROR- layout < 0.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}
	if haveLayout && layout > 32767 {
		fmt.Printf("%s: ERROR- layout > 32767.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}
	if haveLayout && fc.oldLayout == -1 && (layout&192) != 0 {
		fmt.Printf("%s: ERROR- layout %d is inconsistent with old_layout -1.\n",
			fc.fontfilename, layout)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}
	if haveLayout && fc.oldLayout == 0 && (layout&192) != 64 && (layout&255) != 128 {
		fmt.Printf("%s: ERROR- layout %d is inconsistent with old_layout 0.\n",
			fc.fontfilename, layout)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}
	if haveLayout && fc.oldLayout > 0 &&
		((layout&128) == 0 || fc.oldLayout != (layout&63)) {
		fmt.Printf("%s: ERROR- layout %d is inconsistent with old_layout %d.\n",
			fc.fontfilename, layout, fc.oldLayout)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check cmtcount
	if cmtcount < 0 {
		fmt.Printf("%s: ERROR- cmt_count is negative.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Check ffrighttoleft
	if ffrighttoleft < 0 || ffrighttoleft > 1 {
		fmt.Printf("%s: ERROR- rtol out of bounds.\n", fc.fontfilename)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	// Skip comment lines
	for i := 0; i < cmtcount; i++ {
		_, ok := fc.readLine()
		if !ok {
			fmt.Printf("%s: ERROR (fatal)- Unexpected end of file in comments.\n", fc.fontfilename)
			fc.ec++
			fc.weregone(true)
			return
		}
	}

	fc.currline = cmtcount + 1

	// Read 102 required characters (95 ASCII + 7 German)
	for i := 0; i < 102; i++ {
		fc.readchar()
		if fc.gone {
			return
		}
	}

	// Read code-tagged characters
	var oldord int64 = 0
	for {
		line, ok := fc.readLine()
		if !ok {
			break
		}
		fc.currline++

		lineLen := len(line)
		if lineLen-100 > fc.maxlinelength {
			fc.maxlinelength = lineLen - 100
		}
		if lineLen > fc.maxlen+100 {
			fmt.Printf("%s: ERROR- Code tag line way too long in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.ec++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}

		// Parse code tag
		fields := strings.Fields(line)
		if len(fields) < 1 {
			fmt.Printf("%s: Warning- Extra chars after font in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
			break
		}

		theord, err := strconv.ParseInt(fields[0], 0, 64)
		if err != nil {
			fmt.Printf("%s: Warning- Extra chars after font in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
			break
		}

		fc.codetagcnt++

		if theord > 65535 && !fc.bigCodetagWarn {
			fmt.Printf("%s: Warning- Code tag > 65535 in line %d.\n",
				fc.fontfilename, fc.currline)
			fmt.Printf("%s:          (Above warning will only be printed once.)\n",
				fc.fontfilename)
			fc.bigCodetagWarn = true
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}

		if theord == -1 {
			fmt.Printf("%s: ERROR- Code tag -1 (unusable) in line %d.\n",
				fc.fontfilename, fc.currline)
			fc.ec++
			fc.weregone(false)
			if fc.gone {
				return
			}
			break
		}

		if theord >= -255 && theord <= -249 && !fc.deutschCodetagWarn {
			fmt.Printf("%s: Warning- Code tag in old Deutsch area in line %d.\n",
				fc.fontfilename, fc.currline)
			fmt.Printf("%s:          (Above warning will only be printed once.)\n",
				fc.fontfilename)
			fc.deutschCodetagWarn = true
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}

		if theord < 127 && theord > 31 && !fc.asciiCodetagWarn {
			fmt.Printf("%s: Warning- Code tag in ASCII range in line %d.\n",
				fc.fontfilename, fc.currline)
			fmt.Printf("%s:          (Above warning will only be printed once.)\n",
				fc.fontfilename)
			fc.asciiCodetagWarn = true
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
		} else if theord <= oldord && theord >= 0 && oldord >= 0 && !fc.nonincrWarn {
			fmt.Printf("%s: Warning- Non-increasing code tag in line %d.\n",
				fc.fontfilename, fc.currline)
			fmt.Printf("%s:          (Above warning will only be printed once.)\n",
				fc.fontfilename)
			fc.nonincrWarn = true
			fc.wc++
			fc.weregone(false)
			if fc.gone {
				return
			}
		}
		oldord = theord

		fc.readchar()
		if fc.gone {
			return
		}
	}

	// Check spectagcnt
	if fc.spectagcnt != -1 && fc.spectagcnt != fc.codetagcnt {
		fmt.Printf("%s: ERROR- Inconsistent Codetag_Cnt value %d\n",
			fc.fontfilename, fc.spectagcnt)
		fc.ec++
		fc.weregone(false)
		if fc.gone {
			return
		}
	}

	fc.weregone(true)
}

func usageerr(myname string) {
	fmt.Fprintf(os.Stderr, "chkfont by Glenn Chappell <ggc@uiuc.edu>\n")
	fmt.Fprintf(os.Stderr, "Version: %s, date: %s\n", CHKFONT_VERSION, CHKFONT_DATE)
	fmt.Fprintf(os.Stderr, "Checks figlet 2.0/2.1 font files for format errors.\n")
	fmt.Fprintf(os.Stderr, "(Does not modify font files.)\n")
	fmt.Fprintf(os.Stderr, "Usage: %s fontfile ...\n", myname)
	os.Exit(1)
}

func main() {
	myname := filepath.Base(os.Args[0])

	if len(os.Args) < 2 {
		usageerr(myname)
	}

	fc := newFontChecker(myname)

	for _, fontfile := range os.Args[1:] {
		fc.fontfilename = fontfile
		fc.checkit()
	}
}
