// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"html"
	"regexp"
	"strings"
	"testing"
)

var stripHTMLTags = regexp.MustCompile(`<[^>]*>`)

func htmlPlainText(s string) string {
	t := stripHTMLTags.ReplaceAllString(s, "")
	return html.UnescapeString(t)
}

func TestFormatTableRawBlockAlignAndBorders(t *testing.T) {
	lines := formatTableRaw([]string{
		"| Name | Price | Qty |",
		"| ---- | :---: | --: |",
		"| `F3a3Apple`f | Free | `!5`! |",
		"| Orange | Ask, nicely | 3 |",
	}, "c", 100)
	if len(lines) < 4 {
		t.Fatalf("short lines: %d %#v", len(lines), lines)
	}
	if lines[0] != "`c" {
		t.Fatalf("want block align line `c, got %q", lines[0])
	}
	if lines[len(lines)-1] != "`a" {
		t.Fatalf("want closing `a, got %q", lines[len(lines)-1])
	}
	if !strings.Contains(lines[1], "┌") || !strings.Contains(lines[1], "┐") {
		t.Fatalf("top border: %q", lines[1])
	}
	if !strings.Contains(lines[2], "Name") || !strings.Contains(lines[2], "│") {
		t.Fatalf("header: %q", lines[2])
	}
}

func TestConvertMicronTableGuideExample(t *testing.T) {
	src := "`t\n" +
		"| Name | Price | Qty |\n" +
		"| ---- | :---: | --: |\n" +
		"| `F3a3Apple`f | Free | `!5`! |\n" +
		"| Orange | Ask, nicely | 3 |\n" +
		"`t\n"
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML(src)
	plain := htmlPlainText(out)
	for _, want := range []string{"Name", "Price", "Qty", "Apple", "Free", "Orange", "Ask, nicely"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("missing %q in plain %q", want, plain)
		}
	}
	if !strings.Contains(out, "│") || !strings.Contains(out, "┌") {
		t.Fatalf("expected box drawing in HTML: %s", out)
	}
	if !strings.Contains(out, "font-weight:bold") {
		t.Fatalf("expected bold from `!5`!: %s", out)
	}
}

func TestConvertMicronTableNarrowMaxWidth(t *testing.T) {
	src := "`t20\n" +
		"| a | bbbbbbbbbbbbbb |\n" +
		"| --- | --- |\n" +
		"| x | y |\n" +
		"`t\n"
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(src)
	plain := htmlPlainText(out)
	if !strings.Contains(plain, "x") || !strings.Contains(plain, "y") {
		t.Fatal(plain)
	}
}

func TestConvertMicronTableBlockLeftAlign(t *testing.T) {
	src := "`tl\n| a | b |\n| - | - |\n| 1 | 2 |\n`t\n"
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(src)
	if !strings.Contains(out, `text-align:left`) {
		t.Fatalf("want left align wrapper from `tl: %s", out)
	}
}

func TestConvertMicronTableShortBufferNoOutput(t *testing.T) {
	src := "`t\n| only |\n`t\n"
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(src)
	if strings.Contains(out, "only") {
		t.Fatalf("invalid table should not render pipe row as plain: %s", out)
	}
}

func TestParseTableFenceOptions(t *testing.T) {
	a, w := parseTableFenceOptions("")
	if a != "" || w != 0 {
		t.Fatal(a, w)
	}
	a, w = parseTableFenceOptions("c30")
	if a != "c" || w != 30 {
		t.Fatal(a, w)
	}
	a, w = parseTableFenceOptions("40")
	if a != "" || w != 40 {
		t.Fatal(a, w)
	}
}
