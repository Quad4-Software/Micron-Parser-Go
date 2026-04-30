// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestConvertEscapesText(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("a & b < c")
	if !strings.Contains(out, "&amp;") || !strings.Contains(out, "&lt;") {
		t.Fatal(out)
	}
}

func TestCommentLine(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("hello\n# secret\nworld")
	if strings.Contains(out, "secret") {
		t.Fatal(out)
	}
}

func TestHeadingAndHR(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(">> title\n-\n--")
	if !strings.Contains(out, "title") || !strings.Contains(out, "<hr") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "white-space:nowrap") {
		t.Fatal(out)
	}
}

func TestFormatting(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("x `!b` `*i` `_u` `f`plain")
	if !strings.Contains(out, "font-weight:bold") || !strings.Contains(out, "italic") {
		t.Fatal(out)
	}
}

func TestTruecolor(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`FTaabbcchello `BT112233world")
	if !strings.Contains(out, "#aabbcc") || !strings.Contains(out, "#112233") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Fatal(out)
	}
}

func TestFieldAndLink(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`" + "<24|name`val>" + " " + "`" + "[x`example.com`f1]")
	if !strings.Contains(out, `type="text"`) || !strings.Contains(out, `name="name"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `href="nomadnetwork://example.com"`) || !strings.Contains(out, `data-fields="f1"`) {
		t.Fatal(out)
	}
}

func TestLiteralToggle(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`=\na\n`=\n`!b")
	if !strings.Contains(out, "font-weight:bold") {
		t.Fatal(out)
	}
}

func TestPageWrap(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("#!fg=111\n#!bg=222\nok")
	if !strings.Contains(out, "color:#111") || !strings.Contains(out, "background-color:#222") {
		t.Fatal(out)
	}
}

func TestLinkFieldVariableOrderPreserved(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`" + "[x`u`z=1|y=2|f]")
	if !strings.Contains(out, "data-destination") {
		t.Fatal(out)
	}
	iy := strings.Index(out, "y=2")
	iz := strings.Index(out, "z=1")
	if iy < 0 || iz < 0 || iz > iy {
		t.Fatal("expected original order z=1 then y=2 in query", out)
	}
}

func TestDepthReset(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(">> a\n< b")
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatal(out)
	}
}

func TestCheckboxRadio(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`" + "<?|g`L`>" + "`" + "<^|g`R`>")
	if !strings.Contains(out, `type="checkbox"`) || !strings.Contains(out, `type="radio"`) {
		t.Fatal(out)
	}
}

func TestPartialTag(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`{f64a:/page/partial_1.mu}")
	if !strings.Contains(out, `class="Mu-partial"`) || !strings.Contains(out, `data-partial-url="nomadnetwork://f64a:/page/partial_1.mu"`) {
		t.Fatal(out)
	}
}

func TestPartialTagWithRefresh(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`{f64a:/page/refreshing_partial.mu`10}")
	if !strings.Contains(out, `class="Mu-partial"`) || !strings.Contains(out, `data-partial-refresh="10"`) {
		t.Fatal(out)
	}
}
