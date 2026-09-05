# Changelog

Dates use YYYY-MM-DD.

## [1.1.0] - 2026-09-05 [unreleased]

### Added

- Document IR: Parse, RenderHTML, ParseWithDiagnostics, Lint, source spans, and JSON export.
- Builder helpers, incremental parse, and ANSI render for TUI hosts.
- Dialect notes and conformance examples in spec/micron.txt.
- Tree-sitter grammar under tree-sitter-micron/.
- C shared library (make lib) with header in bindings/c/.
- Language bindings for Python, Node, Rust, Java, C#, Ruby, PHP, Zig, Dart, Swift, and Perl.
- Release packaging for multi-platform libmicron and binding artifacts.
- CI bindings smoke (make bindings-test) and GitHub Pages workflow.
- Playground: tabs, diagnostics panel, source color pickers for page and inline colors, and related editor tools.
- Python oracle / semantic fingerprint tests against NomadNet-shaped output.

### Fixed

- Nested section depth capped at 16.
- ForceMonospace no longer breaks Arabic, Persian, Hebrew, and related scripts (dir=auto, logical section indent).

### Changed

- NomadNet Python is the primary dialect authority; micron-parser-js is secondary interop.
- Go toolchain requirement raised to 1.26.5+.
- ConvertMicronToHTML streams HTML again (no Document IR on the hot path).
- ForceMonospace matches micron-parser-js wrapWord: plain ASCII stays bare, Mu-mnt only for specials and non-ASCII cells while complex scripts stay unsplintered.
- Benchmarks updated: native ~0.37 ms; browser WASM ~0.56 ms vs JS ~3.54 ms (~6.31x); Node WASM ~1.42 ms vs JS ~3.39 ms.

## [1.0.6] - 2026-07-04

### Fixed

- Backslash before [ in formatting mode (relay-style timestamps).
- Matching backslash rules in the reference JS parser.

### Added

- Regression corpus, WASM smoke (make test-wasm), vendor JS sync checks, and make verify.
- Stronger CI gates before release artifacts publish.

### Changed

- Module path is local micron-parser-go.
- JS interop uses semantic visible-text comparison across theme and monospace modes.

## [1.0.1] - 2026-05-01

### Fixed

- Monospace formatting only when ForceMonospace is set.

### Changed

- Leaner color and HTML handling in the micron package.

## [1.0.0] - 2026-04-30

Initial stable release of the Micron parser and HTML renderer for Go and WebAssembly.
