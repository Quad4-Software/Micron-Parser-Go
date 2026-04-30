// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"unicode/utf8"
)

// splitAfterSpaceSegments
func splitAfterSpaceSegments(s string) []string {
	if s == "" {
		return []string{""}
	}
	var parts []string
	start := 0
	for start < len(s) {
		rel := strings.IndexByte(s[start:], ' ')
		if rel < 0 {
			parts = append(parts, s[start:])
			break
		}
		sp := start + rel
		parts = append(parts, s[start:sp+1])
		start = sp + 1
	}
	return parts
}

func (p *Parser) splitAtSpaces(line string) string {
	var b strings.Builder
	for _, seg := range splitAfterSpaceSegments(line) {
		b.WriteString(`<span class="Mu-mws">`)
		b.WriteString(p.forceMonospace(seg))
		b.WriteString(`</span>`)
	}
	return b.String()
}

func (p *Parser) forceMonospace(line string) string {
	if !p.ForceMonospace {
		return htmlText(line)
	}
	var b strings.Builder
	for len(line) > 0 {
		r, sz := utf8.DecodeRuneInString(line)
		line = line[sz:]
		b.WriteString(`<span class="Mu-mnt">`)
		b.WriteString(htmlText(string(r)))
		b.WriteString(`</span>`)
	}
	return b.String()
}
