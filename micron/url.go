// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"regexp"
	"strings"
)

var schemePrefix = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)

// FormatNomadnetworkURL ensures URLs have a scheme, matching micron-parser-js.
func FormatNomadnetworkURL(url string) string {
	if url == "" {
		return url
	}
	if schemePrefix.MatchString(url) {
		return url
	}
	return "nomadnetwork://" + url
}

func linkDirectURL(raw string) string {
	return strings.ReplaceAll(strings.ReplaceAll(raw, "nomadnetwork://", ""), "lxmf://", "")
}
