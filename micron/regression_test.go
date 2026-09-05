// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"encoding/json"
	"html"
	"os"
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

func loadRegressionCases(t *testing.T) []struct {
	name string
	opts regressionOpts
	mu   string
	want []string
} {
	t.Helper()
	var cases []struct {
		name string
		opts regressionOpts
		mu   string
		want []string
	}
	err := withRegressionRoot(func(root *os.Root) error {
		names, err := listRegressionMuNames(root)
		if err != nil {
			return err
		}
		for _, muName := range names {
			base := strings.TrimSuffix(muName, ".mu")
			muBytes, err := readRegressionFile(root, muName)
			if err != nil {
				return err
			}
			txtBytes, err := readRegressionFile(root, base+".txt")
			if err != nil {
				t.Fatalf("regression %s: missing %s.txt", base, base)
			}
			opts := regressionOpts{Dark: true, Mono: false}
			if raw, err := readRegressionFile(root, base+".opts.json"); err == nil {
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
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no regression cases found")
	}
	return cases
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
