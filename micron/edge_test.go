// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestEdgeLargeInputNoPanic(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	chunk := strings.Repeat("abcd `!x` ", 2048)
	in := strings.Join([]string{"> title", chunk, "-=", "`B333", chunk, "`b"}, "\n")
	out := p.ConvertMicronToHTML(in)
	if len(out) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestEdgeMalformedFieldAndLinkFallbackToText(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`<broken-field " + "`[broken-link"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "broken-field") || !strings.Contains(out, "broken-link") {
		t.Fatal(out)
	}
}

func TestEdgeLiteralModePreservesFormattingChars(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`=\n`!`_`*`F123 not parsed\n`="
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "`!`_`*`F123 not parsed") {
		t.Fatal(out)
	}
}
