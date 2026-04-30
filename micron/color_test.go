// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "testing"

func TestColorToCSS(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"default", ""},
		{"abc", "#abc"},
		{"ABCDEF", "#ABCDEF"},
		{"g50", "#7f7f7f"},
		{"g00", "#000000"},
		{"g99", "#fcfcfc"},
		{"ggg", "#7f7f7f"},
	}
	for _, tc := range tests {
		if got := ColorToCSS(tc.in); got != tc.want {
			t.Errorf("ColorToCSS(%q) = %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatNomadnetworkURL(t *testing.T) {
	if got := FormatNomadnetworkURL("http://x"); got != "http://x" {
		t.Fatalf("got %q", got)
	}
	if got := FormatNomadnetworkURL("example"); got != "nomadnetwork://example" {
		t.Fatalf("got %q", got)
	}
}

func TestParseHeaderTags(t *testing.T) {
	in := "#!fg=abc\n#!bg=123456\n\nx"
	pc := ParseHeaderTags(in)
	if pc.FG != "abc" || pc.BG != "123456" {
		t.Fatalf("%+v", pc)
	}
}

func TestSplitAfterSpaceSegments(t *testing.T) {
	got := splitAfterSpaceSegments("foo bar  baz")
	want := []string{"foo ", "bar ", " ", "baz"}
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("i=%d got %q want %q", i, got[i], want[i])
		}
	}
}
