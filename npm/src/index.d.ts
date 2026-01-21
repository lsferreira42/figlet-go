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
    render(text: string): RenderResult;
    renderWithFont(text: string, font: string): RenderResult;
    setFont(font: string): FontResult;
    listFonts(): ListFontsResult;
    getVersion(): string;
    setWidth(width: number): boolean;
    setJustification(align: 'left' | 'center' | 'right' | 'auto'): boolean;
    setColors(colors: string[]): boolean;
    /**
     * Set output parser
     * @param parser Parser name (terminal, terminal-color, html)
     */
    setParser(parser: string): boolean;

    /**
     * Set smush mode
     * @param mode Smush mode (0=kerning, -1=full width, 1+=smushing)
     */
    setSmushMode(mode: number): boolean;

    /**
     * Set right-to-left mode
     * @param mode 0 = left, 1 = right, -1 = auto
     */
    setRightToLeft(mode: number): boolean;

    /**
     * Enable/disable paragraph mode
     */
    setParagraph(enabled: boolean): boolean;

    /**
     * Enable/disable deutsch flag
     */
    setDeutsch(enabled: boolean): boolean;

    /**
     * Add a control file
     */
    addControlFile(name: string): boolean;

    /**
     * Clear all control files
     */
    clearControlFiles(): boolean;
}

export interface CreateInstanceOptions {
    /** Font name to use */
    font?: string;
    /** Output width */
    width?: number;
    /** Text justification */
    justification?: 'left' | 'center' | 'right' | 'auto';
    /** Output colors */
    colors?: string[];
    /** Output parser */
    parser?: string;
    /** Smush mode */
    smushMode?: number;
    /** Right-to-left mode */
    rightToLeft?: number;
    /** Paragraph mode */
    paragraph?: boolean;
    /** Deutsch flag */
    deutsch?: boolean;
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
