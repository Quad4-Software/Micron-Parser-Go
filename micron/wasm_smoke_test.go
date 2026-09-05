// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReferenceJSSyncedWithVendor(t *testing.T) {
	ref := filepath.Join("testdata", "micron-parser.js")
	vendor := filepath.Join("..", "web", "static", "vendor", "micron-parser.js")
	refBytes, err := os.ReadFile(ref)
	if err != nil {
		t.Fatal(err)
	}
	vendorBytes, err := os.ReadFile(vendor)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(refBytes, vendorBytes) {
		t.Fatalf("testdata/micron-parser.js and web/static/vendor/micron-parser.js differ; run: make sync-vendor-js")
	}
}

func TestWasmSmoke(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	wasm := filepath.Join("..", "web", "micron.wasm")
	if _, err := os.Stat(wasm); err != nil {
		t.Skip("web/micron.wasm missing; run make wasm")
	}
	script := filepath.Join("testdata", "wasm_smoke.js")
	cmd := exec.Command("node", script)
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wasm smoke failed: %v\n%s", err, out)
	}
}
