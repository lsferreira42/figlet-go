import * as fs from 'fs';
import * as path from 'path';

// Polyfills for Go's wasm_exec.js
if (typeof global.TextEncoder === 'undefined') {
    const { TextEncoder, TextDecoder } = require('util');
    global.TextEncoder = TextEncoder;
    global.TextDecoder = TextDecoder;
}

if (typeof (global as any).performance === 'undefined') {
    (global as any).performance = {
        now: () => Date.now(),
    };
}

if (typeof (global as any).crypto === 'undefined') {
    (global as any).crypto = {
        getRandomValues: (buf: any) => require('crypto').randomFillSync(buf),
    };
}

// Load wasm_exec.js
const wasmExecPath = path.join(__dirname, '../dist/wasm_exec.js');
require(wasmExecPath);

// Global figlet will be set by Go WASM
// We'll export a helper to wait for it
export const waitForFiglet = async () => {
    return new Promise<any>((resolve, reject) => {
        const timeout = setTimeout(() => reject('Timeout waiting for FIGlet'), 5000);
        const check = () => {
            if ((global as any).figlet) {
                clearTimeout(timeout);
                resolve((global as any).figlet);
            } else {
                setTimeout(check, 50);
            }
        };
        check();
    });
};
