// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"fmt"
	htmlpkg "html"
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
		sa := styleAttr(st, s.DefaultBG)
		if sa != "" {
			b.WriteString(`<span style="`)
			b.WriteString(sa)
			b.WriteString(`">`)
			b.WriteString(body)
			b.WriteString(`</span>`)
		} else {
			b.WriteString(body)
		}
	}

	for _, pr := range parts {
		if pr.field != nil || pr.link != nil {
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
		st := pr.style
		if !have || !stylesEqual(&st, &cur) {
			flush()
			cur = st
			have = true
		}
		if pr.html != "" {
			span.WriteString(pr.html)
		} else {
			span.WriteString(htmlText(pr.text))
		}
	}
	flush()
}

func (p *Parser) writeField(b *strings.Builder, f *Field, s *State) {
	sa := styleAttr(f.Style, s.DefaultBG)
	styleOpen := ""
	if sa != "" {
		styleOpen = ` style="` + sa + `"`
	}
	switch f.Kind {
	case FieldCheckbox:
		chk := ""
		if f.Prechecked {
			chk = ` checked`
		}
		fmt.Fprintf(b, `<label%s><input type="checkbox" name="%s" value="%s"%s/> %s</label>`,
			styleOpen, htmlAttr(f.Name), htmlAttr(f.Value), chk, htmlText(f.Label))
	case FieldRadio:
		chk := ""
		if f.Prechecked {
			chk = ` checked`
		}
		fmt.Fprintf(b, `<label%s><input type="radio" name="%s" value="%s"%s/> %s</label>`,
			styleOpen, htmlAttr(f.Name), htmlAttr(f.Value), chk, htmlText(f.Label))
	default:
		t := "text"
		if f.Masked {
			t = "password"
		}
		sz := ""
		if f.Width > 0 {
			sz = fmt.Sprintf(` size="%d"`, f.Width)
		}
		fmt.Fprintf(b, `<input%s type="%s" name="%s" value="%s"%s/>`,
			styleOpen, t, htmlAttr(f.Name), htmlAttr(f.Value), sz)
	}
}

func (p *Parser) writeLink(b *strings.Builder, lk *Link, s *State) {
	sa := styleAttr(lk.Style, s.DefaultBG)
	styleOpen := ""
	if sa != "" {
		styleOpen = ` style="` + sa + `"`
	}
	direct := linkDirectURL(lk.URL)
	if len(lk.Fields) == 0 {
		b.WriteString(`<a class="Mu-nl" href="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" title="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" data-action="openNode" data-destination="`)
		b.WriteString(htmlAttr(direct))
		b.WriteString(`"`)
		b.WriteString(styleOpen)
		b.WriteString(`>`)
		b.WriteString(lk.Label)
		b.WriteString(`</a>`)
		return
	}
	var submit []string
	var reqPairs []string
	foundAll := false
	for _, f := range lk.Fields {
		if f == "*" {
			foundAll = true
			continue
		}
		if strings.Contains(f, "=") {
			reqPairs = append(reqPairs, f)
			continue
		}
		submit = append(submit, f)
	}
	fieldStr := strings.Join(submit, "|")
	if foundAll {
		fieldStr = "*"
	}
	if len(reqPairs) > 0 {
		q := strings.Join(reqPairs, "|")
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
	b.WriteString(htmlAttr(fieldStr))
	b.WriteString(`"`)
	b.WriteString(styleOpen)
	b.WriteString(`>`)
	b.WriteString(lk.Label)
	b.WriteString(`</a>`)
}

func htmlAttr(s string) string {
	return htmlpkg.EscapeString(s)
}
