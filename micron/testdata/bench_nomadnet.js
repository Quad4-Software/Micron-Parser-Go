"use strict";

const fs = require("fs");
const path = require("path");

require(path.join(__dirname, "node_dom_stub.js"));

function loadParser(filePath) {
  const full = path.resolve(filePath);
  let source = fs.readFileSync(full, "utf8");
  source = source.replace(/\nexport default MicronParser;\s*$/, "\nglobalThis.MicronParser = MicronParser;\n");
  eval(source);
  return globalThis.MicronParser;
}

function mean(a) {
  return a.reduce((s, x) => s + x, 0) / a.length;
}

function stdev(a, m) {
  if (a.length < 2) return 0;
  const v = a.reduce((s, x) => s + (x - m) ** 2, 0) / (a.length - 1);
  return Math.sqrt(v);
}

function runBatch(parser, markup, iterations) {
  const t0 = process.hrtime.bigint();
  for (let i = 0; i < iterations; i++) {
    parser.convertMicronToHtml(markup);
  }
  const t1 = process.hrtime.bigint();
  return Number(t1 - t0);
}

const dir = __dirname;
const parserPath = path.join(dir, "micron-parser.js");
const guidePath = path.join(dir, "nomadnet_guide.mu");

const MicronParser = loadParser(parserPath);
const markup = fs.readFileSync(guidePath, "utf8");
const inputBytes = Buffer.byteLength(markup, "utf8");

const dark = true;
const mono = true;

const warmup = 5;
const warmP = new MicronParser(dark, mono);
for (let i = 0; i < warmup; i++) {
  warmP.convertMicronToHtml(markup);
}

let innerIter = 8;
const targetNs = 150 * 1e6;
const calP = new MicronParser(dark, mono);
for (let attempt = 0; attempt < 12; attempt++) {
  const ns = runBatch(calP, markup, innerIter);
  if (ns >= targetNs) break;
  innerIter = Math.min(innerIter * 2, 8192);
}

const runs = 10;
const perRunNsPerOp = [];
const benchP = new MicronParser(dark, mono);

for (let r = 0; r < runs; r++) {
  const totalNs = runBatch(benchP, markup, innerIter);
  perRunNsPerOp.push(totalNs / innerIter);
}

const m = mean(perRunNsPerOp);
const sd = stdev(perRunNsPerOp, m);
const min = Math.min(...perRunNsPerOp);
const max = Math.max(...perRunNsPerOp);
const mibPerSec = inputBytes / (m / 1e9) / (1024 * 1024);

console.log("reference JS (micron-parser.js) — NomadNet guide corpus");
console.log("  input: " + inputBytes + " B");
console.log("  parser: MicronParser(" + dark + ", " + mono + ")");
console.log("  inner iterations per measured run: " + innerIter);
console.log("  runs: " + runs);
console.log("  ns/op per run: " + perRunNsPerOp.map((x) => x.toFixed(0)).join(", "));
console.log("  mean ns/op: " + m.toFixed(0));
console.log("  stdev ns/op: " + sd.toFixed(0));
console.log("  min ns/op: " + min.toFixed(0));
console.log("  max ns/op: " + max.toFixed(0));
console.log("  mean throughput: " + mibPerSec.toFixed(2) + " MiB/s");
