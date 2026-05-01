// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	htmlpkg "html"
	"strconv"
	"strings"
)

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
		var sa strings.Builder
		appendStyleAttr(&sa, st, s.DefaultBG)
		if sa.Len() > 0 {
			b.WriteString(`<span style="`)
			b.WriteString(sa.String())
			b.WriteString(`">`)
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
	var sa strings.Builder
	appendStyleAttr(&sa, f.Style, s.DefaultBG)
	hasStyle := sa.Len() > 0
	switch f.Kind {
	case FieldCheckbox:
		b.WriteString(`<label`)
		if hasStyle {
			b.WriteString(` style="`)
			b.WriteString(sa.String())
			b.WriteString(`"`)
		}
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
		if hasStyle {
			b.WriteString(` style="`)
			b.WriteString(sa.String())
			b.WriteString(`"`)
		}
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
		if hasStyle {
			b.WriteString(` style="`)
			b.WriteString(sa.String())
			b.WriteString(`"`)
		}
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
	var sa strings.Builder
	appendStyleAttr(&sa, lk.Style, s.DefaultBG)
	hasStyle := sa.Len() > 0
	direct := linkDirectURL(lk.URL)
	if len(lk.Fields) == 0 {
		b.WriteString(`<a class="Mu-nl" href="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" title="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" data-action="openNode" data-destination="`)
		b.WriteString(htmlAttr(direct))
		b.WriteString(`"`)
		if hasStyle {
			b.WriteString(` style="`)
			b.WriteString(sa.String())
			b.WriteString(`"`)
		}
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
	if hasStyle {
		b.WriteString(` style="`)
		b.WriteString(sa.String())
		b.WriteString(`"`)
	}
	b.WriteString(`>`)
	b.WriteString(lk.Label)
	b.WriteString(`</a>`)
}

func (p *Parser) writePartial(b *strings.Builder, pt *Partial, s *State) {
	var sa strings.Builder
	appendStyleAttr(&sa, pt.Style, s.DefaultBG)
	b.WriteString(`<div class="Mu-partial" data-partial-url="`)
	b.WriteString(htmlAttr(pt.URL))
	b.WriteString(`"`)
	if pt.RefreshSeconds > 0 {
		b.WriteString(` data-partial-refresh="`)
		b.WriteString(strconv.Itoa(pt.RefreshSeconds))
		b.WriteString(`"`)
	}
	if sa.Len() > 0 {
		b.WriteString(` style="`)
		b.WriteString(sa.String())
		b.WriteString(`"`)
	}
	b.WriteString(`></div>`)
}

func htmlAttr(s string) string {
	return htmlpkg.EscapeString(s)
}
