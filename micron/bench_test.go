// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkConvertMicronToHTML(b *testing.B) {
	const in = "> Title\n-∿\n`!bold` `*italic*\n[`x`example.com`f1]"
	p := Parser{DarkTheme: true, ForceMonospace: true}
	for b.Loop() {
		_ = p.ConvertMicronToHTML(in)
	}
}

// BenchmarkConvertNomadNetGuide uses testdata/nomadnet_guide.mu, the same corpus as the web demo seed.
// It measures native Go (the same code as the WASM build). Example: go test ./micron -bench=BenchmarkConvertNomadNetGuide -benchmem -count=10
func BenchmarkConvertNomadNetGuide(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "nomadnet_guide.mu"))
	if err != nil {
		b.Fatal(err)
	}
	markup := string(data)
	p := Parser{DarkTheme: true, ForceMonospace: true}
	b.SetBytes(int64(len(markup)))
	b.ReportMetric(float64(len(markup)), "B/input")
	b.ResetTimer()
	for b.Loop() {
		_ = p.ConvertMicronToHTML(markup)
	}
}

func BenchmarkConvertMassiveSynthetic(b *testing.B) {
	markup := buildMassiveSyntheticMarkup(2200)
	p := Parser{DarkTheme: true, ForceMonospace: true}
	b.SetBytes(int64(len(markup)))
	b.ReportMetric(float64(len(markup)), "B/input")
	b.ResetTimer()
	for b.Loop() {
		_ = p.ConvertMicronToHTML(markup)
	}
}

func TestNomadNetGuideCorpus(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "nomadnet_guide.mu"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Fatalf("corpus too small: %d bytes", len(data))
	}
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML(string(data))
	if !strings.Contains(out, "<") || len(out) < 100 {
		t.Fatal("unexpected short or empty HTML output")
	}
}

func TestConvertStable(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("a\nb")
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatal(out)
	}
}
