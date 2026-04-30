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
	if strings.Contains(lower, "<script") {
		t.Fatalf("unexpected script-like tag in output")
	}
	if strings.Contains(lower, "javascript:") {
		t.Fatalf("unexpected javascript: URL in output")
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
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		p := Parser{DarkTheme: true, ForceMonospace: true}
		out := p.ConvertMicronToHTML(in)
		if strings.Count(out, "<") != strings.Count(out, ">") {
			t.Fatalf("possibly malformed html: %q", out)
		}
		assertFuzzOutputHTMLSafety(t, out)
	})
}

func FuzzLightThemeConvertMicronToHTML(f *testing.F) {
	f.Add("x")
	f.Add("`{a:/p.mu}")
	f.Fuzz(func(t *testing.T, in string) {
		p := Parser{DarkTheme: false, ForceMonospace: false}
		out := p.ConvertMicronToHTML(in)
		assertFuzzOutputHTMLSafety(t, out)
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
		for _, kv := range strings.Split(keysBlob, "|") {
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
