// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestAccessibilityHeadingRole(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("> Title\n")
	if !strings.Contains(out, `role="heading"`) {
		t.Fatalf("expected role=heading, got: %s", out)
	}
	if !strings.Contains(out, `aria-level="1"`) {
		t.Fatalf("expected aria-level=1, got: %s", out)
	}
	if strings.Count(out, `role="heading"`) != 1 {
		t.Fatalf("expected exactly one heading role, got: %s", out)
	}
}

func TestAccessibilityHeadingLevel(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML(">>> Deep\n")
	if !strings.Contains(out, `aria-level="3"`) {
		t.Fatalf("expected aria-level=3, got: %s", out)
	}
}

func TestAccessibilityOffByDefault(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("> Title\n-=\nx\n")
	for _, frag := range []string{`role="heading"`, `aria-level`, `role="separator"`, `Mu-sr`, `aria-live`} {
		if strings.Contains(out, frag) {
			t.Fatalf("expected no %q without Accessibility, got: %s", frag, out)
		}
	}
}

func TestAccessibilityDividerSeparator(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("-=\n")
	if !strings.Contains(out, `role="separator"`) {
		t.Fatalf("expected role=separator on divider, got: %s", out)
	}
}

func TestAccessibilityFoldGlyphHidden(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("`+> Fold\nbody\n`-\n")
	if !strings.Contains(out, `class="Mu-fold-glyph" aria-hidden="true"`) {
		t.Fatalf("expected aria-hidden fold glyph, got: %s", out)
	}
}

func TestAccessibilityPartialStatus(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("`{file.mu`30`a=b}\n")
	if !strings.Contains(out, `role="status"`) {
		t.Fatalf("expected role=status on partial, got: %s", out)
	}
	if !strings.Contains(out, `aria-live="polite"`) {
		t.Fatalf("expected aria-live=polite on partial, got: %s", out)
	}
}

func TestAccessibilityReadableText(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("hello world\n")
	if !strings.Contains(out, `class="Mu-sr"`) {
		t.Fatalf("expected Mu-sr readable span, got: %s", out)
	}
	if !strings.Contains(out, `aria-hidden="true"`) {
		t.Fatalf("expected aria-hidden visual span, got: %s", out)
	}
}

func TestAccessibilityFieldLabel(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Accessibility: true}
	out := p.ConvertMicronToHTML("`<24|username`alice>\n")
	if !strings.Contains(out, `aria-label="username"`) {
		t.Fatalf("expected aria-label on field, got: %s", out)
	}
	if !strings.Contains(out, `name="username"`) {
		t.Fatalf("expected name attribute preserved, got: %s", out)
	}
}

func TestFieldWidthAndRows(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("`<40|name`bob>\n")
	if !strings.Contains(out, `size="40"`) {
		t.Fatalf("expected size=40 on input, got: %s", out)
	}
	if !strings.Contains(out, "<input") {
		t.Fatalf("expected input element, got: %s", out)
	}
}

func TestFieldRowsTextarea(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("`<20x3|bio`hello>\n")
	if !strings.Contains(out, "<textarea") {
		t.Fatalf("expected textarea for WxR field, got: %s", out)
	}
	if !strings.Contains(out, `rows="3"`) {
		t.Fatalf("expected rows=3, got: %s", out)
	}
	if !strings.Contains(out, `cols="20"`) {
		t.Fatalf("expected cols=20, got: %s", out)
	}
	if !strings.Contains(out, `name="bio"`) {
		t.Fatalf("expected name=bio, got: %s", out)
	}
}

func TestHeadingAnchorSlug(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("> My Section!\n")
	if !strings.Contains(out, `id="my-section"`) {
		t.Fatalf("expected slug anchor id=my-section, got: %s", out)
	}
	if !strings.Contains(out, `micron-header-anchor`) {
		t.Fatalf("expected micron-header-anchor class, got: %s", out)
	}
}

func TestExplicitAnchor(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("text `:target` more\n")
	if !strings.Contains(out, `id="target"`) {
		t.Fatalf("expected explicit anchor id=target, got: %s", out)
	}
	if !strings.Contains(out, `class="micron-anchor"`) {
		t.Fatalf("expected micron-anchor class, got: %s", out)
	}
}

func TestEmptyAnchorResolvesToNextHeading(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("`[next`#]\n> Second\n")
	if !strings.Contains(out, `href="#second"`) {
		t.Fatalf("expected href resolved to #second, got: %s", out)
	}
	if strings.Contains(out, `href="#"`) {
		t.Fatalf("expected no unresolved empty anchor, got: %s", out)
	}
}

func TestEmptyAnchorUnmatchedStays(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("> First\n`[next`#]\n")
	if !strings.Contains(out, `href="#"`) {
		t.Fatalf("expected unresolved empty anchor to stay href=#, got: %s", out)
	}
}

func TestSerifRootClass(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true, Serif: true}
	out := p.ConvertMicronToHTML("body\n")
	if !strings.Contains(out, `Mu-serif`) {
		t.Fatalf("expected Mu-serif root class, got: %s", out)
	}
}

func TestSerifOffByDefault(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("body\n")
	if strings.Contains(out, `Mu-serif`) {
		t.Fatalf("expected no Mu-serif without Serif, got: %s", out)
	}
}

func TestCRLFNormalization(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	crlf := p.ConvertMicronToHTML("> Title\r\nbody\r\n")
	lf := p.ConvertMicronToHTML("> Title\nbody\n")
	if crlf != lf {
		t.Fatalf("CRLF and LF input must render identically.\nCRLF: %s\nLF: %s", crlf, lf)
	}
}

func TestHashLinkNotNomadWrapped(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML("> Target\n`[back`#]\n")
	if strings.Contains(out, `href="nomadnetwork://#"`) {
		t.Fatalf("expected bare # href, got nomadnetwork-wrapped: %s", out)
	}
}
