import { init, createInstance } from '../src/index';

describe('FIGlet NPM Package', () => {
    beforeAll(async () => {
        // Ensure WASM is initialized
        await init();
    });

    test('Basic rendering', async () => {
        const result = await createInstance();
        const art = result.render('Test').result;
        expect(art).toContain('_');
        expect(art.length).toBeGreaterThan(0);
    });

    test('Instance isolation', async () => {
        const instance1 = await createInstance({ deutsch: false });
        const instance2 = await createInstance({ deutsch: true });

        const res1 = instance1.render('[').result;
        const res2 = instance2.render('[').result;

        expect(res1).not.toBe(res2);

        // Verify cross-interference
        const res1_again = instance1.render('[').result;
        expect(res1_again).toBe(res1);
    });

    test('Configuration persistence', async () => {
        const instance = await createInstance({ justification: 'center', width: 100 });

        // Change font, should keep justification
        await instance.setFont('slant');
        const art = instance.render('Test').result;

        // Center justification usually adds leading spaces
        expect(art.startsWith(' ')).toBe(true);
    });

    test('Smush modes', async () => {
        const fwInstance = await createInstance({ font: 'standard', smushMode: -1 });
        const kInstance = await createInstance({ font: 'standard', smushMode: 0 });

        const fwArt = fwInstance.render('WA').result;
        const kArt = kInstance.render('WA').result;

        expect(kArt.length).toBeLessThan(fwArt.length);
    });

    test('Colors and HTML parser', async () => {
        const instance = await createInstance({
            colors: ['red', 'green'],
            parser: 'html'
        });
        const art = instance.render('Test').result;
        expect(art).toContain('<span');
        expect(art).toContain('color: rgb');
    });

    test('List fonts', async () => {
        const instance = await createInstance();
        const fonts = instance.listFonts();
        expect(Array.isArray(fonts.fonts)).toBe(true);
        expect(fonts.fonts.length).toBeGreaterThan(0);
        expect(fonts.fonts).toContain('standard');
    });

    test('API compatibility (top-level functions)', async () => {
        const { render, getVersion } = require('../src/index');
        const art = await render('Test');
        expect(typeof art).toBe('string');

        const version = await getVersion();
        expect(typeof version).toBe('string');
        expect(version).toMatch(/\d+\.\d+\.\d+/);
    });
});
