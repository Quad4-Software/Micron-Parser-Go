# micron-parser-go

Blazingly fast Micron parser and HTML renderer for Go and WebAssembly, based on [micron-parser-js](https://github.com/RFnexus/micron-parser-js). For Go (library) or web based (WASM) applications.

Playground: https://micron-parser-go.quad4.io/

## Requirements

- Go 1.26.2+
- No third-party Go modules (standard library only)
- Node.js (optional): interop tests, reference-JS benchmarks, and the `bench-js` Makefile target

## Library

Import path:

```go
import "git.quad4.io/Go-Libs/micron-parser-go/micron"
```

`micron.Parser` holds only two settings: **`DarkTheme`** picks light or dark default colors for the HTML output, and **`ForceMonospace`** toggles monospace styling for the rendered page. The type has no mutable conversion state; a single `Parser` value is safe to reuse from multiple goroutines.

### Convert Micron to HTML

`ConvertMicronToHTML` parses the full document, applies optional leading `#!fg=` / `#!bg=` header lines (see below), and returns a self-contained HTML fragment safe for insertion into a host page (escaping is applied consistently with the reference implementation).

```go
package main

import (
	"fmt"

	"git.quad4.io/Go-Libs/micron-parser-go/micron"
)

func main() {
	p := micron.Parser{
		DarkTheme:      true,
		ForceMonospace: true,
	}
	src := "> Title\n\nHello `!world`! and `*micron`*.\n"
	html := p.ConvertMicronToHTML(src)
	fmt.Print(html)
}
```

For a light theme and proportional fonts (closer to some terminal themes):

```go
p := micron.Parser{DarkTheme: false, ForceMonospace: false}
html := p.ConvertMicronToHTML(markup)
```

### Page header colors (optional)

Leading lines of the form `#!fg=RGB` and `#!bg=RGB` (three or six hex digits per color) set default page foreground and background. You can read them without rendering via `ParseHeaderTags`, for example to style a surrounding shell or iframe:

```go
markup := "#!fg=ccc\n#!bg=222\n\n> Section\nBody.\n"
colors := micron.ParseHeaderTags(markup)
// colors.FG, colors.BG — may be empty if not set
_ = colors

p := micron.Parser{DarkTheme: true, ForceMonospace: true}
html := p.ConvertMicronToHTML(markup) // header tags are applied during conversion
```

### Link requests and form fields

For applications that render HTML to the client and submit Micron-style links, `CollectFormFields` and `BuildRequestPayload` mirror the WASM helpers: turn a list of input snapshots into a field map, then combine with link `destination` and `fieldsSpec` (e.g. `*` for all fields, or `name|other`).

```go
inputs := []micron.FieldInput{
	{Type: "text", Name: "user", Value: "alice"},
	{Type: "checkbox", Name: "opts", Value: "1", Checked: true},
	{Type: "radio", Name: "plan", Value: "pro", Checked: true},
}
fields := micron.CollectFormFields(inputs)

payload := micron.BuildRequestPayload(
	fields,
	"/page/submit.mu`action=run|amount=10",
	"user|plan",
)
// payload.Destination, payload.Fields, payload.RequestVars — use as needed (JSON tags on RequestPayload)
_ = payload
```

## Performance

Benchmarks use the **NomadNet guide** micron source from Micron-Parser-JS (`11248` input bytes).

| Implementation | Environment | Mean time / conversion | Notes |
|----------------|---------------|------------------------|--------|
| This package (Go) | `go test` native amd64 | ~1.38 ms | 10× `BenchmarkConvertNomadNetGuide` runs; ~1.31–1.47 ms/op, ~4.34 MB/op, 4176 allocs/op |
| This package (Go WASM) | Browser `bench.html` | ~3.37 ms | 10 measured runs (64 inner iterations); stdev ~0.22 ms; min/max ~2.89–3.52 ms; ~3.18 MiB/s |
| Reference [micron-parser-js](https://github.com/RFnexus/micron-parser-js) | Browser `bench.html` | ~41.28 ms | 10 measured runs (8 inner iterations); stdev ~1.04 ms; min/max ~39.68–43.40 ms; ~0.26 MiB/s |
| Reference [micron-parser-js](https://github.com/RFnexus/micron-parser-js) | Node + DOM stub | ~21.0 ms | 10 measured runs; ~19.7–25.9 ms; ~0.51 MiB/s mean throughput |

**WebAssembly:** The browser build uses the same Go code as the native benchmark, but timing includes JS/WASM call overhead and is strongly browser-dependent. It will not match the `go test` numbers above.

**WASM vs reference JS (browser mean):** `12.24x` faster (`41.28 ms / 3.37 ms`).

**Reproduce**

```text
make bench
```

Runs native Go (`bench-go`, `-count=10`) and the Node script (`bench-js`). To summarize Go variance with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```text
go test ./micron -bench=BenchmarkConvertNomadNetGuide -benchmem -count=10 | tee /tmp/go.txt
benchstat /tmp/go.txt
```

## WASM demo

```text
make wasm
```

Open `web/index.html` in a browser (local file or any static server).  
`make wasm` writes both `web/micron.wasm` and `web/wasm_exec.js`. These artifacts are generated into `web/` and are intentionally gitignored.

### JavaScript API (globals)

After `wasm_exec.js` loads the module, the following functions are available on `globalThis` / `window`:

| Symbol | Signature | Purpose |
|--------|-----------|---------|
| `micronConvert` | `(markup: string, darkTheme?: boolean, forceMonospace?: boolean) => string` | Render Micron source to an HTML string. Defaults match the demo: dark `true`, monospace `true`. |
| `micronCollectFields` | `(rootSelector?: string) => string` | JSON string of form field values under `document.querySelector(rootSelector)` (default `#preview`). |
| `micronResolveLink` | `(rootSelector?: string, destination?: string, fieldsSpec?: string) => string` | JSON payload for link navigation, using the same field collection rules as the Go helpers. |

The WASM program registers these and then blocks on the Go scheduler (`select {}`); initialization is synchronous from the host perspective once instantiation completes.

### Application hooks

- **`window.onMicronLink`** — Optional. If defined, the demo calls it when the user activates a rendered Micron link (`data-action="openNode"`). Receives `(payload: object, element: Element)` where `payload` is the JSON from `micronResolveLink`. Use this to route in-app navigation, logging, or analytics without forking the stock HTML.
- **Preview container** — Host pages should keep a single preview root (e.g. `#preview`) so `micronCollectFields` and `micronResolveLink` resolve inputs consistently.

## Quality, verification and security

- Unit tests and edge/smoke suites are in `micron/*_test.go`
- Security tests cover HTML escaping and attribute escaping
- Fuzz targets cover parser conversion and header parsing
- Race/concurrency coverage is included in concurrent conversion tests
- Goroutine leak guard checks repeated conversion paths
- JS interop test compares output signatures against `micron-parser-js`
- Benchmarks: `make bench` (native Go + reference JS, NomadNet corpus)
- Property-based tests in `micron/property_test.go`
- `make fuzz` runs every fuzz target in `micron/fuzz_test.go` (override duration with `FUZZTIME=30s`)

## License

0BSD. See LICENSE.
