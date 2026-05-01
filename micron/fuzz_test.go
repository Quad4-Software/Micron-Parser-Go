// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func assertFuzzOutputHTMLSafety(t *testing.T, out string) {
	t.Helper()
	lower := strings.ToLower(out)
	forbidden := []string{
		"<script",
		"</script",
		"<iframe",
		"<frame",
		"<object",
		"<embed",
		"<link ",
		"<meta ",
		"<base ",
		"<style",
		"</style",
		"<template",
		"<svg",
		"<math",
		"<body",
		"<head",
		"<html",
		"<form",
	}
	for _, frag := range forbidden {
		if strings.Contains(lower, frag) {
			t.Fatalf("forbidden fragment %q in output", frag)
		}
	}
}

func fuzzConvertMicronAllModes(t *testing.T, in string) {
	t.Helper()
	parsers := []Parser{
		{DarkTheme: true, ForceMonospace: true},
		{DarkTheme: true, ForceMonospace: false},
		{DarkTheme: false, ForceMonospace: true},
		{DarkTheme: false, ForceMonospace: false},
	}
	for _, p := range parsers {
		out := p.ConvertMicronToHTML(in)
		if strings.Count(out, "<") != strings.Count(out, ">") {
			t.Fatalf("unbalanced angle brackets for %#v: %q", p, out)
		}
		assertFuzzOutputHTMLSafety(t, out)
	}
}

func FuzzConvertMicronToHTML(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"> title",
		"`!bold `*italic `Fabc color",
		"`<24|name`value>",
		"`[link`example.com`x=y|field]",
		"`=\n`!literal\n`=",
		"<script>alert(1)</script>",
		"<img src=x onerror=alert(1)> plain line no backticks",
		"| $$  \\ $$  /$$$$$$",
		"plain \\\\ two backslashes",
		"`[x`javascript:alert(1)`]",
		"`{x:data:text/html,<svg/onload=alert(1)>}",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		fuzzConvertMicronAllModes(t, in)
	})
}

func FuzzLightThemeConvertMicronToHTML(f *testing.F) {
	for _, s := range []string{
		"x",
		"`{a:/p.mu}",
		"<svg onload=evil>",
		"`\\`*still markup line",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		fuzzConvertMicronAllModes(t, in)
	})
}

func FuzzFormatNomadnetworkURL(f *testing.F) {
	f.Add("")
	f.Add("https://example.com/x")
	f.Add("node/page.mu")
	f.Fuzz(func(t *testing.T, url string) {
		_ = FormatNomadnetworkURL(url)
	})
}

func FuzzBuildRequestPayload(f *testing.F) {
	f.Add("dest`a=1|b=2", "user|x", "k=v|u=w")
	f.Fuzz(func(t *testing.T, destination, fieldsSpec, keysBlob string) {
		fields := map[string]string{}
		for kv := range strings.SplitSeq(keysBlob, "|") {
			k, v, ok := strings.Cut(kv, "=")
			if !ok || k == "" {
				continue
			}
			fields[k] = v
		}
		_ = BuildRequestPayload(fields, destination, fieldsSpec)
	})
}

func FuzzCollectFormFields(f *testing.F) {
	f.Add("text", "n", "v", false)
	f.Add("checkbox", "c", "1", true)
	f.Fuzz(func(t *testing.T, typ, name, val string, checked bool) {
		if len(name) > 512 || len(typ) > 64 || len(val) > 4096 {
			return
		}
		in := []FieldInput{{Type: typ, Name: name, Value: val, Checked: checked}}
		_ = CollectFormFields(in)
	})
}

func FuzzParseHeaderTags(f *testing.F) {
	f.Add("#!fg=111\n#!bg=222\ntext")
	f.Add("text")
	f.Add("#!fg=bad\n#!bg=zzzz")
	f.Fuzz(func(t *testing.T, in string) {
		pc := ParseHeaderTags(in)
		if pc.FG != "" && len(pc.FG) != 3 && len(pc.FG) != 6 {
			t.Fatalf("invalid fg length: %q", pc.FG)
		}
		if pc.BG != "" && len(pc.BG) != 3 && len(pc.BG) != 6 {
			t.Fatalf("invalid bg length: %q", pc.BG)
		}
	})
}
