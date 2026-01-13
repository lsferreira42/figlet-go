#!/usr/bin/env node

/**
 * Build script for figlet-go npm package
 * Copies necessary files to dist/ folder
 */

const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..');
const DIST = path.join(ROOT, 'dist');
const SRC = path.join(ROOT, 'src');

// Ensure dist directory exists
if (!fs.existsSync(DIST)) {
    fs.mkdirSync(DIST, { recursive: true });
}

// Copy source files
console.log('Copying source files...');
fs.copyFileSync(
    path.join(SRC, 'index.js'),
    path.join(DIST, 'index.js')
);
fs.copyFileSync(
    path.join(SRC, 'index.d.ts'),
    path.join(DIST, 'index.d.ts')
);

// Create ESM version
console.log('Creating ESM module...');
let esmContent = fs.readFileSync(path.join(SRC, 'index.js'), 'utf8');
esmContent = esmContent
    .replace(/const fs = require\('fs'\);/g, "import fs from 'fs';")
    .replace(/const path = require\('path'\);/g, "import path from 'path';")
    .replace(/require\(wasmExecPath\);/g, "await import(wasmExecPath);")
    .replace(/module\.exports = \{[\s\S]*\};/, `
export {
    init,
    render,
    renderWithFont,
    listFonts,
    getVersion,
    createInstance,
};

export default {
    init,
    render,
    renderWithFont,
    listFonts,
    getVersion,
    createInstance,
};
`);
fs.writeFileSync(path.join(DIST, 'index.mjs'), esmContent);

// Check if WASM file exists in parent directory
const wasmSrc = path.join(ROOT, '..', 'website', 'figlet.wasm');
const wasmDst = path.join(DIST, 'figlet.wasm');

if (fs.existsSync(wasmSrc)) {
    console.log('Copying WASM file...');
    fs.copyFileSync(wasmSrc, wasmDst);
} else {
    console.warn('Warning: figlet.wasm not found. Run "make build-wasm" first.');
}

// Copy wasm_exec.js
const wasmExecSrc = path.join(ROOT, '..', 'website', 'wasm_exec.js');
const wasmExecDst = path.join(DIST, 'wasm_exec.js');

if (fs.existsSync(wasmExecSrc)) {
    console.log('Copying wasm_exec.js...');
    fs.copyFileSync(wasmExecSrc, wasmExecDst);
} else {
    console.warn('Warning: wasm_exec.js not found.');
}

console.log('Build complete!');
