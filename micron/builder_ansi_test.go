// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestBuilderRoundTrip(t *testing.T) {
	var b Builder
	b.HeaderColors("ccc", "222").
		Heading(1, "Title").
		Text("Hello ").Bold("world").NL().
		Link("Go", "page.mu").NL().
		FieldText("user", "alice", 12).NL().
		HR()
	src := b.String()
	p := Parser{DarkTheme: true}
	doc := p.Parse(src)
	if len(doc.Fingerprint().Sections) != 1 {
		t.Fatalf("sections=%v", doc.Fingerprint().Sections)
	}
	if len(doc.Fingerprint().Links) != 1 {
		t.Fatalf("links=%v", doc.Fingerprint().Links)
	}
	html := p.RenderHTML(doc)
	if !strings.Contains(html, "alice") {
		t.Fatalf("missing field value in html: %s", html)
	}
}

func TestRenderANSI(t *testing.T) {
	p := Parser{}
	doc := p.Parse("> Hi\n`!bold`! text\n")
	out := p.RenderANSI(doc)
	if !strings.Contains(out, "Hi") || !strings.Contains(out, "bold") {
		t.Fatalf("ansi=%q", out)
	}
}

func TestParseIncrementalReuse(t *testing.T) {
	p := Parser{}
	prev := p.Parse("line1\nline2\n")
	next := p.ParseIncremental(prev, "line1\nline2\nline3\n")
	if len(next.Blocks) < len(prev.Blocks) {
		t.Fatalf("expected more blocks got %d prev %d", len(next.Blocks), len(prev.Blocks))
	}
	html := p.RenderHTML(next)
	want := p.ConvertMicronToHTML("line1\nline2\nline3\n")
	if html != want {
		t.Fatalf("incremental render mismatch\nwant %q\ngot  %q", want, html)
	}
}
