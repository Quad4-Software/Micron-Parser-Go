// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "maps"

import "strings"

// FieldInput is a normalized HTML input element snapshot.
type FieldInput struct {
	Type    string
	Name    string
	Value   string
	Checked bool
}

// RequestPayload is a resolved link/request execution payload.
type RequestPayload struct {
	Destination string            `json:"destination"`
	Fields      map[string]string `json:"fields"`
	RequestVars map[string]string `json:"request_vars"`
}

// CollectFormFields converts HTML input snapshots into Micron field semantics.
func CollectFormFields(inputs []FieldInput) map[string]string {
	out := map[string]string{}
	for _, in := range inputs {
		if in.Name == "" {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(in.Type))
		switch t {
		case "checkbox":
			if !in.Checked {
				continue
			}
			if prev, ok := out[in.Name]; ok && prev != "" {
				out[in.Name] = prev + "," + in.Value
			} else {
				out[in.Name] = in.Value
			}
		case "radio":
			if in.Checked {
				out[in.Name] = in.Value
			}
		default:
			out[in.Name] = in.Value
		}
	}
	return out
}

// BuildRequestPayload resolves fields and request vars from link metadata.
func BuildRequestPayload(allFields map[string]string, destination, fieldsSpec string) RequestPayload {
	dest, reqVars := splitDestinationVars(destination)
	selected := map[string]string{}
	if fieldsSpec == "*" {
		maps.Copy(selected, allFields)
	} else {
		for _, name := range splitFieldList(fieldsSpec) {
			if v, ok := allFields[name]; ok {
				selected[name] = v
			}
		}
	}
	return RequestPayload{
		Destination: dest,
		Fields:      selected,
		RequestVars: reqVars,
	}
}

func splitFieldList(fieldsSpec string) []string {
	if strings.TrimSpace(fieldsSpec) == "" {
		return nil
	}
	parts := strings.Split(fieldsSpec, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func splitDestinationVars(destination string) (string, map[string]string) {
	destination = strings.TrimSpace(destination)
	vars := map[string]string{}
	if destination == "" {
		return "", vars
	}
	before, after, ok := strings.Cut(destination, "`")
	if !ok {
		return destination, vars
	}
	base := before
	raw := after
	if raw == "" {
		return base, vars
	}
	for pair := range strings.SplitSeq(raw, "|") {
		if pair == "" {
			continue
		}
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(pair[:eq])
		v := strings.TrimSpace(pair[eq+1:])
		if k == "" {
			continue
		}
		vars[k] = v
	}
	return base, vars
}
