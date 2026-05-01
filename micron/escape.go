// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"html"
	"strings"
)

func htmlText(s string) string {
	return html.EscapeString(s)
}

func appendHTMLText(b *strings.Builder, s string) {
	start := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;"
		case '\'':
			esc = "&#39;"
		default:
			continue
		}
		if start < i {
			b.WriteString(s[start:i])
		}
		b.WriteString(esc)
		start = i + 1
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
}
