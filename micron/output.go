// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

func (p *Parser) makeOutput(s *State, line string) []linePart {
	if s.Literal {
		if line == "\\`=" {
			line = "`="
		}
		st := p.stateToStyle(s)
		if p.ForceMonospace {
			return []linePart{{style: st, html: p.splitAtSpaces(line)}}
		}
		return []linePart{{style: st, text: line}}
	}

	var out []linePart
	part := ""
	modeText := true
	escape := false
	skip := 0
	i := 0

	flushPart := func() {
		if part == "" {
			return
		}
		st := p.stateToStyle(s)
		if p.ForceMonospace {
			out = append(out, linePart{style: st, html: p.splitAtSpaces(part)})
		} else {
			out = append(out, linePart{style: st, text: part})
		}
		part = ""
	}

	for i < len(line) {
		if skip > 0 {
			skip--
			i++
			continue
		}
		if !modeText {
			c := line[i]
			switch c {
			case '_':
				s.Formatting.Underline = !s.Formatting.Underline
			case '!':
				s.Formatting.Bold = !s.Formatting.Bold
			case '*':
				s.Formatting.Italic = !s.Formatting.Italic
			case 'F':
				if i+1 < len(line) && line[i+1] == 'T' && len(line) >= i+8 {
					s.FGColor = line[i+2 : i+8]
					skip = 7
				} else if len(line) >= i+4 {
					s.FGColor = line[i+1 : i+4]
					skip = 3
				}
			case 'f':
				s.FGColor = s.DefaultFG
			case 'B':
				if i+1 < len(line) && line[i+1] == 'T' && len(line) >= i+8 {
					s.BGColor = line[i+2 : i+8]
					skip = 7
					flushPart()
				} else if len(line) >= i+4 {
					s.BGColor = line[i+1 : i+4]
					skip = 3
					flushPart()
				}
			case 'b':
				s.BGColor = s.DefaultBG
				flushPart()
			case '`':
				s.Formatting.Bold = false
				s.Formatting.Underline = false
				s.Formatting.Italic = false
				s.FGColor = s.DefaultFG
				s.BGColor = s.DefaultBG
				s.Align = s.DefaultAlign
				modeText = true
			case 'c':
				s.Align = "center"
			case 'l':
				s.Align = "left"
			case 'r':
				s.Align = "right"
			case 'a':
				s.Align = s.DefaultAlign
			case '<':
				flushPart()
				if sk, f := p.parseField(line, i, s); f != nil {
					out = append(out, linePart{field: f})
					i += sk
					modeText = true
					continue
				}
			case '[':
				flushPart()
				if sk, lk := p.parseLink(line, i, s); lk != nil {
					out = append(out, linePart{link: lk})
					i += sk
					modeText = true
					continue
				}
			}
			modeText = true
			i++
			continue
		}

		c := line[i]
		if escape {
			part += string(c)
			escape = false
			i++
			continue
		}
		if c == '\\' {
			escape = true
			i++
			continue
		}
		if c == '`' {
			if i+1 < len(line) && line[i+1] == '`' {
				flushPart()
				s.Formatting.Bold = false
				s.Formatting.Underline = false
				s.Formatting.Italic = false
				s.FGColor = s.DefaultFG
				s.BGColor = s.DefaultBG
				s.Align = s.DefaultAlign
				i += 2
				continue
			}
			flushPart()
			modeText = false
			i++
			continue
		}
		part += string(c)
		i++
	}
	flushPart()
	return out
}

func (p *Parser) joinLinePartsHTML(parts []linePart, s *State) string {
	var b strings.Builder
	p.appendOutput(&b, parts, s)
	return b.String()
}
