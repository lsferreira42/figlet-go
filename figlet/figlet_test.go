package figlet

import (
	"strings"
	"testing"
)

// TestRender tests the basic Render function
func TestRender(t *testing.T) {
	result, err := Render("Hi")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if result == "" {
		t.Error("Render returned empty string")
	}
	if !strings.Contains(result, "\n") {
		t.Error("Render output should contain newlines")
	}
}

// TestRenderWithFont tests rendering with different fonts
func TestRenderWithFont(t *testing.T) {
	fonts := []string{"standard", "banner", "big", "slant", "small"}
	for _, font := range fonts {
		t.Run(font, func(t *testing.T) {
			result, err := RenderWithFont("Test", font)
			if err != nil {
				t.Fatalf("RenderWithFont(%q) failed: %v", font, err)
			}
			if result == "" {
				t.Errorf("RenderWithFont(%q) returned empty string", font)
			}
		})
	}
}

// TestRenderInvalidFont tests that invalid fonts return an error
func TestRenderInvalidFont(t *testing.T) {
	_, err := RenderWithFont("Test", "nonexistent_font_12345")
	if err == nil {
		t.Error("Expected error for invalid font, got nil")
	}
}

// TestRenderEmptyString tests rendering an empty string
func TestRenderEmptyString(t *testing.T) {
	result, err := Render("")
	if err != nil {
		t.Fatalf("Render empty string failed: %v", err)
	}
	// Empty string should produce empty or minimal output
	_ = result // Just check it doesn't crash
}

// TestWithFont tests the WithFont option
func TestWithFont(t *testing.T) {
	result, err := Render("A", WithFont("banner"))
	if err != nil {
		t.Fatalf("Render with WithFont failed: %v", err)
	}
	// Banner font uses # characters
	if !strings.Contains(result, "#") {
		t.Error("Banner font output should contain # characters")
	}
}

// TestWithWidth tests the WithWidth option
func TestWithWidth(t *testing.T) {
	result, err := Render("Hello World", WithWidth(40))
	if err != nil {
		t.Fatalf("Render with WithWidth failed: %v", err)
	}
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if len(line) > 40 {
			t.Errorf("Line exceeds width 40: len=%d", len(line))
		}
	}
}

// TestWithJustification tests justification options
func TestWithJustification(t *testing.T) {
	tests := []struct {
		name string
		just int
	}{
		{"left", 0},
		{"center", 1},
		{"right", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render("X", WithJustification(tt.just), WithWidth(80))
			if err != nil {
				t.Fatalf("Render with justification %d failed: %v", tt.just, err)
			}
			if result == "" {
				t.Error("Result should not be empty")
			}
		})
	}
}

// TestWithKerning tests kerning mode
func TestWithKerning(t *testing.T) {
	result, err := Render("AB", WithKerning())
	if err != nil {
		t.Fatalf("Render with WithKerning failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestWithFullWidth tests full width mode
func TestWithFullWidth(t *testing.T) {
	result, err := Render("AB", WithFullWidth())
	if err != nil {
		t.Fatalf("Render with WithFullWidth failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestWithSmushing tests smushing mode
func TestWithSmushing(t *testing.T) {
	result, err := Render("AB", WithSmushing())
	if err != nil {
		t.Fatalf("Render with WithSmushing failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestWithOverlapping tests overlapping mode
func TestWithOverlapping(t *testing.T) {
	result, err := Render("AB", WithOverlapping())
	if err != nil {
		t.Fatalf("Render with WithOverlapping failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestListFonts tests that ListFonts returns fonts
func TestListFonts(t *testing.T) {
	fonts := ListFonts()
	if len(fonts) == 0 {
		t.Error("ListFonts returned empty list")
	}
	// Check that standard font is in the list
	found := false
	for _, f := range fonts {
		if f == "standard" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListFonts should include 'standard' font")
	}
}

// TestListFontsContainsExpectedFonts tests that all expected fonts are present
func TestListFontsContainsExpectedFonts(t *testing.T) {
	expectedFonts := []string{
		"standard", "banner", "big", "block", "bubble",
		"digital", "ivrit", "lean", "mini", "mnemonic",
		"script", "shadow", "slant", "small", "smscript",
		"smshadow", "smslant", "term",
	}
	fonts := ListFonts()
	fontMap := make(map[string]bool)
	for _, f := range fonts {
		fontMap[f] = true
	}
	for _, expected := range expectedFonts {
		if !fontMap[expected] {
			t.Errorf("Expected font %q not found in ListFonts()", expected)
		}
	}
}

// TestGetVersion tests version functions
func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion returned empty string")
	}
	if version != "2.2.5" {
		t.Errorf("Expected version '2.2.5', got %q", version)
	}
}

// TestGetVersionInt tests version integer function
func TestGetVersionInt(t *testing.T) {
	versionInt := GetVersionInt()
	if versionInt != 20205 {
		t.Errorf("Expected version int 20205, got %d", versionInt)
	}
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	cfg := New()
	if cfg == nil {
		t.Fatal("New() returned nil")
	}
	if cfg.Fontname != "standard" {
		t.Errorf("Default font should be 'standard', got %q", cfg.Fontname)
	}
	if cfg.Outputwidth != DEFAULTCOLUMNS {
		t.Errorf("Default width should be %d, got %d", DEFAULTCOLUMNS, cfg.Outputwidth)
	}
}

// TestConfigLoadFont tests loading fonts with Config
func TestConfigLoadFont(t *testing.T) {
	cfg := New()
	cfg.Fontname = "banner"
	err := cfg.LoadFont()
	if err != nil {
		t.Fatalf("LoadFont failed: %v", err)
	}
}

// TestConfigLoadInvalidFont tests loading invalid font
func TestConfigLoadInvalidFont(t *testing.T) {
	cfg := New()
	cfg.Fontname = "nonexistent_font_12345"
	err := cfg.LoadFont()
	if err == nil {
		t.Error("Expected error for invalid font, got nil")
	}
}

// TestConfigRenderString tests rendering with Config
func TestConfigRenderString(t *testing.T) {
	cfg := New()
	cfg.Fontname = "small"
	err := cfg.LoadFont()
	if err != nil {
		t.Fatalf("LoadFont failed: %v", err)
	}
	result := cfg.RenderString("Test")
	if result == "" {
		t.Error("RenderString returned empty string")
	}
}

// TestConfigMultipleRenders tests multiple renders with same Config
func TestConfigMultipleRenders(t *testing.T) {
	cfg := New()
	err := cfg.LoadFont()
	if err != nil {
		t.Fatalf("LoadFont failed: %v", err)
	}
	
	texts := []string{"A", "B", "Hello", "World"}
	for _, text := range texts {
		result := cfg.RenderString(text)
		if result == "" {
			t.Errorf("RenderString(%q) returned empty string", text)
		}
	}
}

// TestRenderSpecialCharacters tests rendering special characters
func TestRenderSpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"numbers", "12345"},
		{"punctuation", "!@#$%"},
		{"mixed", "Hello, World!"},
		{"symbols", "+-*/="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.input)
			if err != nil {
				t.Fatalf("Render(%q) failed: %v", tt.input, err)
			}
			if result == "" {
				t.Errorf("Render(%q) returned empty string", tt.input)
			}
		})
	}
}

// TestRenderNewlines tests rendering text with newlines
func TestRenderNewlines(t *testing.T) {
	result, err := Render("A\nB")
	if err != nil {
		t.Fatalf("Render with newline failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestRenderLongText tests rendering longer text
func TestRenderLongText(t *testing.T) {
	longText := "This is a longer text that should be properly wrapped based on the output width setting"
	result, err := Render(longText, WithWidth(80))
	if err != nil {
		t.Fatalf("Render long text failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestAllFontsRender tests that all fonts can render without error
func TestAllFontsRender(t *testing.T) {
	fonts := ListFonts()
	for _, font := range fonts {
		t.Run(font, func(t *testing.T) {
			result, err := RenderWithFont("Test", font)
			if err != nil {
				t.Fatalf("Font %q failed to render: %v", font, err)
			}
			if result == "" {
				t.Errorf("Font %q returned empty result", font)
			}
		})
	}
}

// TestCombinedOptions tests combining multiple options
func TestCombinedOptions(t *testing.T) {
	result, err := Render("Go",
		WithFont("slant"),
		WithWidth(60),
		WithJustification(1),
	)
	if err != nil {
		t.Fatalf("Render with combined options failed: %v", err)
	}
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// TestConstants tests that constants are properly defined
func TestConstants(t *testing.T) {
	if DEFAULTCOLUMNS != 80 {
		t.Errorf("DEFAULTCOLUMNS should be 80, got %d", DEFAULTCOLUMNS)
	}
	if VERSION != "2.2.5" {
		t.Errorf("VERSION should be '2.2.5', got %q", VERSION)
	}
	if FONTFILESUFFIX != ".flf" {
		t.Errorf("FONTFILESUFFIX should be '.flf', got %q", FONTFILESUFFIX)
	}
}

// BenchmarkRender benchmarks the Render function
func BenchmarkRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Render("Hello")
	}
}

// BenchmarkRenderWithFont benchmarks rendering with specific font
func BenchmarkRenderWithFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = RenderWithFont("Hello", "standard")
	}
}

// BenchmarkConfigReuse benchmarks reusing Config for multiple renders
func BenchmarkConfigReuse(b *testing.B) {
	cfg := New()
	_ = cfg.LoadFont()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.RenderString("Hello")
	}
}
