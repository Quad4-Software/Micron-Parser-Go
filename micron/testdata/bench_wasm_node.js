#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const repoRoot = path.resolve(__dirname, "..", "..");
const wasmPath = path.join(repoRoot, "web", "micron.wasm");
const wasmExecPath = path.join(repoRoot, "web", "wasm_exec.js");
const guidePath = path.join(__dirname, "nomadnet_guide.mu");

function mean(a) {
  return a.reduce((s, x) => s + x, 0) / a.length;
}

function stdev(a, m) {
  if (a.length < 2) return 0;
  const v = a.reduce((s, x) => s + (x - m) ** 2, 0) / (a.length - 1);
  return Math.sqrt(v);
}

function runBatch(fn, markup, iterations) {
  const t0 = process.hrtime.bigint();
  for (let i = 0; i < iterations; i++) fn(markup);
  const t1 = process.hrtime.bigint();
  return Number(t1 - t0) / iterations;
}

async function main() {
  if (!fs.existsSync(wasmPath) || !fs.existsSync(wasmExecPath)) {
    console.error("missing wasm artifacts (run make wasm)");
    process.exit(1);
  }
  require(wasmExecPath);
  const go = new Go();
  const result = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
  go.run(result.instance);
  if (typeof globalThis.micronConvert !== "function") {
    console.error("micronConvert missing");
    process.exit(1);
  }

  const markup = fs.readFileSync(guidePath, "utf8");
  const inputBytes = Buffer.byteLength(markup, "utf8");
  const fn = (src) => globalThis.micronConvert(src, true, true);

  for (let i = 0; i < 5; i++) fn(markup);

  let innerIter = 8;
  const targetNs = 150e6;
  for (let attempt = 0; attempt < 12; attempt++) {
    const ns = runBatch(fn, markup, innerIter) * innerIter;
    if (ns >= targetNs) break;
    innerIter = Math.min(innerIter * 2, 8192);
  }

  const runs = 10;
  const perRun = [];
  for (let r = 0; r < runs; r++) perRun.push(runBatch(fn, markup, innerIter));

  const m = mean(perRun);
  const sd = stdev(perRun, m);
  const mib = inputBytes / (m / 1e9) / (1024 * 1024);
  console.log("Go WASM (micronConvert) - NomadNet guide corpus (Node)");
  console.log("  input: " + inputBytes + " B");
  console.log("  inner iterations per measured run: " + innerIter);
  console.log("  runs: " + runs);
  console.log("  mean ns/op: " + m.toFixed(0));
  console.log("  stdev ns/op: " + sd.toFixed(0));
  console.log("  min ns/op: " + Math.min(...perRun).toFixed(0));
  console.log("  max ns/op: " + Math.max(...perRun).toFixed(0));
  console.log("  mean throughput: " + mib.toFixed(2) + " MiB/s");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
