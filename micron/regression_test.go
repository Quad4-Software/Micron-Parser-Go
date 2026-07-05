// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type regressionOpts struct {
	Dark           bool     `json:"dark"`
	Mono           bool     `json:"mono"`
	MustContain    []string `json:"must_contain"`
	MustNotContain []string `json:"must_not_contain"`
}

func regressionDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "regressions")
}

func loadRegressionCases(t *testing.T) []struct {
	name string
	opts regressionOpts
	mu   string
	want []string
} {
	t.Helper()
	dir := regressionDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		name string
		opts regressionOpts
		mu   string
		want []string
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mu") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".mu")
		muBytes, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		txtBytes, err := os.ReadFile(filepath.Join(dir, base+".txt"))
		if err != nil {
			t.Fatalf("regression %s: missing %s.txt", base, base)
		}
		opts := regressionOpts{Dark: true, Mono: false}
		optsPath := filepath.Join(dir, base+".opts.json")
		if raw, err := os.ReadFile(optsPath); err == nil {
			if err := json.Unmarshal(raw, &opts); err != nil {
				t.Fatalf("regression %s: decode opts: %v", base, err)
			}
		}
		var wantLines []string
		for line := range strings.SplitSeq(string(txtBytes), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				wantLines = append(wantLines, line)
			}
		}
		cases = append(cases, struct {
			name string
			opts regressionOpts
			mu   string
			want []string
		}{
			name: base,
			opts: opts,
			mu:   string(muBytes),
			want: wantLines,
		})
	}
	if len(cases) == 0 {
		t.Fatal("no regression cases found")
	}
	return cases
}

func regressionMarkupSeeds(t *testing.T) []string {
	t.Helper()
	dir := regressionDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mu") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, string(raw))
	}
	return out
}

func TestRegressionCorpus(t *testing.T) {
	for _, tc := range loadRegressionCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{DarkTheme: tc.opts.Dark, ForceMonospace: tc.opts.Mono}
			htmlOut := p.ConvertMicronToHTML(tc.mu)
			visible := visibleTextFromHTML(htmlOut)
			compact := strings.ReplaceAll(visible, " ", "")
			for _, frag := range tc.want {
				compactFrag := strings.ReplaceAll(frag, " ", "")
				if !strings.Contains(visible, frag) && !strings.Contains(compact, compactFrag) {
					t.Fatalf("visible text missing %q\ngot: %q\nhtml: %s", frag, visible, htmlOut)
				}
			}
			for _, frag := range tc.opts.MustContain {
				if !strings.Contains(htmlOut, frag) && !strings.Contains(visible, frag) {
					t.Fatalf("output missing required fragment %q\nhtml: %s", frag, htmlOut)
				}
			}
			for _, frag := range tc.opts.MustNotContain {
				if strings.Contains(htmlOut, frag) {
					t.Fatalf("output must not contain %q\nhtml: %s", frag, htmlOut)
				}
			}
		})
	}
}

var visibleTagStripper = regexp.MustCompile(`<[^>]*>`)
var visibleWSCollapse = regexp.MustCompile(`\s+`)

func visibleTextFromHTML(in string) string {
	txt := visibleTagStripper.ReplaceAllString(in, " ")
	txt = html.UnescapeString(txt)
	txt = visibleWSCollapse.ReplaceAllString(txt, " ")
	return strings.TrimSpace(txt)
}
