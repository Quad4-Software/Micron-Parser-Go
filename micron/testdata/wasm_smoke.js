#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const repoRoot = path.resolve(__dirname, "..", "..");
const wasmPath = path.join(repoRoot, "web", "micron.wasm");
const wasmExecPath = path.join(repoRoot, "web", "wasm_exec.js");

function fail(msg) {
    console.error("wasm_smoke:", msg);
    process.exit(1);
}

function visibleText(html) {
    return html
        .replace(/<[^>]+>/g, " ")
        .replace(/\s+/g, " ")
        .trim();
}

const cases = [
    {
        name: "relay-timestamp",
        markup: "`Fe81`!\\[04.07.2026 02:55]: `Ffffigloo: ``Hi from Canada!",
        dark: true,
        mono: true,
        wantVisible: ["[04.07.2026 02:55]:", "igloo:", "Hi from Canada!"],
        htmlMustNot: ["\\["],
    },
    {
        name: "xss-plain",
        markup: "x <script>alert(1)</script> <img src=x onerror=alert(1)>",
        dark: true,
        mono: false,
        wantVisible: ["alert(1)"],
        htmlMustNot: ["<script", "<img"],
    },
];

async function main() {
    if (!fs.existsSync(wasmPath)) {
        fail("missing " + wasmPath + " (run make wasm)");
    }
    if (!fs.existsSync(wasmExecPath)) {
        fail("missing " + wasmExecPath + " (run make wasm)");
    }

    require(wasmExecPath);
    if (typeof Go === "undefined") {
        fail("wasm_exec.js did not define Go");
    }

    const go = new Go();
    const wasmBytes = fs.readFileSync(wasmPath);
    let result;
    try {
        result = await WebAssembly.instantiate(wasmBytes, go.importObject);
    } catch (err) {
        fail("WebAssembly.instantiate failed: " + err);
    }
    go.run(result.instance);

    if (typeof globalThis.micronConvert !== "function") {
        fail("micronConvert was not registered");
    }

    for (const tc of cases) {
        const html = globalThis.micronConvert(tc.markup, tc.dark, tc.mono);
        if (typeof html !== "string") {
            fail(tc.name + ": micronConvert did not return a string");
        }
        const text = visibleText(html);
        const compact = text.replace(/ /g, "");
        for (const frag of tc.wantVisible) {
            const compactFrag = frag.replace(/ /g, "");
            if (!text.includes(frag) && !compact.includes(compactFrag)) {
                fail(tc.name + ": visible text missing " + JSON.stringify(frag) + " in " + JSON.stringify(text));
            }
        }
        for (const frag of tc.htmlMustNot || []) {
            if (html.includes(frag)) {
                fail(tc.name + ": html must not contain " + JSON.stringify(frag));
            }
        }
    }

    console.log("wasm_smoke: ok (" + cases.length + " cases)");
}

main().catch((err) => fail(String(err)));
