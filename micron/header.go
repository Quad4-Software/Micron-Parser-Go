// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// PageColors holds optional page-level colors from leading #! directives.
type PageColors struct {
	FG string
	BG string
}

// ParseHeaderTags reads leading #!fg= / #!bg= lines.
func ParseHeaderTags(markup string) PageColors {
	var out PageColors
	lines := strings.Split(markup, "\n")
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if !strings.HasPrefix(t, "#!") {
			break
		}
		if strings.HasPrefix(t, "#!fg=") {
			c := strings.TrimSpace(t[5:])
			if len(c) == 3 || len(c) == 6 {
				out.FG = c
			}
			continue
		}
		if strings.HasPrefix(t, "#!bg=") {
			c := strings.TrimSpace(t[5:])
			if len(c) == 3 || len(c) == 6 {
				out.BG = c
			}
		}
	}
	return out
}
