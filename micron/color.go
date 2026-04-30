// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"math"
	"strconv"
	"strings"
)

const (
	defaultBG      = "default"
	defaultFGDark  = "ddd"
	defaultFGLight = "222"
)

// ColorToCSS maps Micron color tokens (hex, grayscale gNN, defaults) to CSS
// color strings, or returns an empty string when c is empty, "default", or not
// recognized.
func ColorToCSS(c string) string {
	if c == "" || c == defaultBG {
		return ""
	}
	if len(c) == 3 && isHex3(c) {
		return "#" + c
	}
	if len(c) == 6 && isHex6(c) {
		return "#" + c
	}
	if len(c) == 3 && c[0] == 'g' {
		v, err := strconv.Atoi(c[1:])
		if err != nil || v < 0 {
			v = 50
		}
		if v > 99 {
			v = 99
		}
		h := byte(math.Floor(float64(v) * 2.55))
		return rgbHex(h, h, h)
	}
	return ""
}

func isHex3(s string) bool {
	for i := range 3 {
		if !isHexByte(s[i]) {
			return false
		}
	}
	return true
}

func isHex6(s string) bool {
	for i := range 6 {
		if !isHexByte(s[i]) {
			return false
		}
	}
	return true
}

func isHexByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func rgbHex(r, g, b byte) string {
	const hx = "0123456789abcdef"
	return "#" +
		string([]byte{hx[r>>4], hx[r&0xf]}) +
		string([]byte{hx[g>>4], hx[g&0xf]}) +
		string([]byte{hx[b>>4], hx[b&0xf]})
}

func headingStyle(p *Parser, level int) Style {
	if p.DarkTheme {
		switch level {
		case 1:
			return Style{FG: "222", BG: "bbb", Bold: false, Underline: false, Italic: false}
		case 2:
			return Style{FG: "111", BG: "999", Bold: false, Underline: false, Italic: false}
		case 3:
			return Style{FG: "000", BG: "777", Bold: false, Underline: false, Italic: false}
		}
		return plainStyle(p)
	}
	switch level {
	case 1:
		return Style{FG: "000", BG: "777", Bold: false, Underline: false, Italic: false}
	case 2:
		return Style{FG: "111", BG: "aaa", Bold: false, Underline: false, Italic: false}
	case 3:
		return Style{FG: "222", BG: "ccc", Bold: false, Underline: false, Italic: false}
	}
	return plainStyle(p)
}

func plainStyle(p *Parser) Style {
	fg := defaultFGLight
	if p.DarkTheme {
		fg = defaultFGDark
	}
	return Style{FG: fg, BG: defaultBG, Bold: false, Underline: false, Italic: false}
}

func (p *Parser) stateToStyle(s *State) Style {
	return Style{
		FG:        s.FGColor,
		BG:        s.BGColor,
		Bold:      s.Formatting.Bold,
		Underline: s.Formatting.Underline,
		Italic:    s.Formatting.Italic,
	}
}

func (p *Parser) styleToState(st Style, s *State) {
	s.FGColor = st.FG
	s.BGColor = st.BG
	s.Formatting.Bold = st.Bold
	s.Formatting.Underline = st.Underline
	s.Formatting.Italic = st.Italic
}

func stylesEqual(a, b *Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.FG == b.FG && a.BG == b.BG && a.Bold == b.Bold &&
		a.Underline == b.Underline && a.Italic == b.Italic
}

func styleAttr(st Style, defaultBG string) string {
	var b strings.Builder
	fg := ColorToCSS(st.FG)
	if fg != "" && fg != "default" {
		b.WriteString("color:")
		b.WriteString(fg)
		b.WriteByte(';')
	}
	bg := ColorToCSS(st.BG)
	if bg != "" && st.BG != defaultBG && st.BG != "default" {
		b.WriteString("background-color:")
		b.WriteString(bg)
		b.WriteString(";display:inline-block;")
	}
	if st.Bold {
		b.WriteString("font-weight:bold;")
	}
	if st.Underline {
		if b.Len() > 0 {
			b.WriteString("text-decoration:underline;")
		} else {
			b.WriteString("text-decoration:underline;")
		}
	}
	if st.Italic {
		b.WriteString("font-style:italic;")
	}
	return b.String()
}
