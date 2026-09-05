// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type specExpect struct {
	Blocks    *int        `json:"blocks"`
	Sections  []int       `json:"sections"`
	Links     *int        `json:"links"`
	Fields    *int        `json:"fields"`
	Partials  *int        `json:"partials"`
	HR        *int        `json:"hr"`
	Headers   *PageColors `json:"headers"`
	LiteralOK *bool       `json:"literal_ok"`
}

func TestSpecConformance(t *testing.T) {
	path := filepath.Join("..", "spec", "micron.txt")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	examples := parseSpecExamples(string(raw))
	if len(examples) == 0 {
		t.Fatal("no spec examples found")
	}
	p := Parser{DarkTheme: true, ForceMonospace: false}
	for i, ex := range examples {
		doc := p.Parse(ex.input)
		var exp specExpect
		if err := json.Unmarshal([]byte(ex.expect), &exp); err != nil {
			t.Fatalf("example %d: bad expect JSON %q: %v", i+1, ex.expect, err)
		}
		if exp.Blocks != nil {
			n := countVisibleBlocks(doc)
			if n != *exp.Blocks {
				t.Fatalf("example %d blocks: got %d want %d\ninput=%q", i+1, n, *exp.Blocks, ex.input)
			}
		}
		if exp.Sections != nil {
			fp := doc.Fingerprint()
			if len(fp.Sections) != len(exp.Sections) {
				t.Fatalf("example %d sections len: got %v want %v", i+1, fp.Sections, exp.Sections)
			}
			for j := range exp.Sections {
				if fp.Sections[j] != exp.Sections[j] {
					t.Fatalf("example %d section[%d]: got %d want %d", i+1, j, fp.Sections[j], exp.Sections[j])
				}
			}
		}
		if exp.Links != nil {
			if n := len(doc.Fingerprint().Links); n != *exp.Links {
				t.Fatalf("example %d links: got %d want %d", i+1, n, *exp.Links)
			}
		}
		if exp.Fields != nil {
			if n := len(doc.Fingerprint().Fields); n != *exp.Fields {
				t.Fatalf("example %d fields: got %d want %d", i+1, n, *exp.Fields)
			}
		}
		if exp.Partials != nil {
			if n := len(doc.Fingerprint().Partials); n != *exp.Partials {
				t.Fatalf("example %d partials: got %d want %d", i+1, n, *exp.Partials)
			}
		}
		if exp.HR != nil {
			n := 0
			for _, b := range doc.Blocks {
				if b.Kind == BlockHR {
					n++
				}
			}
			if n != *exp.HR {
				t.Fatalf("example %d hr: got %d want %d", i+1, n, *exp.HR)
			}
		}
		if exp.Headers != nil {
			if doc.Colors.FG != exp.Headers.FG || doc.Colors.BG != exp.Headers.BG {
				t.Fatalf("example %d headers: got %+v want %+v", i+1, doc.Colors, *exp.Headers)
			}
		}
		if exp.LiteralOK != nil && *exp.LiteralOK {
			html := p.RenderHTML(doc)
			if strings.Contains(html, "<script") {
				t.Fatalf("example %d literal produced script tag", i+1)
			}
		}
	}
}

type specExample struct {
	input  string
	expect string
}

func parseSpecExamples(src string) []specExample {
	const fence = "````````````````````````"
	var out []specExample
	lines := strings.Split(src, "\n")
	i := 0
	for i < len(lines) {
		if strings.HasPrefix(lines[i], fence+" example") {
			i++
			var input strings.Builder
			for i < len(lines) && lines[i] != "." {
				if input.Len() > 0 {
					input.WriteByte('\n')
				}
				input.WriteString(lines[i])
				i++
			}
			if i >= len(lines) || lines[i] != "." {
				break
			}
			i++
			var expect strings.Builder
			for i < len(lines) && !strings.HasPrefix(lines[i], fence) {
				if expect.Len() > 0 {
					expect.WriteByte('\n')
				}
				expect.WriteString(lines[i])
				i++
			}
			out = append(out, specExample{input: input.String(), expect: expect.String()})
			if i < len(lines) && strings.HasPrefix(lines[i], fence) {
				i++
			}
			continue
		}
		i++
	}
	return out
}

func countVisibleBlocks(doc *Document) int {
	n := 0
	for _, b := range doc.Blocks {
		switch b.Kind {
		case BlockParagraph, BlockHeading, BlockHR, BlockDivider, BlockPartial:
			n++
		case BlockBlank:
			// blank lines still render as br but spec "visible content" counts content blocks
		}
	}
	return n
}
