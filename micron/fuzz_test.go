// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func FuzzConvertMicronToHTML(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"> title",
		"`!bold `*italic `Fabc color",
		"`<24|name`value>",
		"`[link`example.com`x=y|field]",
		"`=\n`!literal\n`=",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		p := Parser{DarkTheme: true, ForceMonospace: true}
		out := p.ConvertMicronToHTML(in)
		if strings.Count(out, "<") != strings.Count(out, ">") {
			t.Fatalf("possibly malformed html: %q", out)
		}
	})
}

func FuzzParseHeaderTags(f *testing.F) {
	f.Add("#!fg=111\n#!bg=222\ntext")
	f.Add("text")
	f.Add("#!fg=bad\n#!bg=zzzz")
	f.Fuzz(func(t *testing.T, in string) {
		pc := ParseHeaderTags(in)
		if pc.FG != "" && len(pc.FG) != 3 && len(pc.FG) != 6 {
			t.Fatalf("invalid fg length: %q", pc.FG)
		}
		if pc.BG != "" && len(pc.BG) != 3 && len(pc.BG) != 6 {
			t.Fatalf("invalid bg length: %q", pc.BG)
		}
	})
}
