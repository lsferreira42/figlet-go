/**
 * FIGlet-Go - ASCII art text generator
 * WebAssembly module compiled from Go
 */

import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

// Type definitions
export interface RenderResult {
    error: string | null;
    result: string;
}

export interface FontResult {
    error: string | null;
    success: boolean;
}

export interface ListFontsResult {
    error: string | null;
    fonts: string[];
}

export interface FigletInstance {
    render(text: string): RenderResult;
    renderWithFont(text: string, font: string): RenderResult;
    setFont(font: string): FontResult;
    listFonts(): ListFontsResult;
    getVersion(): string;
    setWidth(width: number): boolean;
    setJustification(align: number): boolean;
}

declare global {
    var figlet: FigletInstance | undefined;
    var Go: any;
}

let instance: FigletInstance | null = null;
let initPromise: Promise<FigletInstance> | null = null;

/**
 * Initialize the FIGlet WASM module
 */
export async function init(wasmPath?: string): Promise<FigletInstance> {
    if (instance) {
        return instance;
    }

    if (initPromise) {
        return initPromise;
    }

    initPromise = (async () => {
        // Determine WASM file path
        let wasmBuffer: ArrayBuffer;

        if (typeof window !== 'undefined') {
            // Browser environment
            const path = wasmPath || 'figlet.wasm';
            const response = await fetch(path);
            wasmBuffer = await response.arrayBuffer();
        } else {
            // Node.js environment
            const path = wasmPath || join(dirname(fileURLToPath(import.meta.url)), 'figlet.wasm');
            const buffer = readFileSync(path);
            wasmBuffer = buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength);
        }

        // Load Go WASM support
        if (typeof Go === 'undefined') {
            // In Node.js, we need to load wasm_exec.js
            if (typeof window === 'undefined') {
                const wasmExecPath = join(dirname(fileURLToPath(import.meta.url)), 'wasm_exec.js');
                await import(wasmExecPath);
            }
        }

        const go = new Go();
        const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
        go.run(result.instance);

        // Wait for figlet to be available
        await new Promise<void>((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Timeout waiting for FIGlet to initialize'));
            }, 5000);

            const check = () => {
                if (typeof figlet !== 'undefined') {
                    clearTimeout(timeout);
                    resolve();
                } else {
                    setTimeout(check, 10);
                }
            };
            check();
        });

        instance = figlet!;
        return instance;
    })();

    return initPromise;
}

/**
 * Render text with the default font
 */
export async function render(text: string): Promise<string> {
    const fig = await init();
    const result = fig.render(text);
    if (result.error) {
        throw new Error(result.error);
    }
    return result.result;
}

/**
 * Render text with a specific font
 */
export async function renderWithFont(text: string, font: string): Promise<string> {
    const fig = await init();
    const result = fig.renderWithFont(text, font);
    if (result.error) {
        throw new Error(result.error);
    }
    return result.result;
}

/**
 * List available fonts
 */
export async function listFonts(): Promise<string[]> {
    const fig = await init();
    const result = fig.listFonts();
    if (result.error) {
        throw new Error(result.error);
    }
    return result.fonts;
}

/**
 * Get the FIGlet version
 */
export async function getVersion(): Promise<string> {
    const fig = await init();
    return fig.getVersion();
}

/**
 * Create a configured FIGlet instance
 */
export async function createInstance(options?: {
    font?: string;
    width?: number;
    justification?: 'left' | 'center' | 'right' | 'auto';
}): Promise<FigletInstance> {
    const fig = await init();
    
    if (options?.font) {
        const result = fig.setFont(options.font);
        if (result.error) {
            throw new Error(result.error);
        }
    }
    
    if (options?.width) {
        fig.setWidth(options.width);
    }
    
    if (options?.justification) {
        const alignMap = { auto: -1, left: 0, center: 1, right: 2 };
        fig.setJustification(alignMap[options.justification]);
    }
    
    return fig;
}

// Default export
export default {
    init,
    render,
    renderWithFont,
    listFonts,
    getVersion,
    createInstance,
};
