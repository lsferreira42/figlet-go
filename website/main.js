// FIGlet-Go Playground - Main JavaScript

const state = {
    ready: false,
    currentFont: 'standard',
    fonts: [],
    debounceTimer: null,
};

// DOM Elements
const elements = {
    textInput: document.getElementById('text-input'),
    fontSelect: document.getElementById('font-select'),
    widthInput: document.getElementById('width-input'),
    alignSelect: document.getElementById('align-select'),
    output: document.getElementById('output'),
    copyBtn: document.getElementById('copy-btn'),
    downloadBtn: document.getElementById('download-btn'),
    version: document.getElementById('version'),
    fontGallery: document.getElementById('font-gallery'),
    toast: document.getElementById('toast'),
};

// Initialize the WASM module
async function initWasm() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch('figlet.wasm'),
            go.importObject
        );
        go.run(result.instance);
        
        // Wait for figlet to be ready
        await waitForFiglet();
        
        state.ready = true;
        onFigletReady();
    } catch (error) {
        console.error('Failed to load WASM:', error);
        elements.output.textContent = `Error loading FIGlet: ${error.message}\n\nMake sure figlet.wasm is available.`;
        showToast('Failed to load FIGlet WASM module', 'error');
    }
}

// Wait for the figlet global to be available
function waitForFiglet() {
    return new Promise((resolve) => {
        if (typeof figlet !== 'undefined') {
            resolve();
        } else {
            document.addEventListener('figlet-ready', resolve, { once: true });
        }
    });
}

// Called when FIGlet is ready
function onFigletReady() {
    // Set version
    elements.version.textContent = `v${figlet.getVersion()}`;
    
    // Load fonts
    loadFonts();
    
    // Setup event listeners
    setupEventListeners();
    
    // Initial render
    render();
    
    showToast('FIGlet-Go loaded successfully!', 'success');
}

// Load available fonts
function loadFonts() {
    const result = figlet.listFonts();
    
    if (result.error) {
        console.error('Failed to load fonts:', result.error);
        return;
    }
    
    state.fonts = result.fonts.sort();
    
    // Populate font select
    elements.fontSelect.innerHTML = state.fonts
        .map(font => `<option value="${font}" ${font === 'standard' ? 'selected' : ''}>${font}</option>`)
        .join('');
    
    // Generate font gallery
    generateFontGallery();
}

// Generate font gallery with previews
function generateFontGallery() {
    const previewText = 'Abc';
    const fragment = document.createDocumentFragment();
    
    state.fonts.forEach(font => {
        const card = document.createElement('div');
        card.className = `font-card ${font === state.currentFont ? 'active' : ''}`;
        card.dataset.font = font;
        
        const result = figlet.renderWithFont(previewText, font);
        const preview = result.error ? 'Error loading font' : result.result;
        
        card.innerHTML = `
            <div class="font-card-header">
                <span class="font-card-name">${font}</span>
            </div>
            <div class="font-card-preview">
                <pre>${escapeHtml(preview)}</pre>
            </div>
        `;
        
        card.addEventListener('click', () => selectFont(font));
        fragment.appendChild(card);
    });
    
    elements.fontGallery.innerHTML = '';
    elements.fontGallery.appendChild(fragment);
}

// Setup event listeners
function setupEventListeners() {
    // Text input - debounced
    elements.textInput.addEventListener('input', () => {
        clearTimeout(state.debounceTimer);
        state.debounceTimer = setTimeout(render, 100);
    });
    
    // Font select
    elements.fontSelect.addEventListener('change', (e) => {
        selectFont(e.target.value);
    });
    
    // Width input
    elements.widthInput.addEventListener('change', () => {
        const width = parseInt(elements.widthInput.value) || 80;
        figlet.setWidth(Math.max(20, Math.min(200, width)));
        render();
    });
    
    // Alignment select
    elements.alignSelect.addEventListener('change', () => {
        figlet.setJustification(parseInt(elements.alignSelect.value));
        render();
    });
    
    // Copy button
    elements.copyBtn.addEventListener('click', copyToClipboard);
    
    // Download button
    elements.downloadBtn.addEventListener('click', downloadText);
    
    // Enter key to render
    elements.textInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            clearTimeout(state.debounceTimer);
            render();
        }
    });
}

// Select a font
function selectFont(fontName) {
    state.currentFont = fontName;
    
    // Update select dropdown
    elements.fontSelect.value = fontName;
    
    // Update active state in gallery
    document.querySelectorAll('.font-card').forEach(card => {
        card.classList.toggle('active', card.dataset.font === fontName);
    });
    
    // Set font and render
    const result = figlet.setFont(fontName);
    if (result.error) {
        showToast(`Error loading font: ${result.error}`, 'error');
        return;
    }
    
    render();
}

// Render the text
function render() {
    if (!state.ready) return;
    
    const text = elements.textInput.value || 'Hello';
    const result = figlet.render(text);
    
    if (result.error) {
        elements.output.textContent = `Error: ${result.error}`;
        return;
    }
    
    elements.output.textContent = result.result || '(empty output)';
}

// Copy output to clipboard
async function copyToClipboard() {
    const text = elements.output.textContent;
    
    try {
        await navigator.clipboard.writeText(text);
        elements.copyBtn.classList.add('success');
        showToast('Copied to clipboard!', 'success');
        
        setTimeout(() => {
            elements.copyBtn.classList.remove('success');
        }, 2000);
    } catch (error) {
        // Fallback for older browsers
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        showToast('Copied to clipboard!', 'success');
    }
}

// Download as text file
function downloadText() {
    const text = elements.output.textContent;
    const blob = new Blob([text], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `figlet-${state.currentFont}-${Date.now()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    showToast('Downloaded!', 'success');
}

// Show toast notification
function showToast(message, type = 'info') {
    elements.toast.textContent = message;
    elements.toast.className = `toast ${type}`;
    
    // Trigger reflow for animation
    elements.toast.offsetHeight;
    elements.toast.classList.add('visible');
    
    setTimeout(() => {
        elements.toast.classList.remove('visible');
    }, 3000);
}

// Escape HTML entities
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initWasm);
