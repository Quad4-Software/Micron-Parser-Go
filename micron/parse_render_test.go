// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRenderMatchesConvert(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"> Title\n\nBody `!bold`!\n",
		"#!fg=ccc\n#!bg=222\n> Hi\n",
		"- \n-*\n",
		"`= \nliteral `not`\n`=\n",
		"`[label`page.mu]\n",
		"`<name`value>\n",
		"`{partial.mu}\n",
		"# comment\nvisible\n",
	}
	for _, src := range cases {
		for _, dark := range []bool{true, false} {
			for _, mono := range []bool{true, false} {
				p := Parser{DarkTheme: dark, ForceMonospace: mono}
				want := p.ConvertMicronToHTML(src)
				doc := p.Parse(src)
				got := p.RenderHTML(doc)
				if got != want {
					t.Fatalf("dark=%v mono=%v markup=%q\nwant %q\ngot  %q", dark, mono, src, want, got)
				}
			}
		}
	}
}

func TestParseRenderMatchesConvertCorpus(t *testing.T) {
	files := []string{
		filepath.Join("testdata", "nomadnet_guide.mu"),
		filepath.Join("testdata", "nomadnet_guide_official.mu"),
	}
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Skip(path, err)
		}
		p := Parser{DarkTheme: true, ForceMonospace: true}
		want := p.ConvertMicronToHTML(string(src))
		got := p.RenderHTML(p.Parse(string(src)))
		if got != want {
			t.Fatalf("%s: Parse+RenderHTML diverged from ConvertMicronToHTML", path)
		}
	}
}

func TestParseWithDiagnosticsHeader(t *testing.T) {
	p := Parser{}
	_, diags := p.ParseWithDiagnostics("#!fg=zz\n> Hi\n")
	found := false
	for _, d := range diags {
		if d.Code == "header.fg_invalid" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected header.fg_invalid, got %#v", diags)
	}
}
