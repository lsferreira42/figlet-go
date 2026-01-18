// This file provides the main entry point for the figlet executable.
// The core functionality is in the figlet package (github.com/lsferreira42/figlet-go/figlet).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lsferreira42/figlet-go/figlet"
)

func main() {
	cfg := figlet.New()
	cfg.Argv = os.Args

	getparams(cfg)
	if err := cfg.LoadFont(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", getmyname(cfg.Argv), err)
		os.Exit(1)
	}

	processInput(cfg)
}

func getmyname(argv []string) string {
	if len(argv) == 0 {
		return "figlet"
	}
	name := filepath.Base(argv[0])
	return name
}

func printusage(cfg *figlet.Config, out io.Writer) {
	myname := getmyname(cfg.Argv)
	fmt.Fprintf(out, "Usage: %s [ -cklnoprstvxDELNRSWX ] [ -d fontdirectory ]\n", myname)
	fmt.Fprintf(out, "              [ -f fontfile ] [ -m smushmode ] [ -w outputwidth ]\n")
	fmt.Fprintf(out, "              [ -C controlfile ] [ -I infocode ]\n")
	fmt.Fprintf(out, "              [ --colors color1;color2;... ] [ --parser terminal|terminal-color|html ]\n")
	fmt.Fprintf(out, "              [ message ]\n")
}

func printinfo(cfg *figlet.Config, infonum int) {
	switch infonum {
	case 0:
		fmt.Printf("FIGlet Copyright (C) 1991-2012 Glenn Chappell, Ian Chai, ")
		fmt.Printf("John Cowan,\nChristiaan Keet and Claudio Matsuoka\n")
		fmt.Printf("Internet: <info@figlet.org> ")
		fmt.Printf("Version: %s, date: %s\n\n", figlet.VERSION, figlet.DATE)
		fmt.Printf("FIGlet, along with the various FIGlet fonts")
		fmt.Printf(" and documentation, may be\n")
		fmt.Printf("freely copied and distributed.\n\n")
		fmt.Printf("If you use FIGlet, please send an")
		fmt.Printf(" e-mail message to <info@figlet.org>.\n\n")
		fmt.Printf("The latest version of FIGlet is available from the")
		fmt.Printf(" web site,\n\thttp://www.figlet.org/\n\n")
		printusage(cfg, os.Stdout)
	case 1:
		fmt.Printf("%d\n", figlet.VERSION_INT)
	case 2:
		fmt.Printf("%s\n", cfg.Fontdirname)
	case 3:
		fmt.Printf("%s\n", cfg.Fontname)
	case 4:
		fmt.Printf("%d\n", cfg.Outputwidth)
	case 5:
		fmt.Printf("%s", figlet.FONTFILEMAGICNUMBER)
		fmt.Printf(" %s", figlet.TOILETFILEMAGICNUMBER)
		fmt.Printf("\n")
	}
}

func suffixcmp(s1, s2 string) bool {
	return len(s1) >= len(s2) && s1[len(s1)-len(s2):] == s2
}

func getparams(cfg *figlet.Config) {
	myname := getmyname(cfg.Argv)
	cfg.Fontdirname = "fonts"
	if env := os.Getenv("FIGLET_FONTDIR"); env != "" {
		cfg.Fontdirname = env
	}
	cfg.Fontname = "standard"
	cfg.Smushoverride = figlet.SMO_NO
	cfg.Deutschflag = false
	cfg.Justification = -1
	cfg.Right2left = -1
	cfg.Paragraphflag = false
	infoprint := -1
	cfg.Cmdinput = false
	cfg.Outputwidth = figlet.DEFAULTCOLUMNS

	// Simple getopt implementation
	optind := 1
	for optind < len(cfg.Argv) {
		arg := cfg.Argv[optind]
		if len(arg) == 0 || arg[0] != '-' {
			cfg.Cmdinput = true
			cfg.Optind = optind
			break
		}
		if arg == "--" {
			optind++
			cfg.Cmdinput = true
			cfg.Optind = optind
			break
		}
		
		// Handle long options (--colors, --parser)
		if len(arg) > 2 && arg[0:2] == "--" {
			if strings.HasPrefix(arg, "--colors=") {
				colorsStr := arg[9:]
				colors := parseColors(colorsStr)
				cfg.Colors = colors
				// Only auto-switch to terminal-color if parser is still default terminal
				// Don't override if user explicitly set a parser (like HTML)
				if len(colors) > 0 && cfg.OutputParser != nil && cfg.OutputParser.Name == "terminal" {
					parser, _ := figlet.GetParser("terminal-color")
					cfg.OutputParser = parser
				}
			} else if arg == "--colors" && optind+1 < len(cfg.Argv) {
				colorsStr := cfg.Argv[optind+1]
				colors := parseColors(colorsStr)
				cfg.Colors = colors
				// Only auto-switch to terminal-color if parser is still default terminal
				// Don't override if user explicitly set a parser (like HTML)
				if len(colors) > 0 && cfg.OutputParser != nil && cfg.OutputParser.Name == "terminal" {
					parser, _ := figlet.GetParser("terminal-color")
					cfg.OutputParser = parser
				}
				optind++
			} else if strings.HasPrefix(arg, "--parser=") {
				parserName := arg[9:]
				parser, err := figlet.GetParser(parserName)
				if err == nil {
					cfg.OutputParser = parser
				}
			} else if arg == "--parser" && optind+1 < len(cfg.Argv) {
				parserName := cfg.Argv[optind+1]
				parser, err := figlet.GetParser(parserName)
				if err == nil {
					cfg.OutputParser = parser
				}
				optind++
			}
			optind++
			continue
		}

		for i := 1; i < len(arg); i++ {
			c := arg[i]
			switch c {
			case 'A':
				cfg.Cmdinput = true
			case 'D':
				cfg.Deutschflag = true
			case 'E':
				cfg.Deutschflag = false
			case 'X':
				cfg.Right2left = -1
			case 'L':
				cfg.Right2left = 0
			case 'R':
				cfg.Right2left = 1
			case 'x':
				cfg.Justification = -1
			case 'l':
				cfg.Justification = 0
			case 'c':
				cfg.Justification = 1
			case 'r':
				cfg.Justification = 2
			case 'p':
				cfg.Paragraphflag = true
			case 'n':
				cfg.Paragraphflag = false
			case 's':
				cfg.Smushoverride = figlet.SMO_NO
			case 'k':
				cfg.Smushmode = figlet.SM_KERN
				cfg.Smushoverride = figlet.SMO_YES
			case 'S':
				cfg.Smushmode = figlet.SM_SMUSH
				cfg.Smushoverride = figlet.SMO_FORCE
			case 'o':
				cfg.Smushmode = figlet.SM_SMUSH
				cfg.Smushoverride = figlet.SMO_YES
			case 'W':
				cfg.Smushmode = 0
				cfg.Smushoverride = figlet.SMO_YES
			case 't':
				columns := figlet.GetColumns()
				if columns > 0 {
					cfg.Outputwidth = columns
				}
			case 'v':
				infoprint = 0
			case 'I':
				if i+1 < len(arg) {
					val, _ := strconv.Atoi(arg[i+1:])
					infoprint = val
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					val, _ := strconv.Atoi(cfg.Argv[optind+1])
					infoprint = val
					optind++
				}
			case 'm':
				var val int
				if i+1 < len(arg) {
					val, _ = strconv.Atoi(arg[i+1:])
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					val, _ = strconv.Atoi(cfg.Argv[optind+1])
					optind++
				}
				if val < -1 {
					cfg.Smushoverride = figlet.SMO_NO
					break
				}
				if val == 0 {
					cfg.Smushmode = figlet.SM_KERN
				} else if val == -1 {
					cfg.Smushmode = 0
				} else {
					cfg.Smushmode = (val & 63) | figlet.SM_SMUSH
				}
				cfg.Smushoverride = figlet.SMO_YES
			case 'w':
				var val int
				if i+1 < len(arg) {
					val, _ = strconv.Atoi(arg[i+1:])
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					val, _ = strconv.Atoi(cfg.Argv[optind+1])
					optind++
				}
				if val > 0 {
					cfg.Outputwidth = val
				}
			case 'd':
				if i+1 < len(arg) {
					cfg.Fontdirname = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					cfg.Fontdirname = cfg.Argv[optind+1]
					optind++
				}
			case 'f':
				var name string
				if i+1 < len(arg) {
					name = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					name = cfg.Argv[optind+1]
					optind++
				}
				cfg.Fontname = name
				if suffixcmp(cfg.Fontname, figlet.FONTFILESUFFIX) {
					cfg.Fontname = cfg.Fontname[:len(cfg.Fontname)-len(figlet.FONTFILESUFFIX)]
				} else if suffixcmp(cfg.Fontname, figlet.TOILETFILESUFFIX) {
					cfg.Fontname = cfg.Fontname[:len(cfg.Fontname)-len(figlet.TOILETFILESUFFIX)]
				}
			case 'C':
				var name string
				if i+1 < len(arg) {
					name = arg[i+1:]
					i = len(arg)
				} else if optind+1 < len(cfg.Argv) {
					name = cfg.Argv[optind+1]
					optind++
				}
				cfg.AddControlFile(name)
			case 'N':
				cfg.ClearControlFiles()
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

	if optind < len(cfg.Argv) {
		cfg.Cmdinput = true
		cfg.Optind = optind
	}

	if infoprint >= 0 {
		printinfo(cfg, infoprint)
		os.Exit(0)
	}
}

// parseColors parses a color string (e.g., "red;green;blue" or "FF0000;00FF00")
func parseColors(colorsStr string) []figlet.Color {
	if colorsStr == "" {
		return nil
	}
	
	parts := strings.Split(colorsStr, ";")
	colors := make([]figlet.Color, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Try predefined color names
		switch strings.ToLower(part) {
		case "black":
			colors = append(colors, figlet.ColorBlack)
		case "red":
			colors = append(colors, figlet.ColorRed)
		case "green":
			colors = append(colors, figlet.ColorGreen)
		case "yellow":
			colors = append(colors, figlet.ColorYellow)
		case "blue":
			colors = append(colors, figlet.ColorBlue)
		case "magenta":
			colors = append(colors, figlet.ColorMagenta)
		case "cyan":
			colors = append(colors, figlet.ColorCyan)
		case "white":
			colors = append(colors, figlet.ColorWhite)
		default:
			// Try to parse as hex color
			tc, err := figlet.NewTrueColorFromHexString(part)
			if err == nil {
				colors = append(colors, tc)
			}
		}
	}
	
	return colors
}

func processInput(cfg *figlet.Config) {
	if cfg.Cmdinput && cfg.Optind < len(cfg.Argv) {
		// Build the text from command line arguments
		text := ""
		for i := cfg.Optind; i < len(cfg.Argv); i++ {
			if i > cfg.Optind {
				text += " "
			}
			text += cfg.Argv[i]
		}
		result := cfg.RenderString(text)
		fmt.Print(result)
	} else {
		// Read from stdin
		var input []byte
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				input = append(input, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		if len(input) > 0 {
			result := cfg.RenderString(string(input))
			fmt.Print(result)
		}
	}
}
