# figlet-go

> FIGlet ASCII art text generator - WebAssembly module compiled from Go

[![npm version](https://badge.fury.io/js/figlet-go.svg)](https://badge.fury.io/js/figlet-go)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

Generate beautiful ASCII art text from your JavaScript/Node.js applications using the power of WebAssembly.

## Installation

```bash
npm install figlet-go
```

## Quick Start

```javascript
const figlet = require('figlet-go');

// Simple rendering
const art = await figlet.render('Hello!');
console.log(art);

// Output:
//  _   _      _ _       _ 
// | | | | ___| | | ___ | |
// | |_| |/ _ \ | |/ _ \| |
// |  _  |  __/ | | (_) |_|
// |_| |_|\___|_|_|\___/(_)
```

## API Reference

### `render(text: string): Promise<string>`

Render text using the default font (standard).

```javascript
const art = await figlet.render('Hello');
```

### `renderWithFont(text: string, font: string): Promise<string>`

Render text with a specific font.

```javascript
const art = await figlet.renderWithFont('Hello', 'slant');
```

### `listFonts(): Promise<string[]>`

Get a list of all available fonts.

```javascript
const fonts = await figlet.listFonts();
console.log(fonts); // ['banner', 'big', 'block', ...]
```

### `getVersion(): Promise<string>`

Get the FIGlet version.

```javascript
const version = await figlet.getVersion();
console.log(version); // '2.2.5'
```

### `createInstance(options?): Promise<FigletInstance>`

Create a configured FIGlet instance for more control.

```javascript
const fig = await figlet.createInstance({
    font: 'slant',
    width: 100,
    justification: 'center'
});

const result = fig.render('Hello');
console.log(result.result);
```

#### Options

| Option | Type | Description |
|--------|------|-------------|
| `font` | `string` | Font name to use |
| `width` | `number` | Output width (default: 80) |
| `justification` | `'left' \| 'center' \| 'right' \| 'auto'` | Text alignment |

## Available Fonts

The package includes **146 built-in fonts** from the [FIGlet font database](http://www.figlet.org/fontdb.cgi):

Popular fonts: `standard`, `banner`, `big`, `block`, `slant`, `shadow`, `script`, `small`, `doom`, `graffiti`, `starwars`, `larry3d`, `colossal`, `gothic`, `epic`, `poison`, `roman`, `rounded`, `speed`, `stellar`, and many more!

Use `listFonts()` to see all available fonts.

## Browser Usage

```html
<script src="https://unpkg.com/figlet-go/dist/wasm_exec.js"></script>
<script type="module">
import figlet from 'https://unpkg.com/figlet-go/dist/index.mjs';

// Initialize with WASM path
await figlet.init('https://unpkg.com/figlet-go/dist/figlet.wasm');

const art = await figlet.render('Hello Web!');
console.log(art);
</script>
```

## TypeScript Support

Full TypeScript definitions are included:

```typescript
import figlet, { FigletInstance } from 'figlet-go';

const art: string = await figlet.render('Hello');
const fonts: string[] = await figlet.listFonts();

const instance: FigletInstance = await figlet.createInstance({
    font: 'slant',
    justification: 'center'
});
```

## Examples

### Generate multiple styles

```javascript
const figlet = require('figlet-go');

const fonts = ['standard', 'slant', 'banner', 'big'];

for (const font of fonts) {
    console.log(`\n--- ${font} ---\n`);
    const art = await figlet.renderWithFont('Hello', font);
    console.log(art);
}
```

### Custom width and alignment

```javascript
const fig = await figlet.createInstance({
    font: 'small',
    width: 60,
    justification: 'center'
});

const result = fig.render('Centered');
console.log(result.result);
```

## Performance

This package uses WebAssembly compiled from Go, providing:
- **Fast rendering**: Native-speed text generation
- **146 fonts**: All fonts from figlet.org embedded
- **Consistent output**: Same results across all platforms

## Links

- [GitHub Repository](https://github.com/lsferreira42/figlet-go)
- [Online Playground](https://lsferreira42.github.io/figlet-go/)
- [Go Library Documentation](https://pkg.go.dev/github.com/lsferreira42/figlet-go/figlet)

## License

BSD-3-Clause © Leandro Ferreira

Based on FIGlet by Glenn Chappell, Ian Chai, John Cowan, Christiaan Keet, and Claudio Matsuoka.
