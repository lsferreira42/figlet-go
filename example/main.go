// Example demonstrating how to use the figlet library
package main

import (
	"fmt"
	"log"

	"github.com/lsferreira42/figlet-go/figlet"
)

func main() {
	// Simple usage with default font
	result, err := figlet.Render("Hello!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Default font ===")
	fmt.Print(result)

	// Using a specific font
	result, err = figlet.RenderWithFont("Go!", "slant")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Slant font ===")
	fmt.Print(result)

	// Using options
	result, err = figlet.Render("Options",
		figlet.WithFont("big"),
		figlet.WithWidth(60),
		figlet.WithJustification(1), // center
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Big font, centered ===")
	fmt.Print(result)

	// Using Config directly for more control
	cfg := figlet.New()
	cfg.Fontname = "banner"
	cfg.Outputwidth = 100

	if err := cfg.LoadFont(); err != nil {
		log.Fatal(err)
	}

	result = cfg.RenderString("Config")
	fmt.Println("=== Banner font via Config ===")
	fmt.Print(result)

	// List available fonts
	fmt.Println("=== Available fonts ===")
	fonts := figlet.ListFonts()
	for _, f := range fonts {
		fmt.Printf("  - %s\n", f)
	}

	// Get version info
	fmt.Printf("\nFIGlet version: %s (int: %d)\n", figlet.GetVersion(), figlet.GetVersionInt())

	// With colors (ANSI)
	fmt.Println("\n=== Colors (ANSI) ===")
	result, err = figlet.Render("Colors!",
		figlet.WithColors(figlet.ColorRed, figlet.ColorGreen, figlet.ColorBlue),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)

	// With TrueColor (hex)
	fmt.Println("\n=== TrueColor (hex) ===")
	tcRed, _ := figlet.NewTrueColorFromHexString("FF0000")
	tcGreen, _ := figlet.NewTrueColorFromHexString("00FF00")
	tcBlue, _ := figlet.NewTrueColorFromHexString("0000FF")
	result, err = figlet.Render("TrueColor",
		figlet.WithColors(tcRed, tcGreen, tcBlue),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)

	// HTML output
	fmt.Println("\n=== HTML Output ===")
	result, err = figlet.Render("HTML",
		figlet.WithParser("html"),
		figlet.WithColors(figlet.ColorRed, figlet.ColorBlue),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)

	// Terminal with colors
	fmt.Println("\n=== Terminal with Colors ===")
	result, err = figlet.Render("Terminal Colors",
		figlet.WithParser("terminal-color"),
		figlet.WithColors(figlet.ColorYellow, figlet.ColorCyan, figlet.ColorMagenta),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)
}
