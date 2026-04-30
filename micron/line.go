// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	lineNil = iota
	lineOmit
	lineHTML
)

// lineResult describes how a single input line maps to HTML.
type lineResult struct {
	Kind int
	HTML string
}

func (p *Parser) parseLine(line string, s *State) lineResult {
	if len(line) > 0 {
		if line == "`=" {
			s.Literal = !s.Literal
			return lineResult{Kind: lineNil}
		}
		if !s.Literal {
			if line[0] == '#' {
				return lineResult{Kind: lineOmit}
			}
			if line[0] == '<' {
				s.Depth = 0
				return p.parseLine(line[1:], s)
			}
			if line[0] == '>' {
				i := 0
				for i < len(line) && line[i] == '>' {
					i++
				}
				s.Depth = i
				headingLine := line[i:]
				if len(headingLine) == 0 {
					return lineResult{Kind: lineNil}
				}
				style := headingStyle(p, i)
				latched := p.stateToStyle(s)
				p.styleToState(style, s)
				parts := p.makeOutput(s, headingLine)
				p.styleToState(latched, s)
				inner := p.joinLinePartsHTML(parts, s)
				if inner != "" {
					fg := ColorToCSS(style.FG)
					bg := ColorToCSS(style.BG)
					var hs strings.Builder
					hs.WriteString(`<div style="display:inline-block;width:100%;`)
					if fg != "" {
						fmt.Fprintf(&hs, "color:%s;", fg)
					}
					if bg != "" {
						fmt.Fprintf(&hs, "background-color:%s;", bg)
					}
					hs.WriteString(`"><div style="`)
					hs.WriteString(sectionIndentStyle(s))
					hs.WriteString(`">`)
					hs.WriteString(inner)
					hs.WriteString(`</div></div><br>`)
					return lineResult{Kind: lineHTML, HTML: hs.String()}
				}
				return lineResult{Kind: lineNil}
			}
			if line[0] == '-' {
				if len(line) == 1 {
					fg := ColorToCSS(s.FGColor)
					bg := ColorToCSS(s.BGColor)
					var b strings.Builder
					b.WriteString(`<hr style="all:revert;`)
					if fg != "" {
						fmt.Fprintf(&b, "border-color:%s;", fg)
					}
					b.WriteString(`margin:0.5em 0.5em 0.5em 0.5em;`)
					if bg != "" {
						fmt.Fprintf(&b, "box-shadow:0 0 0 0.5em %s;", bg)
					}
					b.WriteString(sectionIndentStyle(s))
					b.WriteString(`"/>`)
					return lineResult{Kind: lineHTML, HTML: b.String()}
				}
				_, firstSize := utf8.DecodeRuneInString(line)
				r, _ := utf8.DecodeRuneInString(line[firstSize:])
				repeated := strings.Repeat(string(r), 250)
				fg := ColorToCSS(s.FGColor)
				var b strings.Builder
				b.WriteString(`<div style="white-space:pre;white-space:nowrap;overflow:hidden;width:100%;`)
				if fg != "" {
					fmt.Fprintf(&b, "color:%s;", fg)
				}
				if s.BGColor != s.DefaultBG && s.BGColor != "default" {
					if bg := ColorToCSS(s.BGColor); bg != "" {
						fmt.Fprintf(&b, "background-color:%s;", bg)
					}
				}
				b.WriteString(sectionIndentStyle(s))
				b.WriteString(`">`)
				b.WriteString(htmlText(repeated))
				b.WriteString(`</div>`)
				return lineResult{Kind: lineHTML, HTML: b.String()}
			}
		}
		parts := p.makeOutput(s, line)
		inner := p.joinLinePartsHTML(parts, s)
		var b strings.Builder
		b.WriteString(`<div style="text-align:`)
		b.WriteString(s.Align)
		b.WriteString(`;`)
		b.WriteString(sectionIndentStyle(s))
		b.WriteString(`">`)
		b.WriteString(inner)
		b.WriteString(`</div>`)
		wrapped := b.String()
		if s.BGColor != s.DefaultBG && s.BGColor != "default" {
			bg := ColorToCSS(s.BGColor)
			if bg != "" {
				return lineResult{Kind: lineHTML, HTML: fmt.Sprintf(
					`<div style="background-color:%s;width:100%%;display:block;">%s</div>`,
					bg, wrapped)}
			}
		}
		return lineResult{Kind: lineHTML, HTML: wrapped}
	}
	br := `<br>`
	if s.BGColor != s.DefaultBG && s.BGColor != "default" {
		bg := ColorToCSS(s.BGColor)
		if bg != "" {
			return lineResult{Kind: lineHTML, HTML: fmt.Sprintf(
				`<div style="background-color:%s;width:100%%;display:block;height:1.2em;"><div style="%s">%s</div></div>`,
				bg, strings.TrimSuffix(sectionIndentStyle(s), ";"), br)}
		}
	}
	return lineResult{Kind: lineHTML, HTML: br}
}

func sectionIndentStyle(s *State) string {
	ind := (s.Depth - 1) * 2
	if ind < 0 {
		ind = 0
	}
	if ind <= 0 {
		return ""
	}
	return fmt.Sprintf("margin-left:%.1fem;", float64(ind)*0.6)
}
