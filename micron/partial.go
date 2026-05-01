// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
)

func (p *Parser) parsePartial(line string, start int, s *State) (skip int, pt *Partial) {
	if start < 0 || start >= len(line) {
		return 0, nil
	}
	end := strings.IndexByte(line[start+1:], '}')
	if end < 0 {
		return 0, nil
	}
	end += start + 1
	raw := strings.TrimSpace(line[start+1 : end])
	if raw == "" {
		return 0, nil
	}
	sep := strings.IndexByte(raw, '`')
	urlPart := raw
	refreshPart := ""
	if sep >= 0 {
		urlPart = raw[:sep]
		refreshPart = raw[sep+1:]
	}
	url := strings.TrimSpace(urlPart)
	if url == "" {
		return 0, nil
	}
	refresh := 0
	if refreshPart != "" {
		secs, err := strconv.Atoi(strings.TrimSpace(refreshPart))
		if err == nil && secs > 0 {
			refresh = secs
		}
	}
	return end - start + 1, &Partial{
		URL:            FormatNomadnetworkURL(url),
		RefreshSeconds: refresh,
		Style:          p.stateToStyle(s),
	}
}
