// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// Go cannot assert pixel kerning, but it can assert ForceMonospace does not
// slice complex scripts into per-glyph Mu-mnt cells (which collide or gap) and
// that page-level HR rules do not paint a box-shadow halo into the host gutter.

func TestLanguagesTemplateShapingRuns(t *testing.T) {
	path := filepath.Join("testdata", "languages.mu")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML(string(src))

	// Complex scripts stay as contiguous UTF-8 runs (not Mu-mnt cells).
	mustContain := []string{
		"こんにちは",
		"你好",
		"안녕하세요",
		"سلام",
		"مرحبا",
		"שלום",
		"नमस्ते",
		"สวัสดี",
		"Привет",
	}
	for _, frag := range mustContain {
		if !strings.Contains(out, frag) {
			t.Fatalf("missing %q in languages HTML", frag)
		}
		// A multi-rune complex word must not start inside a Mu-mnt cell.
		if utf8.RuneCountInString(frag) > 1 {
			r, _ := utf8.DecodeRuneInString(frag)
			bad := `<span class="Mu-mnt">` + string(r)
			if strings.Contains(out, bad) {
				t.Fatalf("%q must not be splintered into Mu-mnt (found %q)", frag, bad)
			}
		}
	}
	if !strings.Contains(stripTags(out), "The quick brown fox") {
		t.Fatalf("missing English body text in languages HTML")
	}
	if !strings.Contains(out, `dir="auto"`) {
		t.Fatalf("expected dir=auto root for RTL scripts")
	}
}

func TestLanguagesTemplateHRNoPageBGHalo(t *testing.T) {
	src := "#!fg=e8e8e8\n#!bg=1a1a1a\n> Title\n-\nBody\n"
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML(src)
	if strings.Contains(out, "box-shadow:0 0 0 0.5em") {
		t.Fatalf("HR must not box-shadow when bg equals page default: %s", out)
	}
	if !strings.Contains(out, "<hr") {
		t.Fatalf("expected hr element: %s", out)
	}
}

func TestHRBoxShadowOnlyWhenBGDiffers(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	same := p.ConvertMicronToHTML("#!bg=222\n-\n")
	if strings.Contains(same, "box-shadow:") {
		t.Fatalf("no halo when hr bg matches page bg: %s", same)
	}
	doc := &Document{
		DefaultFG: "ddd",
		DefaultBG: "222",
		Blocks: []Block{{
			Kind:         BlockHR,
			HeadingStyle: Style{FG: "ddd", BG: "444"},
		}},
	}
	html := p.RenderHTML(doc)
	if !strings.Contains(html, "box-shadow:0 0 0 0.5em") {
		t.Fatalf("expected halo when hr bg differs: %s", html)
	}
}

func TestCJKHiraganaHangulStayUnsplintered(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	cases := []string{
		"日本語",
		"中文汉字",
		"한국어",
		"こんにちは",
	}
	for _, in := range cases {
		out := p.ConvertMicronToHTML(in)
		if !strings.Contains(out, in) {
			t.Fatalf("missing continuous text %q in %s", in, out)
		}
		if strings.Contains(out, `class="Mu-mnt"`) {
			t.Fatalf("%q must not be split into Mu-mnt cells: %s", in, out)
		}
	}
}

func TestThaiDevanagariStayUnsplintered(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	thai := "ภาษาไทย"
	hindi := "हिन्दी"
	for _, in := range []string{thai, hindi} {
		out := p.ConvertMicronToHTML(in)
		if !strings.Contains(out, in) {
			t.Fatalf("missing continuous text %q in %s", in, out)
		}
		if strings.Contains(out, `class="Mu-mnt"`) {
			t.Fatalf("%q must not be split into Mu-mnt cells: %s", in, out)
		}
	}
}

func TestMixedLatinCJKMonospace(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("AB日本語CD")
	if countMuMnt(out) != 0 {
		t.Fatalf("Latin stays bare and CJK stays unsplintered, want 0 Mu-mnt got %d in %s", countMuMnt(out), out)
	}
	if !strings.Contains(out, "日本語") {
		t.Fatalf("CJK run missing: %s", out)
	}
	if !strings.Contains(out, `class="Mu-mws"`) {
		t.Fatalf("mixed Latin/CJK word must use Mu-mws: %s", out)
	}
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for i := range len(s) {
		c := s[i]
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteByte(c)
		}
	}
	return b.String()
}
