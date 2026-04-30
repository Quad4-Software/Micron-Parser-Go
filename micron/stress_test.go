// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func buildMassiveSyntheticMarkup(blocks int) string {
	var b strings.Builder
	b.Grow(blocks * 220)
	b.WriteString("#!fg=ddd\n#!bg=111\n")
	for range blocks {
		b.WriteString("> Section ")
		b.WriteString(strings.Repeat("x", 6))
		b.WriteString("\n")
		b.WriteString("`!bold` `_underline` `*italic` text block ")
		b.WriteString(strings.Repeat("data ", 4))
		b.WriteString("\n")
		b.WriteString("`<24|field_name`field_value>\n")
		b.WriteString("`[open`node.example`field_name|q=1|mode=stress]\n")
		b.WriteString("-\n")
	}
	return b.String()
}

func TestStressMassiveDocument(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	markup := buildMassiveSyntheticMarkup(1800)
	out := p.ConvertMicronToHTML(markup)
	if len(out) < len(markup) {
		t.Fatalf("unexpected output size: out=%d in=%d", len(out), len(markup))
	}
	if !strings.Contains(out, `data-action="openNode"`) || !strings.Contains(out, `type="text"`) {
		t.Fatal("stress output missing expected link/field HTML")
	}
	out2 := p.ConvertMicronToHTML(markup)
	if out != out2 {
		t.Fatal("non-deterministic output on repeated stress conversion")
	}
}

func TestStressHugeLiteralLine(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	huge := strings.Repeat("abc <>& \" ' xyz ", 5000)
	in := "`=\n" + huge + "\n`=\n"
	out := p.ConvertMicronToHTML(in)
	if len(out) < 1000 {
		t.Fatalf("unexpected short stress output: %d", len(out))
	}
	if !strings.Contains(out, "&lt;") || !strings.Contains(out, "&amp;") {
		t.Fatal("expected escaped literal content missing in stress output")
	}
}
