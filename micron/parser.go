// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// ConvertMicronToHTML renders Micron markup to a self-contained HTML fragment.
// Text is escaped; only parser-emitted tags and attributes appear in the output.
func (p *Parser) ConvertMicronToHTML(markup string) string {
	pc := ParseHeaderTags(markup)
	plain := plainStyle(p)
	defaultFG := pc.FG
	if defaultFG == "" {
		defaultFG = plain.FG
	}
	defaultBGVal := plain.BG
	if pc.BG != "" {
		defaultBGVal = pc.BG
	}
	s := State{
		Literal:      false,
		Depth:        0,
		FGColor:      defaultFG,
		BGColor:      defaultBGVal,
		DefaultAlign: "left",
		Align:        "left",
		DefaultFG:    defaultFG,
		DefaultBG:    defaultBGVal,
	}
	var b strings.Builder
	for start := 0; start <= len(markup); {
		nextRel := strings.IndexByte(markup[start:], '\n')
		line := ""
		if nextRel < 0 {
			line = markup[start:]
			start = len(markup) + 1
		} else {
			next := start + nextRel
			line = markup[start:next]
			start = next + 1
		}
		r := p.parseLine(line, &s)
		switch r.Kind {
		case lineOmit:
			continue
		case lineNil:
			b.WriteString("<br>")
		case lineHTML:
			b.WriteString(r.HTML)
		}
	}
	out := b.String()
	wrap := ""
	if defaultFG != "" && defaultFG != "default" {
		if fg := ColorToCSS(defaultFG); fg != "" {
			wrap += "color:" + fg + ";"
		}
	}
	if defaultBGVal != "" && defaultBGVal != "default" {
		if bg := ColorToCSS(defaultBGVal); bg != "" {
			wrap += "background-color:" + bg + ";"
		}
	}
	if wrap != "" {
		return `<div style="` + wrap + `">` + out + `</div>`
	}
	return out
}
