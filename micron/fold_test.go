// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestFoldingHeadings(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	markup := "#!fold v >\n`+> Open section\nvisible\n`-> Closed section\nhidden\n"
	out := p.ConvertMicronToHTML(markup)
	if !strings.Contains(out, "<details") {
		t.Fatalf("expected <details> tags for collapsible headings, got: %s", out)
	}
	if !strings.Contains(out, `open data-mu-fold="open"`) {
		t.Fatalf("expected an open details for `+>, got: %s", out)
	}
	if !strings.Contains(out, `<details class="Mu-fold" data-mu-fold="collapsed">`) {
		t.Fatalf("expected a collapsed details for `->, got: %s", out)
	}
	if !strings.Contains(out, "<span class=\"Mu-fold-glyph\">v</span>") {
		t.Fatalf("expected custom fold glyph from #!fold, got: %s", out)
	}
}

func TestFoldingNesting(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	markup := "`+> Parent\n`+> Child\ncontent\n<\n"
	out := p.ConvertMicronToHTML(markup)
	openCount := strings.Count(out, `<details`)
	closeCount := strings.Count(out, `</details>`)
	if openCount != closeCount {
		t.Fatalf("mismatched details tags: open=%d close=%d in %s", openCount, closeCount, out)
	}
	if openCount != 2 {
		t.Fatalf("expected 2 nested details, got %d: %s", openCount, out)
	}
}
