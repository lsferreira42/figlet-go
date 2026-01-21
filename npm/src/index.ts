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

// Public API interface
export interface FigletInstance {
    render(text: string): RenderResult;
    renderWithFont(text: string, font: string): RenderResult;
    setFont(font: string): FontResult;
    listFonts(): ListFontsResult;
    getVersion(): string;
    setWidth(width: number): boolean;
    setJustification(align: 'left' | 'center' | 'right' | 'auto'): boolean;
    setColors(colors: string[]): boolean;
    setParser(parser: string): boolean;
    setSmushMode(mode: number): boolean;
    setRightToLeft(mode: number): boolean;
    setParagraph(enabled: boolean): boolean;
    setDeutsch(enabled: boolean): boolean;
    addControlFile(name: string): boolean;
    clearControlFiles(): boolean;
}

// Internal WASM bridge interface
interface WasmFiglet {
    createInstance(): { error: string | null; handle: number };
    render(handle: number, text: string): RenderResult;
    renderWithFont(handle: number, text: string, font: string): RenderResult;
    setFont(handle: number, font: string): FontResult;
    listFonts(): ListFontsResult;
    getVersion(): string;
    setWidth(handle: number, width: number): { success: boolean };
    setJustification(handle: number, align: number): { success: boolean };
    setColors(handle: number, colors: string[]): { success: boolean };
    setParser(handle: number, parser: string): { success: boolean };
    setSmushMode(handle: number, mode: number): { success: boolean };
    setRightToLeft(handle: number, mode: number): { success: boolean };
    setParagraph(handle: number, enabled: boolean): { success: boolean };
    setDeutsch(handle: number, enabled: boolean): { success: boolean };
    addControlFile(handle: number, name: string): { success: boolean };
    clearControlFiles(handle: number): { success: boolean };
}

declare global {
    var figlet: WasmFiglet | undefined;
    var Go: any;
}

let instance: WasmFiglet | null = null;
let initPromise: Promise<WasmFiglet> | null = null;

/**
 * Initialize the FIGlet WASM module
 */
export async function init(wasmPath?: string): Promise<WasmFiglet> {
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
    const wasm = await init();
    const result = wasm.render(0, text);
    if (result.error) {
        throw new Error(result.error);
    }
    return result.result;
}

/**
 * Render text with a specific font
 */
export async function renderWithFont(text: string, font: string): Promise<string> {
    const wasm = await init();
    const result = wasm.renderWithFont(0, text, font);
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

class FIGlet implements FigletInstance {
    private handle: number;
    private wasm: WasmFiglet;

    constructor(wasm: WasmFiglet, handle: number) {
        this.wasm = wasm;
        this.handle = handle;
    }

    render(text: string): RenderResult {
        return this.wasm.render(this.handle, text);
    }

    renderWithFont(text: string, font: string): RenderResult {
        return this.wasm.renderWithFont(this.handle, text, font);
    }

    setFont(font: string): FontResult {
        return this.wasm.setFont(this.handle, font);
    }

    listFonts(): ListFontsResult {
        return this.wasm.listFonts();
    }

    getVersion(): string {
        return this.wasm.getVersion();
    }

    setWidth(width: number): boolean {
        const result = this.wasm.setWidth(this.handle, width);
        return result.success;
    }

    setJustification(align: 'left' | 'center' | 'right' | 'auto'): boolean {
        const alignMap = { auto: -1, left: 0, center: 1, right: 2 };
        const result = this.wasm.setJustification(this.handle, alignMap[align]);
        return result.success;
    }

    setColors(colors: string[]): boolean {
        const result = this.wasm.setColors(this.handle, colors);
        return result.success;
    }

    setParser(parser: string): boolean {
        const result = this.wasm.setParser(this.handle, parser);
        return result.success;
    }

    setSmushMode(mode: number): boolean {
        const result = this.wasm.setSmushMode(this.handle, mode);
        return result.success;
    }

    setRightToLeft(mode: number): boolean {
        const result = this.wasm.setRightToLeft(this.handle, mode);
        return result.success;
    }

    setParagraph(enabled: boolean): boolean {
        const result = this.wasm.setParagraph(this.handle, enabled);
        return result.success;
    }

    setDeutsch(enabled: boolean): boolean {
        const result = this.wasm.setDeutsch(this.handle, enabled);
        return result.success;
    }

    addControlFile(name: string): boolean {
        const result = this.wasm.addControlFile(this.handle, name);
        return result.success;
    }

    clearControlFiles(): boolean {
        const result = this.wasm.clearControlFiles(this.handle);
        return result.success;
    }
}

/**
 * Create a configured FIGlet instance
 */
export async function createInstance(options?: {
    font?: string;
    width?: number;
    justification?: 'left' | 'center' | 'right' | 'auto';
    colors?: string[];
    parser?: string;
    smushMode?: number;
    rightToLeft?: number;
    paragraph?: boolean;
    deutsch?: boolean;
}): Promise<FigletInstance> {
    const wasm = await init();
    const result = wasm.createInstance();

    if (result.error) {
        throw new Error(result.error);
    }

    const fig = new FIGlet(wasm, result.handle);

    if (options?.font) {
        const res = fig.setFont(options.font);
        if (res.error) {
            throw new Error(res.error);
        }
    }

    if (options?.width) {
        fig.setWidth(options.width);
    }

    if (options?.justification) {
        fig.setJustification(options.justification);
    }

    if (options?.colors) {
        fig.setColors(options.colors);
    }

    if (options?.parser) {
        fig.setParser(options.parser);
    }

    if (options?.smushMode !== undefined) {
        fig.setSmushMode(options.smushMode);
    }

    if (options?.rightToLeft !== undefined) {
        fig.setRightToLeft(options.rightToLeft);
    }

    if (options?.paragraph !== undefined) {
        fig.setParagraph(options.paragraph);
    }

    if (options?.deutsch !== undefined) {
        fig.setDeutsch(options.deutsch);
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
