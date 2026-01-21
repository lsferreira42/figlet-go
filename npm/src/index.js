/**
 * FIGlet-Go - ASCII art text generator
 * WebAssembly module compiled from Go
 */

const fs = require('fs');
const path = require('path');

let instance = null;
let initPromise = null;
let goInstance = null;

// Load wasm_exec.js which defines the Go class
const wasmExecPath = path.join(__dirname, '..', 'dist', 'wasm_exec.js');
if (typeof Go === 'undefined' && fs.existsSync(wasmExecPath)) {
    require(wasmExecPath);
}

/**
 * Initialize the FIGlet WASM module
 * @param {string} [wasmPath] - Optional path to the WASM file
 * @returns {Promise<object>} - The FIGlet instance
 */
async function init(wasmPath) {
    if (instance) {
        return instance;
    }

    if (initPromise) {
        return initPromise;
    }

    initPromise = (async () => {
        // Determine WASM file path
        const defaultPath = path.join(__dirname, '..', 'dist', 'figlet.wasm');
        const finalPath = wasmPath || defaultPath;

        // Read WASM file
        const wasmBuffer = fs.readFileSync(finalPath);

        // Create Go instance and run WASM
        goInstance = new Go();
        const result = await WebAssembly.instantiate(wasmBuffer, goInstance.importObject);
        goInstance.run(result.instance);

        // Wait for figlet to be available
        await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Timeout waiting for FIGlet to initialize'));
            }, 5000);

            const check = () => {
                if (typeof global.figlet !== 'undefined') {
                    clearTimeout(timeout);
                    resolve();
                } else {
                    setTimeout(check, 10);
                }
            };
            check();
        });

        instance = global.figlet;
        return instance;
    })();

    return initPromise;
}

/**
 * Render text with the default font
 * @param {string} text - The text to render
 * @returns {Promise<string>} - The rendered ASCII art
 */
async function render(text) {
    const fig = await init();
    const result = fig.render(text);
    if (result.error) {
        throw new Error(result.error);
    }
    return result.result;
}

/**
 * Render text with a specific font
 * @param {string} text - The text to render
 * @param {string} font - The font name to use
 * @returns {Promise<string>} - The rendered ASCII art
 */
async function renderWithFont(text, font) {
    const fig = await init();
    const result = fig.renderWithFont(text, font);
    if (result.error) {
        throw new Error(result.error);
    }
    return result.result;
}

/**
 * List available fonts
 * @returns {Promise<string[]>} - Array of font names
 */
async function listFonts() {
    const fig = await init();
    const result = fig.listFonts();
    if (result.error) {
        throw new Error(result.error);
    }
    return result.fonts;
}

/**
 * Get the FIGlet version
 * @returns {Promise<string>} - Version string
 */
async function getVersion() {
    const fig = await init();
    return fig.getVersion();
}

class FIGlet {
    constructor(wasm, handle) {
        this.wasm = wasm;
        this.handle = handle;
    }

    render(text) {
        return this.wasm.render(this.handle, text);
    }

    renderWithFont(text, font) {
        return this.wasm.renderWithFont(this.handle, text, font);
    }

    setFont(font) {
        return this.wasm.setFont(this.handle, font);
    }

    listFonts() {
        return this.wasm.listFonts();
    }

    getVersion() {
        return this.wasm.getVersion();
    }

    setWidth(width) {
        const result = this.wasm.setWidth(this.handle, width);
        return result.success;
    }

    setJustification(align) {
        const alignMap = { auto: -1, left: 0, center: 1, right: 2 };
        const result = this.wasm.setJustification(this.handle, alignMap[align]);
        return result.success;
    }

    setColors(colors) {
        const result = this.wasm.setColors(this.handle, colors);
        return result.success;
    }

    setParser(parser) {
        const result = this.wasm.setParser(this.handle, parser);
        return result.success;
    }

    setSmushMode(mode) {
        const result = this.wasm.setSmushMode(this.handle, mode);
        return result.success;
    }

    setRightToLeft(mode) {
        const result = this.wasm.setRightToLeft(this.handle, mode);
        return result.success;
    }

    setParagraph(enabled) {
        const result = this.wasm.setParagraph(this.handle, enabled);
        return result.success;
    }

    setDeutsch(enabled) {
        const result = this.wasm.setDeutsch(this.handle, enabled);
        return result.success;
    }

    addControlFile(name) {
        const result = this.wasm.addControlFile(this.handle, name);
        return result.success;
    }

    clearControlFiles() {
        const result = this.wasm.clearControlFiles(this.handle);
        return result.success;
    }
}

/**
 * Create a configured FIGlet instance
 * @param {object} [options] - Configuration options
 * @param {string} [options.font] - Font name
 * @param {number} [options.width] - Output width
 * @param {'left'|'center'|'right'|'auto'} [options.justification] - Text alignment
 * @param {string[]} [options.colors] - Array of color names or hex strings
 * @param {string} [options.parser] - Parser name (terminal, terminal-color, html)
 * @param {number} [options.smushMode] - Smush mode (0=kerning, -1=full width, 1+=smushing)
 * @param {number} [options.rightToLeft] - Right-to-left mode (0=left, 1=right, -1=auto)
 * @param {boolean} [options.paragraph] - Paragraph mode
 * @param {boolean} [options.deutsch] - Deutsch mode
 * @returns {Promise<object>} - Configured FIGlet instance
 */
async function createInstance(options = {}) {
    const wasm = await init();
    const result = wasm.createInstance();

    if (result.error) {
        throw new Error(result.error);
    }

    const fig = new FIGlet(wasm, result.handle);

    if (options.font) {
        const res = fig.setFont(options.font);
        if (res.error) {
            throw new Error(res.error);
        }
    }

    if (options.width) {
        fig.setWidth(options.width);
    }

    if (options.justification) {
        fig.setJustification(options.justification);
    }

    if (options.colors) {
        fig.setColors(options.colors);
    }

    if (options.parser) {
        fig.setParser(options.parser);
    }

    if (options.smushMode !== undefined) {
        fig.setSmushMode(options.smushMode);
    }

    if (options.rightToLeft !== undefined) {
        fig.setRightToLeft(options.rightToLeft);
    }

    if (options.paragraph !== undefined) {
        fig.setParagraph(options.paragraph);
    }

    if (options.deutsch !== undefined) {
        fig.setDeutsch(options.deutsch);
    }

    return fig;
}

module.exports = {
    init,
    render,
    renderWithFont,
    listFonts,
    getVersion,
    createInstance,
};
