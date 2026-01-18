# Colors and Output Formats

FIGlet-Go supports colors (ANSI and TrueColor) and multiple output formats (terminal, terminal with colors, and HTML).

## Colors

### ANSI Colors

FIGlet-Go supports 8 basic ANSI colors:

- `black`
- `red`
- `green`
- `yellow`
- `blue`
- `magenta`
- `cyan`
- `white`

### TrueColor (24-bit RGB)

You can also use TrueColor by specifying hexadecimal color codes (with or without the `#` prefix):

- `FF0000` (red)
- `00FF00` (green)
- `0000FF` (blue)
- `FF00FF` (magenta)
- Any other 6-digit hex code

### Using Colors

Colors are specified using the `--colors` option, with multiple colors separated by semicolons (`;`):

```bash
# Single ANSI color
./figlet-bin --colors red "Hello"

# Multiple ANSI colors (will cycle through them)
./figlet-bin --colors 'red;green;blue' "Colors"

# TrueColor (hex codes)
./figlet-bin --colors 'FF0000;00FF00;0000FF' "TrueColor"

# Mixed ANSI and TrueColor
./figlet-bin --colors 'red;00FF00;blue' "Mixed"

# With # prefix for hex colors
./figlet-bin --colors '#FF0000' "Red"
```

**Note:** Colors cycle through each character of the rendered text. If you specify 3 colors, they will be applied to characters in a repeating pattern: color1, color2, color3, color1, color2, color3, etc.

## Output Formats (Parsers)

FIGlet-Go supports three output formats:

### 1. Terminal (Default)

Plain text output without any formatting codes. This is the default behavior.

```bash
# Default (terminal parser)
./figlet-bin "Hello"

# Explicit terminal parser
./figlet-bin --parser terminal "Hello"
```

### 2. Terminal with Colors

Output with ANSI color codes for terminal display. This parser is automatically selected when you use `--colors` (unless you explicitly set a different parser).

```bash
# With colors (automatically uses terminal-color parser)
./figlet-bin --colors 'red;green;blue' "Hello"

# Explicit terminal-color parser
./figlet-bin --parser terminal-color --colors 'red;green;blue' "Hello"

# Terminal-color parser without colors (works, but no colors applied)
./figlet-bin --parser terminal-color "Hello"
```

### 3. HTML

Output formatted as HTML with `<code>` tags and HTML entities. Colors are rendered as `<span>` tags with inline styles.

```bash
# HTML output without colors
./figlet-bin --parser html "Hello"

# HTML output with colors
./figlet-bin --parser html --colors 'red;green;blue' "Hello"

# HTML output with TrueColor
./figlet-bin --parser html --colors 'FF0000;00FF00;0000FF' "Hello"
```

**HTML Output Features:**
- Text is wrapped in `<code>` tags
- Spaces are converted to `&nbsp;`
- Newlines are converted to `<br>`
- Colors are rendered as `<span style='color: rgb(r,g,b);'>` tags

## Combining Options

You can combine colors and parsers with other FIGlet options:

```bash
# Different font with colors
./figlet-bin -f banner --colors 'red;green;blue' "Hello"

# Centered text with HTML output
./figlet-bin -c --parser html --colors red "Centered"

# Custom width with colored output
./figlet-bin -w 100 --parser terminal-color --colors 'FF0000;00FF00' "Wide"

# Slant font with HTML and colors
./figlet-bin -f slant --parser html --colors 'red;blue' "Slant"
```

## Examples

### Example 1: Rainbow Text

```bash
./figlet-bin --parser terminal-color --colors 'red;yellow;green;cyan;blue;magenta' "Rainbow"
```

### Example 2: HTML Output for Web

```bash
./figlet-bin --parser html --colors 'FF0000;00FF00;0000FF' "Web Ready" > output.html
```

### Example 3: TrueColor Gradient

```bash
./figlet-bin --parser terminal-color --colors 'FF0000;FF3300;FF6600;FF9900;FFCC00;FFFF00' "Gradient"
```

### Example 4: Simple Colored Banner

```bash
./figlet-bin -f banner --colors red "WARNING"
```

## Color Cycling

When you specify multiple colors, they are applied in a cycling pattern to each character position in the rendered output. For example, with `--colors 'red;green;blue'`:

- Character 1 → red
- Character 2 → green
- Character 3 → blue
- Character 4 → red (cycles back)
- Character 5 → green
- And so on...

This applies to each character position in the ASCII art, not to each input character, so long characters may have multiple positions with the same color.

## Invalid Colors

If you specify an invalid color name or hex code, FIGlet-Go will handle it gracefully:

- Invalid color names are ignored
- Invalid hex codes are ignored
- The rendering continues without colors for invalid entries
- No error is thrown (the command succeeds)

```bash
# Invalid color - will render without colors
./figlet-bin --parser terminal-color --colors invalid "Hello"

# Mix of valid and invalid - valid colors are used
./figlet-bin --parser terminal-color --colors 'red;invalid;blue' "Hello"
```

## Parser Selection

The parser selection works as follows:

1. If you specify `--parser`, that parser is used
2. If you specify `--colors` without `--parser`, the parser defaults to `terminal-color`
3. If you specify both `--parser html` and `--colors`, HTML output is used with colored spans
4. If you specify `--parser terminal` and `--colors`, colors are ignored (terminal parser doesn't support colors)

## Tips

1. **For terminal display**: Use `--parser terminal-color` (or just `--colors`) for colored output
2. **For web/HTML**: Use `--parser html` with or without colors
3. **For plain text**: Use `--parser terminal` or no parser option
4. **TrueColor**: Works best with terminals that support 24-bit color (most modern terminals)
5. **ANSI colors**: Work on virtually all terminals, including older ones

## See Also

- [Library Documentation](lib.md) - Using colors and parsers in Go code
- [Main README](README.md) - General FIGlet-Go documentation
