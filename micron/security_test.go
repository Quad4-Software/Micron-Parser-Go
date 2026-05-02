// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestSecurityEscapesRawHTMLPayload(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML(`<img src=x onerror=alert(1)><script>alert(1)</script>`)
	if strings.Contains(out, "<script") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "&lt;script") {
		t.Fatal(out)
	}
}

func TestSecurityEscapesLinkAndFieldAttributes(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`" + "<24|x\" autofocus`v\">" + " " + "`" + "[ok`example.com?x=\"y\"`q=\"x\"|f]"
	out := p.ConvertMicronToHTML(in)
	if strings.Contains(out, `name="x" autofocus`) {
		t.Fatal(out)
	}
	if strings.Contains(out, `" on`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, "&#34;") {
		t.Fatal(out)
	}
}

func TestSecurityKeepsMarkupInsideTextEscaped(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	out := p.ConvertMicronToHTML("normal <b>tag</b> `!not bold")
	if strings.Contains(out, "<b>") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "&lt;b&gt;tag&lt;/b&gt;") {
		t.Fatal(out)
	}
}

func TestSecurityPlainLineAngleBracketsEscapedAllModes(t *testing.T) {
	in := "<svg><script>alert(1)</script></svg>"
	for _, p := range []Parser{
		{DarkTheme: true, ForceMonospace: true},
		{DarkTheme: true, ForceMonospace: false},
		{DarkTheme: false, ForceMonospace: true},
		{DarkTheme: false, ForceMonospace: false},
	} {
		out := p.ConvertMicronToHTML(in)
		if strings.Contains(strings.ToLower(out), "<script") {
			t.Fatalf("parser %#v leaked script tag: %s", p, out)
		}
		assertFuzzOutputHTMLSafety(t, out)
	}
}

func TestSecurityMalformedMarkupStillBalancesAngles(t *testing.T) {
	in := "<<<>><<div>>>>>>"
	for _, p := range []Parser{
		{DarkTheme: true, ForceMonospace: true},
		{DarkTheme: false, ForceMonospace: false},
	} {
		out := p.ConvertMicronToHTML(in)
		if strings.Count(out, "<") != strings.Count(out, ">") {
			t.Fatalf("parser %#v: %q", p, out)
		}
		assertFuzzOutputHTMLSafety(t, out)
	}
}

func TestSecurityDangerousJavascriptSlashSlashHrefNeutralized(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`[x`javascript://alert(1)`]"
	out := p.ConvertMicronToHTML(in)
	if strings.Contains(strings.ToLower(out), `href="javascript:`) {
		t.Fatalf("javascript URL must not appear raw in href: %s", out)
	}
	if !strings.Contains(out, `href="nomadnetwork://javascript://alert(1)"`) {
		t.Fatalf("expected nomadnetwork-wrapped javascript URL in href: %s", out)
	}
	assertFuzzOutputHTMLSafety(t, out)
}

func TestSecurityDangerousVBScriptSlashSlashHrefNeutralized(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`[x`vbscript://msgbox(1)`]"
	out := p.ConvertMicronToHTML(in)
	if strings.Contains(strings.ToLower(out), `href="vbscript:`) {
		t.Fatalf("vbscript URL must not appear raw in href: %s", out)
	}
	assertFuzzOutputHTMLSafety(t, out)
}

func TestSecurityFileSchemeHrefNeutralized(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`[local`file:///etc/passwd`]"
	out := p.ConvertMicronToHTML(in)
	if strings.Contains(strings.ToLower(out), `href="file:`) {
		t.Fatalf("file URL must not appear raw in href: %s", out)
	}
	assertFuzzOutputHTMLSafety(t, out)
}

func TestSecurityHTMLAttrsStripAsciiControls(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := "`[x`http://evil.example/\x00path\nmore`]"
	out := p.ConvertMicronToHTML(in)
	if strings.Contains(out, "\x00") {
		t.Fatal("NUL must not appear in rendered HTML output")
	}
	if strings.Contains(out, "\n") && strings.Contains(out, `href="`) {
		t.Fatal("ASCII newline must not appear inside emitted attributes")
	}
}

func TestSecurityMalformedFormattingConstructsDoNotPanic(t *testing.T) {
	inputs := []string{
		"`[`no-close-bracket",
		"`{no-close-brace",
		"`[`]",
		"`<<||||>",
	}
	for _, in := range inputs {
		for _, p := range []Parser{
			{DarkTheme: true, ForceMonospace: true},
			{DarkTheme: true, ForceMonospace: false},
		} {
			out := p.ConvertMicronToHTML(in)
			assertFuzzOutputHTMLSafety(t, out)
			if strings.Count(out, "<") != strings.Count(out, ">") {
				t.Fatalf("%q parser %#v unbalanced: %q", in, p, out)
			}
		}
	}
}
