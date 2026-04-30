// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
)

func (p *Parser) parseField(line string, start int, s *State) (skip int, f *Field) {
	if start < 0 || start >= len(line) {
		return 0, nil
	}
	fieldStart := start + 1
	bt := strings.IndexByte(line[fieldStart:], '`')
	if bt < 0 {
		return 0, nil
	}
	bt += fieldStart
	fieldContent := line[fieldStart:bt]
	masked := false
	width := 24
	kind := FieldText
	name := fieldContent
	value := ""
	prechecked := false

	if strings.Contains(fieldContent, "|") {
		parts := strings.Split(fieldContent, "|")
		flags := parts[0]
		name = parts[1]
		if strings.Contains(flags, "^") {
			kind = FieldRadio
			flags = strings.ReplaceAll(flags, "^", "")
		} else if strings.Contains(flags, "?") {
			kind = FieldCheckbox
			flags = strings.ReplaceAll(flags, "?", "")
		} else if strings.Contains(flags, "!") {
			masked = true
			flags = strings.ReplaceAll(flags, "!", "")
		}
		if flags != "" {
			if w, err := strconv.Atoi(flags); err == nil {
				if w > 256 {
					w = 256
				}
				if w > 0 {
					width = w
				}
			}
		}
		if len(parts) > 2 {
			value = parts[2]
		}
		if len(parts) > 3 && parts[3] == "*" {
			prechecked = true
		}
	}

	end := strings.IndexByte(line[bt+1:], '>')
	if end < 0 {
		return 0, nil
	}
	end += bt + 1
	data := line[bt+1 : end]
	st := p.stateToStyle(s)
	sk := end - start + 1

	switch kind {
	case FieldCheckbox, FieldRadio:
		v := value
		if v == "" {
			v = data
		}
		return sk, &Field{
			Kind:       kind,
			Name:       name,
			Value:      v,
			Label:      data,
			Prechecked: prechecked,
			Style:      st,
		}
	default:
		return sk, &Field{
			Kind:   FieldText,
			Name:   name,
			Width:  width,
			Masked: masked,
			Value:  data,
			Style:  st,
		}
	}
}
