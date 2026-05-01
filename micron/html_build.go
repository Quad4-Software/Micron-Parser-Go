// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	htmlpkg "html"
	"strconv"
	"strings"
)

func appendQuotedHTMLStyleAttr(b *strings.Builder, st Style, defaultBG string) bool {
	var tmp strings.Builder
	tmp.Grow(96)
	appendStyleAttr(&tmp, st, defaultBG)
	if tmp.Len() == 0 {
		return false
	}
	b.WriteString(` style="`)
	b.WriteString(tmp.String())
	b.WriteByte('"')
	return true
}

func appendStyledSpanOpen(b *strings.Builder, st Style, defaultBG string) bool {
	var tmp strings.Builder
	tmp.Grow(96)
	appendStyleAttr(&tmp, st, defaultBG)
	if tmp.Len() == 0 {
		return false
	}
	b.WriteString(`<span style="`)
	b.WriteString(tmp.String())
	b.WriteString(`">`)
	return true
}

func (p *Parser) appendOutput(b *strings.Builder, parts []linePart, s *State) {
	var cur Style
	var have bool
	var span strings.Builder

	flush := func() {
		if !have {
			return
		}
		st := cur
		body := span.String()
		span.Reset()
		have = false
		if body == "" {
			return
		}
		if appendStyledSpanOpen(b, st, s.DefaultBG) {
			b.WriteString(body)
			b.WriteString(`</span>`)
		} else {
			b.WriteString(body)
		}
	}

	for _, pr := range parts {
		if pr.field != nil || pr.link != nil || pr.partial != nil {
			flush()
		}
		if pr.field != nil {
			p.writeField(b, pr.field, s)
			continue
		}
		if pr.link != nil {
			p.writeLink(b, pr.link, s)
			continue
		}
		if pr.partial != nil {
			p.writePartial(b, pr.partial, s)
			continue
		}
		st := pr.style
		if !have || !stylesEqual(&st, &cur) {
			flush()
			cur = st
			have = true
		}
		if pr.html != "" {
			span.WriteString(pr.html)
		} else if p.ForceMonospace {
			p.appendSplitAtSpaces(&span, pr.text)
		} else {
			appendHTMLText(&span, pr.text)
		}
	}
	flush()
}

func (p *Parser) writeField(b *strings.Builder, f *Field, s *State) {
	switch f.Kind {
	case FieldCheckbox:
		b.WriteString(`<label`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(`><input type="checkbox" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Prechecked {
			b.WriteString(` checked`)
		}
		b.WriteString(`/> `)
		appendHTMLText(b, f.Label)
		b.WriteString(`</label>`)
	case FieldRadio:
		b.WriteString(`<label`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(`><input type="radio" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Prechecked {
			b.WriteString(` checked`)
		}
		b.WriteString(`/> `)
		appendHTMLText(b, f.Label)
		b.WriteString(`</label>`)
	default:
		t := "text"
		if f.Masked {
			t = "password"
		}
		b.WriteString(`<input`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(` type="`)
		b.WriteString(t)
		b.WriteString(`" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Width > 0 {
			b.WriteString(` size="`)
			b.WriteString(strconv.Itoa(f.Width))
			b.WriteString(`"`)
		}
		b.WriteString(`/>`)
	}
}

func (p *Parser) writeLink(b *strings.Builder, lk *Link, s *State) {
	direct := linkDirectURL(lk.URL)
	if len(lk.Fields) == 0 {
		b.WriteString(`<a class="Mu-nl" href="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" title="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" data-action="openNode" data-destination="`)
		b.WriteString(htmlAttr(direct))
		b.WriteString(`"`)
		appendQuotedHTMLStyleAttr(b, lk.Style, s.DefaultBG)
		b.WriteString(`>`)
		b.WriteString(lk.Label)
		b.WriteString(`</a>`)
		return
	}
	var fieldStr strings.Builder
	var reqPairs strings.Builder
	foundAll := false
	for _, f := range lk.Fields {
		if f == "*" {
			foundAll = true
			continue
		}
		if strings.Contains(f, "=") {
			if reqPairs.Len() > 0 {
				reqPairs.WriteByte('|')
			}
			reqPairs.WriteString(f)
			continue
		}
		if fieldStr.Len() > 0 {
			fieldStr.WriteByte('|')
		}
		fieldStr.WriteString(f)
	}
	if foundAll {
		fieldStr.Reset()
		fieldStr.WriteByte('*')
	}
	if reqPairs.Len() > 0 {
		q := reqPairs.String()
		if strings.Contains(direct, "`") {
			direct = direct + "|" + q
		} else {
			direct = direct + "`" + q
		}
	}
	b.WriteString(`<a class="Mu-nl" href="`)
	b.WriteString(htmlAttr(lk.URL))
	b.WriteString(`" title="`)
	b.WriteString(htmlAttr(lk.URL))
	b.WriteString(`" data-action="openNode" data-destination="`)
	b.WriteString(htmlAttr(direct))
	b.WriteString(`" data-fields="`)
	b.WriteString(htmlAttr(fieldStr.String()))
	b.WriteString(`"`)
	appendQuotedHTMLStyleAttr(b, lk.Style, s.DefaultBG)
	b.WriteString(`>`)
	b.WriteString(lk.Label)
	b.WriteString(`</a>`)
}

func (p *Parser) writePartial(b *strings.Builder, pt *Partial, s *State) {
	b.WriteString(`<div class="Mu-partial" data-partial-url="`)
	b.WriteString(htmlAttr(pt.URL))
	b.WriteString(`"`)
	if pt.RefreshSeconds > 0 {
		b.WriteString(` data-partial-refresh="`)
		b.WriteString(strconv.Itoa(pt.RefreshSeconds))
		b.WriteString(`"`)
	}
	appendQuotedHTMLStyleAttr(b, pt.Style, s.DefaultBG)
	b.WriteString(`></div>`)
}

func htmlAttr(s string) string {
	return htmlpkg.EscapeString(s)
}
