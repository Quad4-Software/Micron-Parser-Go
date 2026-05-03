// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNomadNetGuideOfficial renders the upstream NomadNet "Outputting
// Formatted Text" topic and sanity-checks the result. The fixture is a
// committed snapshot in micron/testdata/nomadnet_guide_official.mu, taken
// directly from nomadnet/ui/textui/Guide.py via the sync_nomadnet_guide.py
// helper. NomadNet's MicronParser.py is the canonical source of truth for the
// dialect; matching this corpus exercises the lexical paths that are exercised
// by real NomadNet pages.
func TestNomadNetGuideOfficial(t *testing.T) {
	path := filepath.Join("testdata", "nomadnet_guide_official.mu")
	src, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("snapshot missing; run testdata/sync_nomadnet_guide.py to refresh")
		}
		t.Fatal(err)
	}
	if len(src) == 0 {
		t.Fatal("empty snapshot")
	}
	matrix := []struct {
		dark bool
		mono bool
	}{
		{dark: true, mono: false},
		{dark: true, mono: true},
		{dark: false, mono: false},
		{dark: false, mono: true},
	}
	for _, m := range matrix {
		p := &Parser{DarkTheme: m.dark, ForceMonospace: m.mono}
		out := p.ConvertMicronToHTML(string(src))
		if !strings.HasPrefix(out, `<div style="line-height:1.5;`) {
			t.Fatalf("dark=%v mono=%v: missing container prefix", m.dark, m.mono)
		}
		if !strings.HasSuffix(out, `</div>`) {
			t.Fatalf("dark=%v mono=%v: missing closing div", m.dark, m.mono)
		}
		if strings.Contains(out, "<script") {
			t.Fatalf("dark=%v mono=%v: parser must not emit raw <script>", m.dark, m.mono)
		}
		assertFuzzOutputHTMLSafety(t, out)
	}
}

// TestNomadNetGuideOfficialMatchesJS round-trips the same NomadNet fixture
// through micron-parser-js (when node is available) and compares structural
// metadata (anchor destinations, link field-specs, input names and types).
//
// Whole-text equality is not asserted on this fixture: micron-parser-js uses
// innerHTML+= for literal text, so raw "<x`y>" characters in a literal block
// become real HTML elements (which DOMPurify or the tag stripper then
// removes), whereas the Go renderer always escapes them. Both behaviours are
// safe, but the visible text after tag-strip differs by design.
func TestNomadNetGuideOfficialMatchesJS(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	path := filepath.Join("testdata", "nomadnet_guide_official.mu")
	src, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("snapshot missing; run testdata/sync_nomadnet_guide.py to refresh")
		}
		t.Fatal(err)
	}
	cases := []interopCase{{
		Name:   "nomadnet-guide-official-dark-plain",
		Markup: string(src),
		Dark:   true,
		Mono:   false,
	}}
	jsOutputs := runJSInterop(t, cases)
	if len(jsOutputs) != 1 {
		t.Fatalf("unexpected JS output count: %d", len(jsOutputs))
	}
	p := &Parser{DarkTheme: true, ForceMonospace: false}
	goOut := p.ConvertMicronToHTML(string(src))
	goSig := signatureFromHTML(goOut)
	jsSig := signatureFromHTML(jsOutputs[0])
	mismatches := []string{}
	if !slicesEqualSorted(goSig.Hrefs, jsSig.Hrefs) {
		mismatches = append(mismatches, "Hrefs")
	}
	if !slicesEqualSorted(goSig.Destinations, jsSig.Destinations) {
		mismatches = append(mismatches, "Destinations")
	}
	if !slicesEqualSorted(goSig.Fields, jsSig.Fields) {
		mismatches = append(mismatches, "Fields")
	}
	if !slicesEqualSorted(goSig.InputTypes, jsSig.InputTypes) {
		mismatches = append(mismatches, "InputTypes")
	}
	if !slicesEqualSorted(goSig.InputNames, jsSig.InputNames) {
		mismatches = append(mismatches, "InputNames")
	}
	if len(mismatches) > 0 {
		t.Fatalf("structural mismatch on official NomadNet guide: %v\nGo: %#v\nJS: %#v",
			mismatches, goSig, jsSig)
	}
}

func slicesEqualSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
