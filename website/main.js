// FIGlet-Go Playground - Main JavaScript

const state = {
    ready: false,
    currentFont: 'banner3-D',
    fonts: [],
    debounceTimer: null,
};

// DOM Elements
const elements = {
    textInput: document.getElementById('text-input'),
    fontSelect: document.getElementById('font-select'),
    widthInput: document.getElementById('width-input'),
    alignSelect: document.getElementById('align-select'),
    parserSelect: document.getElementById('parser-select'),
    colorsInput: document.getElementById('colors-input'),
    output: document.getElementById('output'),
    copyBtn: document.getElementById('copy-btn'),
    downloadBtn: document.getElementById('download-btn'),
    fontGallery: document.getElementById('font-gallery'),
    toast: document.getElementById('toast'),
    animationSelect: document.getElementById('animation-select'),
    animationDelay: document.getElementById('animation-delay'),
    playBtn: document.getElementById('play-btn'),
};

const animation = {
    playing: false,
    frames: [],
    currentFrame: 0,
    timer: null,
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
    // Load fonts
    loadFonts();

    // Setup initial WASM state based on defaults
    const colors = elements.colorsInput.value.split(';').map(c => c.trim()).filter(c => c);
    if (colors.length > 0) {
        figlet.setColors(colors);
        figlet.setParser('html');
        // Update UI state for parser select
        elements.parserSelect.disabled = true;
        elements.parserSelect.style.opacity = '0.6';
        elements.parserSelect.style.cursor = 'not-allowed';
    } else {
        figlet.setParser(elements.parserSelect.value);
    }

    figlet.setFont(state.currentFont);

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
        .map(font => `<option value="${font}" ${font === state.currentFont ? 'selected' : ''}>${font}</option>`)
        .join('');

    // Generate font gallery
    generateFontGallery();
}

// Generate font gallery with previews
function generateFontGallery() {
    const previewText = 'Hi';
    const fragment = document.createDocumentFragment();

    state.fonts.forEach(font => {
        const card = document.createElement('div');
        card.className = `font-card ${font === state.currentFont ? 'active' : ''}`;
        card.dataset.font = font;

        const result = figlet.renderWithFont(previewText, font);
        let preview = result.error ? 'Error loading font' : result.result;

        // Trim leading whitespace from each line (for right-to-left fonts like ivrit)
        preview = preview.split('\n').map(line => line.trimStart()).join('\n');

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
        const clampedWidth = Math.max(20, Math.min(200, width));
        const result = figlet.setWidth(clampedWidth);
        if (result.error) {
            showToast(`Width error: ${result.error}`, 'error');
        }
        render();
    });

    // Alignment select
    elements.alignSelect.addEventListener('change', () => {
        figlet.setJustification(parseInt(elements.alignSelect.value));
        render();
    });

    // Parser select
    elements.parserSelect.addEventListener('change', () => {
        // Don't allow changing parser if colors are set
        const colorsStr = elements.colorsInput.value.trim();
        if (colorsStr !== '') {
            // Force back to HTML
            elements.parserSelect.value = 'html';
            showToast('HTML output is required when colors are set', 'info');
            return;
        }

        const parser = elements.parserSelect.value;
        const result = figlet.setParser(parser);
        if (result.error) {
            showToast(`Parser error: ${result.error}`, 'error');
        }
        render();
    });

    // Colors input - debounced
    elements.colorsInput.addEventListener('input', () => {
        clearTimeout(state.debounceTimer);
        state.debounceTimer = setTimeout(() => {
            const colorsStr = elements.colorsInput.value.trim();
            if (colorsStr === '') {
                // Clear colors - re-enable parser select
                const result = figlet.setColors([]);
                if (result.error) {
                    showToast(`Color error: ${result.error}`, 'error');
                } else {
                    // Re-enable parser select
                    elements.parserSelect.disabled = false;
                    elements.parserSelect.style.opacity = '1';
                    elements.parserSelect.style.cursor = 'pointer';
                }
            } else {
                // Parse colors
                const colors = colorsStr.split(';').map(c => c.trim()).filter(c => c);
                const result = figlet.setColors(colors);
                if (result.error) {
                    showToast(`Color error: ${result.error}`, 'error');
                } else {
                    // Force HTML parser when colors are set
                    if (elements.parserSelect.value !== 'html') {
                        elements.parserSelect.value = 'html';
                        const parserResult = figlet.setParser('html');
                        if (parserResult.error) {
                            showToast(`Parser error: ${parserResult.error}`, 'error');
                        }
                    }
                    // Disable parser select
                    elements.parserSelect.disabled = true;
                    elements.parserSelect.style.opacity = '0.6';
                    elements.parserSelect.style.cursor = 'not-allowed';
                }
            }
            render();
        }, 300);
    });

    // Copy button
    elements.copyBtn.addEventListener('click', copyToClipboard);

    // Download button
    elements.downloadBtn.addEventListener('click', downloadText);

    // Play button
    elements.playBtn.addEventListener('click', toggleAnimation);

    // Animation settings changes
    [elements.animationSelect, elements.animationDelay].forEach(el => {
        el.addEventListener('change', () => {
            if (animation.playing) stopAnimation();
        });
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

// Toggle Animation
async function toggleAnimation() {
    if (animation.playing) {
        stopAnimation();
    } else {
        await startAnimation();
    }
}

// Start Animation
async function startAnimation() {
    const text = elements.textInput.value || 'Hello';
    const animType = elements.animationSelect.value;
    const delay = parseInt(elements.animationDelay.value) || 50;

    if (!animType) {
        showToast('Please select an animation type', 'info');
        return;
    }

    // Prepare UI
    elements.playBtn.disabled = true;
    const span = elements.playBtn.querySelector('span');
    span.textContent = 'Generating...';

    // Small delay to allow UI to update
    await new Promise(r => setTimeout(r, 10));

    try {
        const result = figlet.generateAnimation(text, animType, delay);
        if (result.error) {
            showToast(`Generating failed: ${result.error}`, 'error');
            elements.playBtn.disabled = false;
            span.textContent = 'Play';
            return;
        }

        animation.frames = result.frames;
        animation.currentFrame = 0;
        animation.playing = true;

        elements.playBtn.disabled = false;
        elements.playBtn.classList.add('playing');
        span.textContent = 'Stop';

        playNextFrame();
    } catch (e) {
        console.error(e);
        showToast('Animation failed to start', 'error');
        elements.playBtn.disabled = false;
        span.textContent = 'Play';
    }
}

// Stop Animation
function stopAnimation() {
    animation.playing = false;
    clearTimeout(animation.timer);

    elements.playBtn.classList.remove('playing');
    elements.playBtn.querySelector('span').textContent = 'Play';

    render();
}

// Play Next Frame
function playNextFrame() {
    if (!animation.playing) return;

    const frame = animation.frames[animation.currentFrame];
    const parser = elements.parserSelect.value;

    if (parser === 'html') {
        elements.output.innerHTML = frame.content;
    } else {
        elements.output.textContent = frame.content;
    }

    animation.currentFrame = (animation.currentFrame + 1) % animation.frames.length;
    animation.timer = setTimeout(playNextFrame, frame.delay);
}

// Render the text
function render() {
    if (!state.ready || animation.playing) return;

    const text = elements.textInput.value || 'Hello';
    const parser = elements.parserSelect.value;
    const result = figlet.render(text);

    if (result.error) {
        elements.output.textContent = `Error: ${result.error}`;
        return;
    }

    const outputText = result.result || '(empty output)';

    // If HTML parser, render as HTML, otherwise as text
    if (parser === 'html') {
        elements.output.innerHTML = outputText;
    } else {
        elements.output.textContent = outputText;
    }
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
