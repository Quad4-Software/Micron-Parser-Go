// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"math/rand"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf8"
)

func randomPlainMicronLine(rng *rand.Rand, maxLen int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !@#$%^&*()-_=+[]{}|;:\",.<>/?~"
	var b strings.Builder
	b.WriteByte('a')
	n := 1 + rng.Intn(maxLen)
	for range n {
		b.WriteByte(alphabet[rng.Intn(len(alphabet))])
	}
	return b.String()
}

func randomInteropPlainLine(rng *rand.Rand, maxLen int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 *()-_=+[]{}|;:,./?~@#$%^!"
	var b strings.Builder
	b.WriteByte('a')
	n := 1 + rng.Intn(maxLen)
	for range n {
		b.WriteByte(alphabet[rng.Intn(len(alphabet))])
	}
	return b.String()
}

func countMuMnt(html string) int {
	return strings.Count(html, `class="Mu-mnt"`)
}

func TestPlainLinePreservesBackslashesFiglet(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	line := `| $$  \ $$  /$$$$$$`
	out := p.ConvertMicronToHTML(line)
	if strings.Count(out, "\\") != 1 {
		t.Fatalf("want one literal backslash in HTML, got %q", out)
	}
	row := `\_______/`
	out2 := p.ConvertMicronToHTML(row)
	if countMuMnt(out2) != utf8.RuneCountInString(row) {
		t.Fatalf("Mu-mnt count mismatch for %q", row)
	}
}

func TestMarkupLinePreservesFigletBackslashSlash(t *testing.T) {
	line := "|  |__   `---|  |----`|  | |  \\  /  | x"
	p := Parser{DarkTheme: true, ForceMonospace: true}
	out := p.ConvertMicronToHTML(line)
	if strings.Count(out, "\\") < 1 {
		t.Fatalf("want preserved ASCII \\ near /, got %q", out)
	}
}

func TestMonospaceMuMntCountDynamicPlainASCII(t *testing.T) {
	rng := rand.New(rand.NewSource(80201))
	for range 300 {
		line := randomPlainMicronLine(rng, 96)
		p := Parser{
			DarkTheme:      rng.Intn(2) == 0,
			ForceMonospace: true,
		}
		out := p.ConvertMicronToHTML(line)
		want := utf8.RuneCountInString(line)
		got := countMuMnt(out)
		if got != want {
			t.Fatalf("DarkTheme=%v line=%q runes=%d Mu-mnt=%d\n%s", p.DarkTheme, line, want, got, out)
		}
	}
}

func TestInteropASCIIPlainDynamic(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	rng := rand.New(rand.NewSource(91002))
	cases := make([]interopCase, 0, 80)
	for range 80 {
		line := randomInteropPlainLine(rng, 72)
		dark := rng.Intn(2) == 0
		mono := rng.Intn(2) == 0
		cases = append(cases, interopCase{
			Name:   "dyn-ascii-" + line[:min(8, len(line))],
			Markup: line,
			Dark:   dark,
			Mono:   mono,
		})
	}
	jsOutputs := runJSInterop(t, cases)
	for i, tc := range cases {
		if tc.Mono {
			continue
		}
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

func TestMonospaceMuMntMultilineJoinConsistent(t *testing.T) {
	rng := rand.New(rand.NewSource(44033))
	p := Parser{DarkTheme: true, ForceMonospace: true}
	lines := make([]string, 16)
	sumRunes := 0
	sumSoloMu := 0
	for i := range lines {
		lines[i] = randomPlainMicronLine(rng, 64)
		sumRunes += utf8.RuneCountInString(lines[i])
		sumSoloMu += countMuMnt(p.ConvertMicronToHTML(lines[i]))
	}
	doc := strings.Join(lines, "\n")
	combined := p.ConvertMicronToHTML(doc)
	got := countMuMnt(combined)
	if sumRunes != sumSoloMu {
		t.Fatalf("single-line Mu-mnt sum %d != rune sum %d", sumSoloMu, sumRunes)
	}
	if got != sumSoloMu {
		t.Fatalf("combined Mu-mnt %d != per-line sum %d\n%s", got, sumSoloMu, combined)
	}
}
