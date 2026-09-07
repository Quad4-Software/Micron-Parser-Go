# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

GOROOT := $(shell go env GOROOT)
WASM_OUT := web/micron.wasm
WASM_JS := web/wasm_exec.js
REF_JS := micron/testdata/micron-parser.js
VENDOR_JS := web/static/vendor/micron-parser.js
DIST := dist
LIB_NAME := libmicron
UNAME_S := $(shell uname -s 2>/dev/null || echo unknown)

# Reticulum / rngit dual-push origin (GitHub fetch, GitHub+RNS push).
RNS_DEST ?= 06a54b505bb67b25ef3f8097e8001edc
RNS_GROUP ?= public
RNS_REPO ?= micron-parser-go
RNS_REMOTE ?= rns://$(RNS_DEST)/$(RNS_GROUP)/$(RNS_REPO)
GH_REPO ?= Quad4-Software/Micron-Parser-Go
CHANGELOG ?= CHANGELOG.md
# TAG required for rngit-release / rngit-notes / rngit-push-tag (e.g. TAG=v1.1.0).
TAG ?=

ifeq ($(OS),Windows_NT)
  LIB_FILE := $(LIB_NAME).dll
else ifneq (,$(findstring MINGW,$(UNAME_S)))
  LIB_FILE := $(LIB_NAME).dll
else ifneq (,$(findstring MSYS,$(UNAME_S)))
  LIB_FILE := $(LIB_NAME).dll
else ifneq (,$(findstring CYGWIN,$(UNAME_S)))
  LIB_FILE := $(LIB_NAME).dll
else ifeq ($(UNAME_S),Darwin)
  LIB_FILE := $(LIB_NAME).dylib
else
  LIB_FILE := $(LIB_NAME).so
endif

.PHONY: all test test-race test-smoke test-interop test-interop-python test-wasm fuzz wasm serve-web lib lib-install bindings-test clean cover bench bench-go bench-js bench-wasm sync-vendor-js check-vendor-js verify fmt fix lint lint-gosec vet rngit-notes rngit-push rngit-push-tag rngit-release rngit-release-local rngit-list

all: check-vendor-js test wasm

test:
	go test -count=1 -cover ./...

test-race:
	go test -count=1 -race ./...

test-smoke:
	go test -count=1 ./micron -run 'TestSmoke|TestEdge|TestSecurity|TestConcurrent|TestNoGoroutineLeak|TestRegressionCorpus'

test-interop:
	go test -count=1 ./micron -run 'TestInterop'

test-interop-python:
	go test -count=1 ./micron -run 'TestPythonOracle'

test-wasm: wasm
	node ./micron/testdata/wasm_smoke.js

sync-vendor-js:
	cp $(REF_JS) $(VENDOR_JS)

check-vendor-js:
	@diff -q $(REF_JS) $(VENDOR_JS) >/dev/null || (echo "reference JS out of sync; run: make sync-vendor-js" >&2; exit 1)

verify: check-vendor-js fmt-check lint vet lint-gosec test-race test-interop test-interop-python test-wasm fuzz

FUZZTIME ?= 3s

fuzz:
	set -e; for fuzz in \
		FuzzConvertMicronToHTML \
		FuzzLightThemeConvertMicronToHTML \
		FuzzFormatNomadnetworkURL \
		FuzzBuildRequestPayload \
		FuzzCollectFormFields \
		FuzzParseHeaderTags; do \
		go test ./micron -run=^$$ -fuzz=$$fuzz -fuzztime=$(FUZZTIME) -parallel=1 -timeout=30m; \
	done

GOLANGCI_LINT ?= golangci-lint
GOSEC ?= gosec

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './bindings/*' -not -path './examples/_scratch/*')
	go fix ./...
	$(GOLANGCI_LINT) fmt ./...

fmt-check:
	@unformatted=$$(gofmt -l $$(find . -name '*.go' -not -path './bindings/*' -not -path './examples/_scratch/*')); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:" >&2; \
		echo "$$unformatted" >&2; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) fmt --diff ./...

fix:
	go fix ./...
	$(GOLANGCI_LINT) run --fix-diff ./...

lint:
	GOTOOLCHAIN=go1.27.1 $(GOLANGCI_LINT) run ./...

lint-gosec:
	GOTOOLCHAIN=go1.27.1 $(GOSEC) -exclude-dir=examples -exclude-dir=bindings -exclude=G404,G304,G204,G306,G103,G115 ./...

vet:
	GOTOOLCHAIN=go1.27.1 go vet ./...

cover:
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

bench: bench-go bench-js bench-wasm

bench-go:
	go test ./micron -bench=BenchmarkConvertNomadNetGuide -benchmem -count=10 -timeout=30m

bench-js:
	node ./micron/testdata/bench_nomadnet.js

bench-wasm: wasm
	node ./micron/testdata/bench_wasm_node.js

wasm:
	VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo dev); \
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w -X main.version=$$VERSION" -o $(WASM_OUT) ./cmd/wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $(WASM_JS)

serve-web: wasm
	python3 -m http.server 8080 --directory web

lib:
	mkdir -p $(DIST)
	CGO_ENABLED=1 go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
		-o $(DIST)/$(LIB_FILE) ./cmd/libmicron
	cp bindings/c/micron.h $(DIST)/micron.h
	rm -f $(DIST)/$(LIB_NAME).h

lib-install: lib
	@echo "Built $(DIST)/$(LIB_FILE) and $(DIST)/micron.h"

bindings-test: lib
	@set -e; \
	MICRON_LIB_PATH="$(CURDIR)/$(DIST)/$(LIB_FILE)"; \
	export MICRON_LIB_PATH; \
	if command -v python3 >/dev/null 2>&1; then \
		python3 bindings/python/smoke_test.py; \
	fi; \
	if command -v node >/dev/null 2>&1; then \
		(cd bindings/node && npm install --omit=dev --no-fund --no-audit >/dev/null && node smoke_test.js); \
	fi; \
	if command -v ruby >/dev/null 2>&1; then \
		ruby -e 'require "ffi"' 2>/dev/null || gem install ffi --user-install --quiet; \
		ruby bindings/ruby/smoke_test.rb; \
	fi; \
	if command -v cargo >/dev/null 2>&1; then \
		cargo run --manifest-path bindings/rust/micron/Cargo.toml --quiet --bin micron-smoke; \
	fi; \
	if command -v javac >/dev/null 2>&1; then \
		JNA_JAR=$$(ls /tmp/micron-jna/jna-*.jar 2>/dev/null | head -1); \
		if [ -z "$$JNA_JAR" ]; then \
			mkdir -p /tmp/micron-jna; \
			curl -fsSL -o /tmp/micron-jna/jna-5.14.0.jar \
				https://repo1.maven.org/maven2/net/java/dev/jna/jna/5.14.0/jna-5.14.0.jar; \
			JNA_JAR=/tmp/micron-jna/jna-5.14.0.jar; \
		fi; \
		rm -rf /tmp/micron-jna/out && mkdir -p /tmp/micron-jna/out; \
		javac -cp "$$JNA_JAR" -d /tmp/micron-jna/out bindings/java/src/main/java/io/quad4/micron/*.java; \
		java -cp "/tmp/micron-jna/out:$$JNA_JAR" -Djna.library.path="$(CURDIR)/$(DIST)" io.quad4.micron.Micron; \
	fi; \
	if command -v dotnet >/dev/null 2>&1; then \
		(cd bindings/csharp/smoke && dotnet run -q); \
	fi; \
	if command -v gcc >/dev/null 2>&1; then \
		gcc -o $(DIST)/micron_c_smoke bindings/c/smoke.c -Ibindings/c -L$(DIST) -lmicron \
			-Wl,-rpath,'$$ORIGIN'; \
		$(DIST)/micron_c_smoke; \
	fi; \
	if command -v php >/dev/null 2>&1; then \
		php -d extension=ffi -d ffi.enable=true bindings/php/smoke_test.php; \
	fi; \
	if command -v zig >/dev/null 2>&1; then \
		(cd bindings/zig && zig build smoke); \
	fi; \
	if command -v dart >/dev/null 2>&1; then \
		(cd bindings/dart && dart pub get >/dev/null && dart run bin/smoke.dart); \
	fi; \
	if command -v swift >/dev/null 2>&1; then \
		(cd bindings/swift && swift run micron-smoke); \
	fi; \
	if command -v perl >/dev/null 2>&1; then \
		export PERL5LIB="$${HOME}/perl5/lib/perl5:$${PERL5LIB:-}"; \
		perl -MFFI::Platypus -e 1 2>/dev/null || \
			PERL_MM_USE_DEFAULT=1 cpan FFI::Platypus >/dev/null 2>&1 || true; \
		if perl -MFFI::Platypus -e 1 2>/dev/null; then \
			perl bindings/perl/smoke_test.pl; \
		fi; \
	fi

clean:
	rm -f $(WASM_OUT) $(WASM_JS) coverage.out coverage.html
	rm -rf $(DIST)

# Print CHANGELOG section for TAG (strips leading v).
rngit-notes:
	@test -n "$(TAG)" || (echo "set TAG=vX.Y.Z" >&2; exit 2)
	@./scripts/changelog-entry.sh "$(TAG)" "$(CHANGELOG)"

# Push current branch to dual-push origin (GitHub + RNS).
rngit-push:
	git push origin HEAD

# Push TAG to dual-push origin. Example: make rngit-push-tag TAG=v1.1.0
rngit-push-tag:
	@test -n "$(TAG)" || (echo "set TAG=vX.Y.Z" >&2; exit 2)
	git push origin "refs/tags/$(TAG)"

# Publish rngit release for TAG using GitHub release assets + CHANGELOG notes.
# Example: make rngit-release TAG=v1.1.0
rngit-release:
	@test -n "$(TAG)" || (echo "set TAG=vX.Y.Z" >&2; exit 2)
	RNS_REMOTE="$(RNS_REMOTE)" GH_REPO="$(GH_REPO)" CHANGELOG="$(CHANGELOG)" \
		./scripts/rngit-release.sh "$(TAG)"

# Same as rngit-release but upload from a local directory (default: dist).
# Example: make rngit-release-local TAG=v1.1.0 ARTIFACTS=./dist
ARTIFACTS ?= $(DIST)
rngit-release-local:
	@test -n "$(TAG)" || (echo "set TAG=vX.Y.Z" >&2; exit 2)
	RNS_REMOTE="$(RNS_REMOTE)" CHANGELOG="$(CHANGELOG)" \
		./scripts/rngit-release.sh "$(TAG)" "$(ARTIFACTS)"

rngit-list:
	rngit release "$(RNS_REMOTE)" list
