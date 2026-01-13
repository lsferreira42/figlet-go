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

/**
 * Create a configured FIGlet instance
 * @param {object} [options] - Configuration options
 * @param {string} [options.font] - Font name
 * @param {number} [options.width] - Output width
 * @param {'left'|'center'|'right'|'auto'} [options.justification] - Text alignment
 * @returns {Promise<object>} - Configured FIGlet instance
 */
async function createInstance(options = {}) {
    const fig = await init();
    
    if (options.font) {
        const result = fig.setFont(options.font);
        if (result.error) {
            throw new Error(result.error);
        }
    }
    
    if (options.width) {
        fig.setWidth(options.width);
    }
    
    if (options.justification) {
        const alignMap = { auto: -1, left: 0, center: 1, right: 2 };
        fig.setJustification(alignMap[options.justification]);
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
