// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"bytes"
	"encoding/json"
	"html"
	"maps"
	"math/rand"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type interopCase struct {
	Name   string `json:"name"`
	Markup string `json:"markup"`
	Dark   bool   `json:"dark"`
	Mono   bool   `json:"mono"`
}

type htmlSig struct {
	TagCount         map[string]int
	TextNormalized   string
	Hrefs            []string
	Destinations     []string
	Fields           []string
	InputTypes       []string
	InputNames       []string
	BoldCount        int
	UnderlineCount   int
	ItalicCount      int
	HeadingBlockUsed bool
}

func TestInteropWithReferenceJS(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	cases := interopCorpus()
	jsOutputs := runJSInterop(t, cases)
	if len(jsOutputs) != len(cases) {
		t.Fatalf("interop output mismatch: got=%d want=%d", len(jsOutputs), len(cases))
	}
	for i, tc := range cases {
		p := &Parser{DarkTheme: tc.Dark, ForceMonospace: tc.Mono}
		goOut := p.ConvertMicronToHTML(tc.Markup)
		jsOut := jsOutputs[i]
		goSig := signatureFromHTML(goOut)
		jsSig := signatureFromHTML(jsOut)
		if !sigsEqual(goSig, jsSig) {
			t.Fatalf("interop mismatch on %s\nGo: %#v\nJS: %#v\nGo HTML: %s\nJS HTML: %s",
				tc.Name, goSig, jsSig, goOut, jsOut)
		}
	}
}

func TestInteropRandomizedDeterministic(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	rng := rand.New(rand.NewSource(1337))
	cases := make([]interopCase, 0, 120)
	for i := range 120 {
		cases = append(cases, interopCase{
			Name:   "rnd-" + boolName(i%2 == 0, "a", "b") + "-" + boolName(i%3 == 0, "x", "y"),
			Markup: randomMarkup(rng),
			Dark:   i%2 == 0,
			Mono:   i%3 == 0,
		})
	}
	jsOutputs := runJSInterop(t, cases)
	if len(jsOutputs) != len(cases) {
		t.Fatalf("interop random output mismatch: got=%d want=%d", len(jsOutputs), len(cases))
	}
	for i, tc := range cases {
		p := &Parser{DarkTheme: tc.Dark, ForceMonospace: tc.Mono}
		goOut := p.ConvertMicronToHTML(tc.Markup)
		jsOut := jsOutputs[i]
		goSig := signatureFromHTML(goOut)
		jsSig := signatureFromHTML(jsOut)
		if !sigsEqual(goSig, jsSig) {
			t.Fatalf("random interop mismatch on %s\nGo: %#v\nJS: %#v\nGo HTML: %s\nJS HTML: %s",
				tc.Name, goSig, jsSig, goOut, jsOut)
		}
	}
}

func runJSInterop(t *testing.T, cases []interopCase) []string {
	t.Helper()
	raw, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	harness := filepath.Join("testdata", "js_harness.js")
	parser := filepath.Join("testdata", "micron-parser.js")
	cmd := exec.Command("node", harness, parser)
	cmd.Stdin = bytes.NewReader(raw)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("js harness failed: %v stderr=%s", err, stderr.String())
	}
	var outputs []string
	if err := json.Unmarshal(out.Bytes(), &outputs); err != nil {
		t.Fatalf("decode js output: %v raw=%s", err, out.String())
	}
	return outputs
}

var tagStripper = regexp.MustCompile(`<[^>]*>`)
var wsCollapse = regexp.MustCompile(`\s+`)
var tagMatcher = regexp.MustCompile(`<\s*([a-zA-Z0-9]+)\b`)

func signatureFromHTML(in string) htmlSig {
	txt := tagStripper.ReplaceAllString(in, " ")
	txt = html.UnescapeString(txt)
	txt = wsCollapse.ReplaceAllString(txt, " ")
	txt = strings.TrimSpace(txt)
	tagCount := map[string]int{}
	for _, m := range tagMatcher.FindAllStringSubmatch(in, -1) {
		tagCount[strings.ToLower(m[1])]++
	}
	return htmlSig{
		TagCount:         tagCount,
		TextNormalized:   txt,
		Hrefs:            attrValues(in, "a", "href"),
		Destinations:     attrValues(in, "a", "data-destination"),
		Fields:           attrValues(in, "a", "data-fields"),
		InputTypes:       attrValues(in, "input", "type"),
		InputNames:       attrValues(in, "input", "name"),
		BoldCount:        strings.Count(in, "font-weight:bold"),
		UnderlineCount:   strings.Count(in, "text-decoration:underline"),
		ItalicCount:      strings.Count(in, "font-style:italic"),
		HeadingBlockUsed: strings.Contains(in, "display:inline-block;width:100%"),
	}
}

func attrValues(in, tagName, attr string) []string {
	pat := `<` + tagName + `\b[^>]*\b` + attr + `="([^"]*)"`
	re := regexp.MustCompile(pat)
	matches := re.FindAllStringSubmatch(in, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, html.UnescapeString(m[1]))
	}
	slices.Sort(out)
	return out
}

func sigsEqual(a, b htmlSig) bool {
	return maps.Equal(a.TagCount, b.TagCount) &&
		a.TextNormalized == b.TextNormalized &&
		slices.Equal(a.Hrefs, b.Hrefs) &&
		slices.Equal(a.Destinations, b.Destinations) &&
		slices.Equal(a.Fields, b.Fields) &&
		slices.Equal(a.InputTypes, b.InputTypes) &&
		slices.Equal(a.InputNames, b.InputNames) &&
		a.BoldCount == b.BoldCount &&
		a.UnderlineCount == b.UnderlineCount &&
		a.ItalicCount == b.ItalicCount &&
		a.HeadingBlockUsed == b.HeadingBlockUsed
}

func interopCorpus() []interopCase {
	markups := []struct {
		name   string
		markup string
	}{
		{name: "plain", markup: "hello world"},
		{name: "heading", markup: "> heading"},
		{name: "heading-divider", markup: "> heading\n-∿"},
		{name: "inline-format", markup: "`!b `*i `_u"},
		{name: "fg-bg-truecolor", markup: "`FT112233fg `BT445566bg"},
		{name: "field-text", markup: "`<24|name`alice>"},
		{name: "field-password", markup: "`<!|token`secret>"},
		{name: "checkbox-radio", markup: "`<?|agree`Yes`> `<^|pick`One`>"},
		{name: "link-basic", markup: "`[open`example.com]"},
		{name: "link-fields-vars", markup: "`[open`example.com`foo=1|bar=2|field1]"},
		{name: "literal", markup: "`=\n`!literal\n`="},
		{name: "header-colors", markup: "#!fg=ddd\n#!bg=111\nx"},
		{name: "depth-reset", markup: ">> nested\n< reset"},
		{name: "empty-lines", markup: "a\n\nb"},
		{name: "comment", markup: "shown\n# hidden"},
		{name: "ascii-cicada", markup: ".,~::::: ::: .,~::::: :::. :::::::-. :::.\n,;;;'''''' ;;; ,;;;'''''' ;;';; ;;, '';, ;;';;\n[[[ [[[ [[[ ,[[ '[[, '[[ [[ ,[[ '[[,\n$$$ $$$ $$$ c$$$cc$$$c $$, $$ c$$$cc$$$c\n'88bo,__,o, 888 '88bo,__,o, 888 888 888_,o8P' 888 888\n\"YUMMMMMP\" MMM \"YUMMMMMP\" YMM M\"M 'MMMMP\"' YMM M\"M"},
		{name: "figlet-backslash", markup: "| $$  \\ $$  /$$$$$$"},
	}
	out := make([]interopCase, 0, len(markups)*4)
	for _, m := range markups {
		for _, dark := range []bool{true, false} {
			for _, mono := range []bool{false, true} {
				out = append(out, interopCase{
					Name:   m.name + "-" + boolName(dark, "dark", "light") + "-" + boolName(mono, "mono", "plain"),
					Markup: m.markup,
					Dark:   dark,
					Mono:   mono,
				})
			}
		}
	}
	return out
}

func boolName(v bool, yes, no string) string {
	if v {
		return yes
	}
	return no
}

func randomMarkup(r *rand.Rand) string {
	tokens := []string{
		"hello", "world", "> section", ">> sub", "-", "`!bold", "`_u", "`*i",
		"`F123col", "`B456bg", "`f", "`b", "`c", "`l", "`r", "`a",
		"`<16|name`v>", "`<!|pwd`sec>", "`[open`example.com`k=v|f1]", "# comment", "",
	}
	lineCount := 1 + r.Intn(12)
	lines := make([]string, 0, lineCount)
	for range lineCount {
		parts := 1 + r.Intn(3)
		chunks := make([]string, 0, parts)
		for range parts {
			chunks = append(chunks, tokens[r.Intn(len(tokens))])
		}
		lines = append(lines, strings.Join(chunks, " "))
	}
	return strings.Join(lines, "\n")
}
