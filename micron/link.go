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
	parts := strings.Split(linkData, "`")
	var label, url, fields string
	switch len(parts) {
	case 1:
		url = linkData
	case 2:
		label = parts[0]
		url = parts[1]
	case 3:
		label = parts[0]
		url = parts[1]
		fields = parts[2]
	default:
		return 0, nil
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
		fieldList = strings.Split(fields, "|")
	}
	return end - start + 1, &Link{
		URL:    url,
		Label:  label,
		Fields: fieldList,
		Style:  p.stateToStyle(s),
	}
}
