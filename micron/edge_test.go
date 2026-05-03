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

func TestBareHeadingNoExtraLine(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := ">>\nbare heading content"
	out := p.ConvertMicronToHTML(in)
	brCount := strings.Count(out, "<br>")
	if brCount != 0 {
		t.Fatalf("expected 0 <br> (bare heading should not emit line), got %d: %s", brCount, out)
	}
	if !strings.Contains(out, "bare heading content") {
		t.Fatalf("expected content to be present: %s", out)
	}
	if !strings.Contains(out, "margin-left:1.2em") {
		t.Fatalf("expected section indent for depth 2: %s", out)
	}
}

func TestLiteralToggleNoExtraLine(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`=\nliteral text\n`="
	out := p.ConvertMicronToHTML(in)
	brCount := strings.Count(out, "<br>")
	if brCount != 0 {
		t.Fatalf("expected 0 <br> (literal toggle should not emit line), got %d: %s", brCount, out)
	}
	if !strings.Contains(out, "literal text") {
		t.Fatalf("expected literal text to be present: %s", out)
	}
}

func TestLiteralToggleTrimmedWhitespace(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "  `=   \npreserve\n  `=  "
	out := p.ConvertMicronToHTML(in)
	if strings.Count(out, "<br>") != 0 {
		t.Fatalf("trimmed toggle lines must not emit blank rows: %s", out)
	}
	if !strings.Contains(out, "preserve") {
		t.Fatal(out)
	}
}

func TestBareHeadingOnlySpaces(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := ">>   \nnext"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "next") || !strings.Contains(out, "margin-left:1.2em") {
		t.Fatal(out)
	}
	if strings.Contains(out, `display:inline-block;width:100%`) {
		t.Fatal("whitespace-only heading must not emit heading block", out)
	}
}

func TestBareHeadingSetsSectionIndent(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := ">>>\nindented content"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "margin-left:2.4em") {
		t.Fatalf("expected section indent for depth 3: %s", out)
	}
}

func TestBareHeadingIssue25Pattern(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := ">Anonymous Node\n\n>>\n`[Node`:/page/index.mu] content"
	out := p.ConvertMicronToHTML(in)

	if !strings.Contains(out, "Anonymous Node") {
		t.Fatalf("expected heading content to be present: %s", out)
	}

	if !strings.Contains(out, "margin-left:1.2em") {
		t.Fatalf("expected bare heading to set section indent for subsequent content: %s", out)
	}

	brCount := strings.Count(out, "<br>")
	if brCount != 2 {
		t.Fatalf("expected 2 <br> (one from content heading, one from empty line), got %d: %s", brCount, out)
	}
}

func TestCyrillicCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "Привет мир \u041D\u043E\u043C\u0430\u0434"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "Привет") || !strings.Contains(out, "\u041D\u043E\u043C\u0430\u0434") {
		t.Fatalf("expected Cyrillic text to be preserved: %s", out)
	}
}

func TestCJKCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u4E2D\u6587\u6D4B\u8BD5 \u65E5\u672C\u8A9E\u30C6\u30B9\u30C8 \uD55C\uAD6D\uC5B4"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u4E2D\u6587") {
		t.Fatalf("expected Chinese text to be preserved: %s", out)
	}
	if !strings.Contains(out, "\u65E5\u672C\u8A9E") {
		t.Fatalf("expected Japanese text to be preserved: %s", out)
	}
	if !strings.Contains(out, "\uD55C\uAD6D\uC5B4") {
		t.Fatalf("expected Korean text to be preserved: %s", out)
	}
}

func TestGreekCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u0395\u03BB\u03BB\u03B7\u03BD\u03B9\u03BA\u03AC \u0391\u0392\u0393"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u0395\u03BB\u03BB\u03B7\u03BD\u03B9\u03BA\u03AC") {
		t.Fatalf("expected Greek text to be preserved: %s", out)
	}
}

func TestArabicCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u0627\u0644\u0639\u0631\u0628\u064A\u0629 \u062A\u062C\u0631\u0628\u0629"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u0627\u0644\u0639\u0631\u0628\u064A\u0629") {
		t.Fatalf("expected Arabic text to be preserved: %s", out)
	}
}

func TestHebrewCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u05E2\u05D1\u05E8\u05D9\u05EA \u05D8\u05E7\u05E1\u05D8"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u05E2\u05D1\u05E8\u05D9\u05EA") {
		t.Fatalf("expected Hebrew text to be preserved: %s", out)
	}
}

func TestThaiCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u0E20\u0E32\u0E29\u0E32\u0E44\u0E17\u0E22 \u0E17\u0E14\u0E2A\u0E2D\u0E1A"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u0E20\u0E32\u0E29\u0E32") {
		t.Fatalf("expected Thai text to be preserved: %s", out)
	}
}

func TestDevanagariCharacters(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "\u0939\u093F\u0928\u094D\u0926\u0940 \u092A\u0930\u0940\u0915\u094D\u0937\u093E"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u0939\u093F\u0928\u094D\u0926\u0940") {
		t.Fatalf("expected Devanagari text to be preserved: %s", out)
	}
}

func TestLineHeightApplied(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "hello world"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "line-height:1.5") {
		t.Fatalf("expected line-height:1.5 to be applied: %s", out)
	}
}

func TestUnicodeWithFormatting(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`!\u4E2D\u6587` `*\u041F\u0440\u0438\u0432\u0435\u0442`*"
	out := p.ConvertMicronToHTML(in)
	if !strings.Contains(out, "\u4E2D\u6587") {
		t.Fatalf("expected Chinese bold text to be preserved: %s", out)
	}
	if !strings.Contains(out, "\u041F\u0440\u0438\u0432\u0435\u0442") {
		t.Fatalf("expected Cyrillic italic text to be preserved: %s", out)
	}
}
