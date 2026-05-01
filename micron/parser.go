// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// ConvertMicronToHTML renders Micron markup to a self-contained HTML fragment.
// Text is escaped; only parser-emitted tags and attributes appear in the output.
// The caller supplies the full document; optional leading #!fg= / #!bg= lines
// affect default colors. The returned string is safe to treat as an HTML
// fragment only together with a sensible host CSP and link handling policy.
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
	if len(markup) > 0 {
		// HTML expansion varies by content; this reduces re-grows for common docs.
		b.Grow(len(markup) + len(markup)/2)
	}
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
		k := p.parseLineInto(&b, line, &s)
		switch k {
		case lineOmit:
			continue
		case lineNil:
			b.WriteString("<br>")
		}
	}
	var wrap strings.Builder
	wrap.Grow(64)
	if defaultFG != "" && defaultFG != "default" && tryAppendColorProperty(&wrap, "color:", defaultFG) {
		wrap.WriteByte(';')
	}
	if defaultBGVal != "" && defaultBGVal != "default" && tryAppendColorProperty(&wrap, "background-color:", defaultBGVal) {
		wrap.WriteByte(';')
	}
	if wrap.Len() > 0 {
		var out strings.Builder
		out.Grow(b.Len() + wrap.Len() + 24)
		out.WriteString(`<div style="`)
		out.WriteString(wrap.String())
		out.WriteString(`">`)
		out.WriteString(b.String())
		out.WriteString(`</div>`)
		return out.String()
	}
	return b.String()
}
