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

function main() {
  const parserFile = process.argv[2];
  if (!parserFile) {
    throw new Error("expected parser file argument");
  }
  const MicronParser = loadParser(parserFile);
  const input = fs.readFileSync(0, "utf8");
  const cases = JSON.parse(input);
  const outputs = cases.map((c) => {
    const p = new MicronParser(Boolean(c.dark), Boolean(c.mono));
    return p.convertMicronToHtml(String(c.markup));
  });
  process.stdout.write(JSON.stringify(outputs));
}

main();
