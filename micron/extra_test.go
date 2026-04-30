// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestForceMonospaceSpans(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("hi there")
	if !strings.Contains(out, `class="Mu-mws"`) || !strings.Contains(out, `class="Mu-mnt"`) {
		t.Fatal(out)
	}
}

func TestLightThemeHeading(t *testing.T) {
	p := Parser{DarkTheme: false, ForceMonospace: false}
	out := p.ConvertMicronToHTML("> H")
	if !strings.Contains(out, "background-color:#777") {
		t.Fatal(out)
	}
}

func TestMaskedField(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`" + "<!|n`d`>")
	if !strings.Contains(out, `type="password"`) {
		t.Fatal(out)
	}
}

func TestEmptyLineWithBG(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`B999`\n")
	if !strings.Contains(out, "height:1.2em") {
		t.Fatal(out)
	}
}

func TestDepthIndent(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(">> x")
	if !strings.Contains(out, "margin-left:") {
		t.Fatal(out)
	}
}

func TestLinkNoFieldsBranch(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`" + "[x`only.com`]")
	if !strings.Contains(out, `data-action="openNode"`) || strings.Contains(out, `data-fields`) {
		t.Fatal(out)
	}
}

func TestFThreeCharFG(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`F123hello")
	if !strings.Contains(out, "#123") {
		t.Fatal(out)
	}
}

func TestDividerUnicode(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("-∿")
	if !strings.Contains(out, "white-space:nowrap") {
		t.Fatal(out)
	}
}

func TestDoubleBacktickReset(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("`!a``b")
	if !strings.Contains(out, "b") {
		t.Fatal(out)
	}
}
