// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func TestSmokeRepresentativeDocument(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	in := strings.Join([]string{
		"#!fg=ddd",
		"#!bg=111",
		"> Welcome",
		"-∿",
		"`!bold text",
		"`_underlined",
		"`<24|name`alice>",
		"`[open`node.example`*|q=1]",
		"",
	}, "\n")
	out := p.ConvertMicronToHTML(in)
	mustContain := []string{
		"background-color:#111",
		"Welcome",
		"white-space:nowrap",
		`type="text"`,
		`class="Mu-nl"`,
		`data-destination=`,
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output: %s", want, out)
		}
	}
}
