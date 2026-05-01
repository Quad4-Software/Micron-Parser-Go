// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

func (p *Parser) parseLink(line string, start int, s *State) (skip int, lk *Link) {
	if start < 0 || start >= len(line) {
		return 0, nil
	}
	end := strings.IndexByte(line[start+1:], ']')
	if end < 0 {
		return 0, nil
	}
	end += start + 1
	linkData := line[start+1 : end]
	var label, url, fields string
	i1 := strings.IndexByte(linkData, '`')
	if i1 < 0 {
		url = linkData
	} else {
		label = linkData[:i1]
		rest := linkData[i1+1:]
		i2 := strings.IndexByte(rest, '`')
		if i2 < 0 {
			url = rest
		} else {
			url = rest[:i2]
			fields = rest[i2+1:]
			if strings.IndexByte(fields, '`') >= 0 {
				return 0, nil
			}
		}
	}
	if url == "" {
		return 0, nil
	}
	if label == "" {
		label = url
	}
	url = FormatNomadnetworkURL(url)
	if p.ForceMonospace {
		label = p.splitAtSpaces(label)
	} else {
		label = htmlText(label)
	}
	var fieldList []string
	if fields != "" {
		fieldList = splitPipeList(fields)
	}
	return end - start + 1, &Link{
		URL:    url,
		Label:  label,
		Fields: fieldList,
		Style:  p.stateToStyle(s),
	}
}

func splitPipeList(s string) []string {
	if s == "" {
		return nil
	}
	out := make([]string, 0, strings.Count(s, "|")+1)
	start := 0
	for start <= len(s) {
		rel := strings.IndexByte(s[start:], '|')
		if rel < 0 {
			out = append(out, s[start:])
			return out
		}
		next := start + rel
		out = append(out, s[start:next])
		start = next + 1
	}
	return out
}
