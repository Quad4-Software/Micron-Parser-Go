# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

GOROOT := $(shell go env GOROOT)
WASM_OUT := web/micron.wasm
WASM_JS := web/wasm_exec.js
REF_JS := micron/testdata/micron-parser.js
VENDOR_JS := web/static/vendor/micron-parser.js

.PHONY: all test test-race test-smoke test-interop test-wasm fuzz wasm clean cover bench bench-go bench-js sync-vendor-js check-vendor-js verify

all: check-vendor-js test wasm

test:
	go test -count=1 -cover ./...

test-race:
	go test -count=1 -race ./...

test-smoke:
	go test -count=1 ./micron -run 'TestSmoke|TestEdge|TestSecurity|TestConcurrent|TestNoGoroutineLeak|TestRegressionCorpus'

test-interop:
	go test -count=1 ./micron -run 'TestInterop'

test-wasm: wasm
	node ./micron/testdata/wasm_smoke.js

sync-vendor-js:
	cp $(REF_JS) $(VENDOR_JS)

check-vendor-js:
	@diff -q $(REF_JS) $(VENDOR_JS) >/dev/null || (echo "reference JS out of sync; run: make sync-vendor-js" >&2; exit 1)

verify: check-vendor-js test-race test-interop test-wasm fuzz

FUZZTIME ?= 3s

fuzz:
	set -e; for fuzz in \
		FuzzConvertMicronToHTML \
		FuzzLightThemeConvertMicronToHTML \
		FuzzFormatNomadnetworkURL \
		FuzzBuildRequestPayload \
		FuzzCollectFormFields \
		FuzzParseHeaderTags; do \
		go test ./micron -run=^$$ -fuzz=$$fuzz -fuzztime=$(FUZZTIME); \
	done

lint:
	revive -config revive.toml -formatter friendly ./micron/*

cover:
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

bench: bench-go bench-js

bench-go:
	go test ./micron -bench=BenchmarkConvertNomadNetGuide -benchmem -count=10 -timeout=30m

bench-js:
	node ./micron/testdata/bench_nomadnet.js

wasm:
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o $(WASM_OUT) ./cmd/wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $(WASM_JS)

clean:
	rm -f $(WASM_OUT) $(WASM_JS) coverage.out coverage.html
