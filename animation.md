# FIGlet Animations

This document explains the animation features in `figlet-go`.

## Animation Options

You can animate your FIGlet banners using the `--animation` flag.

### `--animation $type`

Supported animation types:

- **reveal**: Reveals the text character by character, simulating a typing effect.
- **scroll**: Scrolls the text from right to left across the terminal.
- **rain**: Each row of the FIGlet characters "falls" from the top of the banner to its final position.
- **wave**: A vertical wave effect that moves across the rendered text.
- **explosion**: Animates the characters flying outwards from their original positions.

Example:
```bash
figlet-go --animation reveal "Hello World"
```

### `--animation-delay $ms`

Sets the delay between frames in milliseconds. Default is 50ms.

Example:
```bash
figlet-go --animation scroll --animation-delay 100 "Slow Scroll"
```

## HTML Output (Browser-Ready)

When using the `--parser html` flag with animations, `figlet-go` generates a **standalone HTML animation player**.

### Polished Terminal Aesthetic
The generated HTML features:
- **Left-aligned layout**: Perfectly matches terminal behavior.
- **Dark theme**: Solid terminal-black background (`#0c0c0c`).
- **High-performance JS engine**: A lightweight player for fluid, flicker-free playback.
- **Optimized fonts**: Uses a professional monospaced stack (Cascadia Code, Ubuntu Mono, etc.).

Example:
```bash
figlet-go --parser html --animation wave --colors "red;green;blue" "FIGlet Wave" > animation.html
```
Open `animation.html` in any browser to view the animated result.

## Stable Color Mapping

All animations support **stable color mapping**. This means characters maintain their assigned colors as they move, rather than having colors fixed to terminal positions. This ensures high-fidelity gradients and patterns even during complex "explosion" or "wave" effects.

## Exporting and Playing Animations

You can export an animation to a file and play it back later.

### `--export $file`

Saves the animation frames to a file. The export includes all ANSI escape codes for colors and character positioning.

Example:
```bash
figlet-go --animation rain "Rainy Day" --export rain.ani
```

### `--animation-file $file`

Plays back an exported animation file.

Example:
```bash
figlet-go --animation-file rain.ani
```

> [!NOTE]
> Exported animations can also be viewed using standard terminal tools like `cat` if they contain enough frames and appropriate escape codes, but using `figlet-go --animation-file` ensures the correct timing.
坐
