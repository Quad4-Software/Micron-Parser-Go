// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestForceMonospaceSpans(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	plain := p.ConvertMicronToHTML("hi there")
	if strings.Contains(plain, `class="Mu-mnt"`) || strings.Contains(plain, `class="Mu-mws"`) {
		t.Fatalf("plain ASCII must stay bare under ForceMonospace: %s", plain)
	}
	special := p.ConvertMicronToHTML("a<b")
	if !strings.Contains(special, `class="Mu-mws"`) || !strings.Contains(special, `class="Mu-mnt"`) {
		t.Fatalf("HTML-special ASCII must use Mu-mws/Mu-mnt: %s", special)
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
	if !strings.Contains(out, "margin-inline-start:") {
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

func TestForceMonospaceLinkLabel(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("`!`[" + "  • public`:/page/group.mu`g=public]`! (17 repositories)")
	if strings.Contains(out, "&lt;span class=&#34;Mu-mws&#34;&gt;") {
		t.Fatalf("link label must not escape ForceMonospace spans: %s", out)
	}
	if !strings.Contains(out, `<a class="Mu-nl"`) {
		t.Fatalf("expected a Mu-nl link: %s", out)
	}
	if !strings.Contains(out, `<span class="Mu-mnt">•</span>`) {
		t.Fatalf("expected Mu-mnt bullet inside link label: %s", out)
	}
}

func TestCincinnatusAsciiColor(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("`FEE0\n" + " █████╗")
	if !strings.Contains(out, "color:#EE0") && !strings.Contains(out, "color:#ee0") {
		t.Fatalf("expected #EE0 foreground for FEE0: %s", out)
	}
	if !strings.Contains(out, `<span class="Mu-mnt">█</span>`) {
		t.Fatalf("expected Mu-mnt cell for block-drawing char: %s", out)
	}
}
