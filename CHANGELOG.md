# Changelog

All notable changes to this project are documented in this file. Dates use ISO 8601 (YYYY-MM-DD).

## [1.0.6] - 2026-07-04

### Fixed

- Backslash before `[` is handled correctly in formatting mode so relay-style timestamps such as `\[04.07.2026 02:55]:` render as `[04.07.2026 02:55]:` without a visible escape character.
- Reference JavaScript parser aligned for the same backslash rules in text and formatting modes.

### Added

- Regression corpus under `micron/testdata/regressions/` with table-driven tests for relay timestamps, figlet backslashes, and XSS escaping.
- WASM smoke test (`make test-wasm`) exercising `micronConvert` via Node and `wasm_exec.js`.
- `make check-vendor-js` / `make sync-vendor-js` to keep `micron/testdata/micron-parser.js` and `web/static/vendor/micron-parser.js` in sync.
- `make verify` for a single local QA pass (vendor check, race tests, interop, wasm smoke, fuzz).
- Release workflow gate: tests run before WASM artifacts are published.
- CI: wasm smoke, vendor JS diff in lint job, weekly scheduled fuzz (120s).

### Changed

- Go toolchain requirement raised to 1.26.4.
- Module path is now the local `micron-parser-go` (no remote git host in the import path).
- JS interop covers all theme and monospace combinations using semantic visible-text comparison.
- README quality section documents regression corpus, wasm smoke, and verify targets.

## [1.0.1] - 2026-05-01

### Fixed

- Line parsing applies monospace formatting only when `ForceMonospace` is set.

### Changed

- Reduced allocations by refactoring color and HTML handling in the `micron` package.
- README updates.

## [1.0.0] - 2026-04-30

Initial stable release of the Micron parser and HTML renderer for Go and WebAssembly.
