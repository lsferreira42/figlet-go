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
}
