# micron-parser-go

Micron parser and HTML renderer for Go and WebAssembly.
Dialect authority is NomadNet Python
([MicronParser.py](https://github.com/markqvist/NomadNet/blob/master/nomadnet/ui/textui/MicronParser.py)).
micron-parser-js is a secondary web interop target.

Playground: https://micron-parser-go.quad4.io/

Dialect notes:

```text
spec/micron.txt
```

## Requirements

- Go 1.27.1+
- Standard library only
- Node.js (optional, interop / JS bench)
- Python 3 (optional, NomadNet oracle)

## Usage

```go
import "micron-parser-go/micron"

p := micron.Parser{DarkTheme: true, ForceMonospace: true}
html := p.ConvertMicronToHTML("> Title\n\nHello `!world`! and `*micron`*.\n")
```

Parser is safe to reuse across goroutines. Related APIs:

```text
Parse / RenderHTML / ParseWithDiagnostics / Lint
RenderANSI / ParseIncremental / Builder
ParseHeaderTags
CollectFormFields / BuildRequestPayload
```

Header colors (leading lines, 3 or 6 hex digits):

```text
#!fg=RGB
#!bg=RGB
```

```go
colors := micron.ParseHeaderTags(markup)
html := p.ConvertMicronToHTML(markup)
```

Form / link helpers:

```go
fields := micron.CollectFormFields(inputs)
payload := micron.BuildRequestPayload(fields, dest, "user|plan")
```

## Performance

Corpus:

```text
micron/testdata/nomadnet_guide.mu
```

| Implementation | Environment | Mean | Notes |
|----------------|-------------|------|-------|
| Go native | `go test` amd64 | ~0.37 ms | ~0.35-0.38 ms/op, ~0.30 MB/op, ~2006 allocs/op, ~85 MB/s |
| Go WASM | browser `bench.html` | ~0.56 ms | 10 runs (512 inner iterations); stdev ~0.026 ms; min/max ~0.54-0.63 ms; ~19.2 MiB/s |
| micron-parser-js | browser `bench.html` | ~3.54 ms | 10 runs (64 inner iterations); stdev ~0.11 ms; min/max ~3.41-3.78 ms; ~3.0 MiB/s |
| Go WASM | Node + `wasm_exec.js` | ~1.42 ms | ~21.3 MiB/s (10 runs) |
| micron-parser-js | Node + DOM stub | ~3.39 ms | ~9.0 MiB/s (10 runs) |

**WASM vs reference JS (browser mean):** `6.31x` faster.

```text
make bench
```

Runs `bench-go`, `bench-js`, and `bench-wasm`. For the in-browser head-to-head: `make serve-web`, then open `/bench.html`.

## WASM

```text
make wasm
make serve-web
```

Open http://127.0.0.1:8080/ (also playground.html, bench.html).

```text
web/micron.wasm
web/wasm_exec.js
```

Globals after load:

```text
micronConvert(markup, darkTheme?, forceMonospace?) -> string
micronLint(markup) -> string
micronParse(markup, darkTheme?, forceMonospace?) -> string
micronCollectFields(rootSelector?) -> string
micronResolveLink(rootSelector?, destination?, fieldsSpec?) -> string
```

Optional demo hook: window.onMicronLink(payload, element).
Preview root defaults to #preview.

## Make

```text
make test
make verify
make lint
make fuzz
make bench
make wasm
```

make verify runs vendor check, fmt/lint/vet/gosec, race tests, JS + Python
interop, wasm smoke, and fuzz.

## License

0BSD. See LICENSE.
