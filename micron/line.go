// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	lineNil = iota
	lineOmit
	lineHTML
)

func (p *Parser) parseLineInto(out *strings.Builder, line string, s *State) int {
	if len(line) > 0 {
		if line == "`=" {
			s.Literal = !s.Literal
			return lineNil
		}
		if !s.Literal {
			if line[0] == '#' {
				return lineOmit
			}
			if line[0] == '<' {
				s.Depth = 0
				return p.parseLineInto(out, line[1:], s)
			}
			if line[0] == '>' {
				i := 0
				for i < len(line) && line[i] == '>' {
					i++
				}
				s.Depth = i
				headingLine := line[i:]
				if len(headingLine) == 0 {
					return lineNil
				}
				style := headingStyle(p, i)
				latched := p.stateToStyle(s)
				p.styleToState(style, s)
				parts := p.makeOutput(s, headingLine)
				p.styleToState(latched, s)
				inner := p.joinLinePartsHTML(parts, s)
				if inner != "" {
					var hs strings.Builder
					hs.WriteString(`<div style="display:inline-block;width:100%;`)
					if tryAppendColorProperty(&hs, "color:", style.FG) {
						hs.WriteByte(';')
					}
					if tryAppendColorProperty(&hs, "background-color:", style.BG) {
						hs.WriteByte(';')
					}
					hs.WriteString(`"><div style="`)
					appendSectionIndentStyle(&hs, s)
					hs.WriteString(`">`)
					hs.WriteString(inner)
					hs.WriteString(`</div></div><br>`)
					out.WriteString(hs.String())
					return lineHTML
				}
				return lineNil
			}
			if line[0] == '-' {
				if len(line) == 1 {
					var b strings.Builder
					b.WriteString(`<hr style="all:revert;`)
					if tryAppendColorProperty(&b, "border-color:", s.FGColor) {
						b.WriteByte(';')
					}
					b.WriteString(`margin:0.5em 0.5em 0.5em 0.5em;`)
					if micronColorToken(s.BGColor) {
						b.WriteString(`box-shadow:0 0 0 0.5em `)
						writeMicronColorHex(&b, s.BGColor)
						b.WriteByte(';')
					}
					appendSectionIndentStyle(&b, s)
					b.WriteString(`"/>`)
					out.WriteString(b.String())
					return lineHTML
				}
				_, firstSize := utf8.DecodeRuneInString(line)
				r, _ := utf8.DecodeRuneInString(line[firstSize:])
				var b strings.Builder
				b.WriteString(`<div style="white-space:pre;white-space:nowrap;overflow:hidden;width:100%;`)
				if tryAppendColorProperty(&b, "color:", s.FGColor) {
					b.WriteByte(';')
				}
				if s.BGColor != s.DefaultBG && s.BGColor != "default" && tryAppendColorProperty(&b, "background-color:", s.BGColor) {
					b.WriteByte(';')
				}
				appendSectionIndentStyle(&b, s)
				b.WriteString(`">`)
				var tmp [utf8.UTFMax]byte
				n := utf8.EncodeRune(tmp[:], r)
				rText := string(tmp[:n])
				for range 250 {
					appendHTMLText(&b, rText)
				}
				b.WriteString(`</div>`)
				out.WriteString(b.String())
				return lineHTML
			}
		}
		if !s.Literal && strings.IndexByte(line, '`') < 0 {
			parts := p.makeOutput(s, line)
			inner := p.joinLinePartsHTML(parts, s)
			appendWrappedAlignedLineHTML(out, inner, s)
			return lineHTML
		}
		if !p.ForceMonospace && s.Literal {
			text := line
			if line == "\\`=" {
				text = "`="
			}
			appendWrappedAlignedLineHTML(out, p.fastPlainInner(text, s), s)
			return lineHTML
		}
		parts := p.makeOutput(s, line)
		inner := p.joinLinePartsHTML(parts, s)
		appendWrappedAlignedLineHTML(out, inner, s)
		return lineHTML
	}
	if s.BGColor != s.DefaultBG && s.BGColor != "default" {
		var b strings.Builder
		b.WriteString(`<div style="`)
		if tryAppendColorProperty(&b, "background-color:", s.BGColor) {
			b.WriteString(`;width:100%;display:block;height:1.2em;"><div style="`)
			appendSectionIndentStyleNoSemi(&b, s)
			b.WriteString(`">`)
			b.WriteString(`<br>`)
			b.WriteString(`</div></div>`)
			out.WriteString(b.String())
			return lineHTML
		}
	}
	out.WriteString(`<br>`)
	return lineHTML
}

func (p *Parser) fastPlainInner(line string, s *State) string {
	var body strings.Builder
	if p.ForceMonospace {
		p.appendSplitAtSpaces(&body, line)
	} else {
		appendHTMLText(&body, line)
	}
	sa := cachedStateStyleAttr(s)
	if sa == "" {
		return body.String()
	}
	var out strings.Builder
	out.WriteString(`<span style="`)
	out.WriteString(sa)
	out.WriteString(`">`)
	out.WriteString(body.String())
	out.WriteString(`</span>`)
	return out.String()
}

func cachedStateStyleAttr(s *State) string {
	key := stateStyleKey{
		FG:        s.FGColor,
		BG:        s.BGColor,
		Bold:      s.Formatting.Bold,
		Underline: s.Formatting.Underline,
		Italic:    s.Formatting.Italic,
	}
	if s.styleAttrMap != nil {
		if v, ok := s.styleAttrMap[key]; ok {
			return v
		}
	} else {
		s.styleAttrMap = make(map[stateStyleKey]string, 8)
	}
	v := styleAttr(Style{
		FG:        key.FG,
		BG:        key.BG,
		Bold:      key.Bold,
		Underline: key.Underline,
		Italic:    key.Italic,
	}, s.DefaultBG)
	s.styleAttrMap[key] = v
	return v
}

func appendWrappedAlignedLineHTML(out *strings.Builder, inner string, s *State) {
	var b strings.Builder
	b.WriteString(`<div style="text-align:`)
	b.WriteString(s.Align)
	b.WriteString(`;`)
	appendSectionIndentStyle(&b, s)
	b.WriteString(`">`)
	b.WriteString(inner)
	b.WriteString(`</div>`)
	wrapped := b.String()
	if s.BGColor != s.DefaultBG && s.BGColor != "default" {
		var bgWrap strings.Builder
		bgWrap.WriteString(`<div style="`)
		if tryAppendColorProperty(&bgWrap, "background-color:", s.BGColor) {
			bgWrap.WriteString(`;width:100%;display:block;">`)
			bgWrap.WriteString(wrapped)
			bgWrap.WriteString(`</div>`)
			out.WriteString(bgWrap.String())
			return
		}
	}
	out.WriteString(wrapped)
}

func sectionIndentStyleEm(s *State) float64 {
	ind := max((s.Depth-1)*2, 0)
	if ind <= 0 {
		return 0
	}
	return float64(ind) * 0.6
}

func appendSectionIndentStyle(b *strings.Builder, s *State) {
	em := sectionIndentStyleEm(s)
	if em <= 0 {
		return
	}
	b.WriteString("margin-left:")
	b.WriteString(strconv.FormatFloat(em, 'f', 1, 64))
	b.WriteString("em;")
}

func appendSectionIndentStyleNoSemi(b *strings.Builder, s *State) {
	em := sectionIndentStyleEm(s)
	if em <= 0 {
		return
	}
	b.WriteString("margin-left:")
	b.WriteString(strconv.FormatFloat(em, 'f', 1, 64))
	b.WriteString("em")
}
