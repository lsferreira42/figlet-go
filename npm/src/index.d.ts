/**
 * FIGlet-Go - ASCII art text generator
 * WebAssembly module compiled from Go
 */

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
    /**
     * Render text with the current font
     */
    render(text: string): RenderResult;
    
    /**
     * Render text with a specific font
     */
    renderWithFont(text: string, font: string): RenderResult;
    
    /**
     * Set the current font
     */
    setFont(font: string): FontResult;
    
    /**
     * List available fonts
     */
    listFonts(): ListFontsResult;
    
    /**
     * Get the FIGlet version
     */
    getVersion(): string;
    
    /**
     * Set the output width
     */
    setWidth(width: number): boolean;
    
    /**
     * Set text justification
     * @param align -1 = auto, 0 = left, 1 = center, 2 = right
     */
    setJustification(align: number): boolean;
}

export interface CreateInstanceOptions {
    /** Font name to use */
    font?: string;
    /** Output width */
    width?: number;
    /** Text justification */
    justification?: 'left' | 'center' | 'right' | 'auto';
}

/**
 * Initialize the FIGlet WASM module
 * @param wasmPath - Optional path to the WASM file
 */
export function init(wasmPath?: string): Promise<FigletInstance>;

/**
 * Render text with the default font
 * @param text - The text to render
 */
export function render(text: string): Promise<string>;

/**
 * Render text with a specific font
 * @param text - The text to render
 * @param font - The font name to use
 */
export function renderWithFont(text: string, font: string): Promise<string>;

/**
 * List available fonts
 */
export function listFonts(): Promise<string[]>;

/**
 * Get the FIGlet version
 */
export function getVersion(): Promise<string>;

/**
 * Create a configured FIGlet instance
 * @param options - Configuration options
 */
export function createInstance(options?: CreateInstanceOptions): Promise<FigletInstance>;

declare const figletGo: {
    init: typeof init;
    render: typeof render;
    renderWithFont: typeof renderWithFont;
    listFonts: typeof listFonts;
    getVersion: typeof getVersion;
    createInstance: typeof createInstance;
};

export default figletGo;
