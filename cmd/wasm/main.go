//go:build js && wasm

// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"git.quad4.io/Go-Libs/micron-parser-go/micron"
)

func main() {
	convert := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return js.Undefined()
		}
		markup := args[0].String()
		dark := true
		if len(args) > 1 {
			dark = args[1].Bool()
		}
		mono := true
		if len(args) > 2 {
			mono = args[2].Bool()
		}
		p := micron.Parser{DarkTheme: dark, ForceMonospace: mono}
		return p.ConvertMicronToHTML(markup)
	})
	collectFields := js.FuncOf(func(this js.Value, args []js.Value) any {
		rootSel := "#preview"
		if len(args) > 0 {
			rootSel = strings.TrimSpace(args[0].String())
		}
		inputs := collectInputs(rootSel)
		fields := micron.CollectFormFields(inputs)
		raw, _ := json.Marshal(fields)
		return string(raw)
	})
	resolveLink := js.FuncOf(func(this js.Value, args []js.Value) any {
		rootSel := "#preview"
		if len(args) > 0 {
			rootSel = strings.TrimSpace(args[0].String())
		}
		destination := ""
		if len(args) > 1 {
			destination = args[1].String()
		}
		fieldsSpec := ""
		if len(args) > 2 {
			fieldsSpec = args[2].String()
		}
		fields := micron.CollectFormFields(collectInputs(rootSel))
		payload := micron.BuildRequestPayload(fields, destination, fieldsSpec)
		raw, _ := json.Marshal(payload)
		return string(raw)
	})
	js.Global().Set("micronConvert", convert)
	js.Global().Set("micronCollectFields", collectFields)
	js.Global().Set("micronResolveLink", resolveLink)
	select {}
}

func collectInputs(rootSelector string) []micron.FieldInput {
	doc := js.Global().Get("document")
	root := doc
	if rootSelector != "" {
		if el := doc.Call("querySelector", rootSelector); !el.IsNull() && !el.IsUndefined() {
			root = el
		}
	}
	nodes := root.Call("querySelectorAll", "input[name],textarea[name]")
	n := nodes.Get("length").Int()
	out := make([]micron.FieldInput, 0, n)
	for i := 0; i < n; i++ {
		el := nodes.Index(i)
		typ := strings.ToLower(el.Get("type").String())
		if strings.EqualFold(el.Get("tagName").String(), "TEXTAREA") {
			typ = "text"
		}
		out = append(out, micron.FieldInput{
			Type:    typ,
			Name:    el.Get("name").String(),
			Value:   el.Get("value").String(),
			Checked: el.Get("checked").Bool(),
		})
	}
	return out
}
