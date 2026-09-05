//go:build cgo

// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

// Package main builds libmicron as a C shared library (-buildmode=c-shared).
package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"unsafe"

	"micron-parser-go/micron"
)

func main() {}

func goString(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

func cString(s string) *C.char {
	return C.CString(s)
}

//export micron_convert
func micron_convert(markup *C.char, darkTheme, forceMonospace C.int) *C.char {
	p := micron.Parser{
		DarkTheme:      darkTheme != 0,
		ForceMonospace: forceMonospace != 0,
	}
	return cString(p.ConvertMicronToHTML(goString(markup)))
}

//export micron_parse_header_tags
func micron_parse_header_tags(markup *C.char) *C.char {
	colors := micron.ParseHeaderTags(goString(markup))
	raw, err := json.Marshal(struct {
		FG string `json:"fg"`
		BG string `json:"bg"`
	}{FG: colors.FG, BG: colors.BG})
	if err != nil {
		return cString(`{"fg":"","bg":""}`)
	}
	return cString(string(raw))
}

//export micron_collect_form_fields
func micron_collect_form_fields(inputsJSON *C.char) *C.char {
	var inputs []micron.FieldInput
	if raw := goString(inputsJSON); raw != "" {
		if err := json.Unmarshal([]byte(raw), &inputs); err != nil {
			return cString(`{}`)
		}
	}
	fields := micron.CollectFormFields(inputs)
	out, err := json.Marshal(fields)
	if err != nil {
		return cString(`{}`)
	}
	return cString(string(out))
}

//export micron_build_request_payload
func micron_build_request_payload(fieldsJSON, destination, fieldsSpec *C.char) *C.char {
	fields := map[string]string{}
	if raw := goString(fieldsJSON); raw != "" {
		_ = json.Unmarshal([]byte(raw), &fields)
	}
	payload := micron.BuildRequestPayload(fields, goString(destination), goString(fieldsSpec))
	out, err := json.Marshal(payload)
	if err != nil {
		return cString(`{"destination":"","fields":{},"request_vars":{}}`)
	}
	return cString(string(out))
}

//export micron_free
func micron_free(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

//export micron_lint
func micron_lint(markup *C.char) *C.char {
	p := micron.Parser{}
	raw, err := micron.DiagnosticsJSON(p.Lint(goString(markup)))
	if err != nil {
		return cString("[]")
	}
	return cString(string(raw))
}

//export micron_parse_json
func micron_parse_json(markup *C.char, darkTheme, forceMonospace C.int) *C.char {
	p := micron.Parser{
		DarkTheme:      darkTheme != 0,
		ForceMonospace: forceMonospace != 0,
	}
	raw, err := p.Parse(goString(markup)).DocumentJSON()
	if err != nil {
		return cString("null")
	}
	return cString(string(raw))
}
