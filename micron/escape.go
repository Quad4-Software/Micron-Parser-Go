// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "html"

func htmlText(s string) string {
	return html.EscapeString(s)
}
