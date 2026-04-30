// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "testing"

func TestCollectFormFields(t *testing.T) {
	inputs := []FieldInput{
		{Type: "text", Name: "user", Value: "alice"},
		{Type: "password", Name: "pass", Value: "hidden"},
		{Type: "checkbox", Name: "role", Value: "admin", Checked: true},
		{Type: "checkbox", Name: "role", Value: "writer", Checked: true},
		{Type: "checkbox", Name: "role", Value: "reader", Checked: false},
		{Type: "radio", Name: "theme", Value: "dark", Checked: false},
		{Type: "radio", Name: "theme", Value: "light", Checked: true},
	}
	got := CollectFormFields(inputs)
	if got["user"] != "alice" || got["pass"] != "hidden" {
		t.Fatalf("unexpected text/password values: %#v", got)
	}
	if got["role"] != "admin,writer" {
		t.Fatalf("unexpected checkbox concat: %#v", got["role"])
	}
	if got["theme"] != "light" {
		t.Fatalf("unexpected radio value: %#v", got["theme"])
	}
}

func TestBuildRequestPayload(t *testing.T) {
	fields := map[string]string{
		"user":  "alice",
		"token": "x1",
	}
	got := BuildRequestPayload(fields, "dest/path`a=1|b=2", "user|missing")
	if got.Destination != "dest/path" {
		t.Fatalf("dest: %q", got.Destination)
	}
	if got.RequestVars["a"] != "1" || got.RequestVars["b"] != "2" {
		t.Fatalf("vars: %#v", got.RequestVars)
	}
	if len(got.Fields) != 1 || got.Fields["user"] != "alice" {
		t.Fatalf("fields: %#v", got.Fields)
	}
}

func TestBuildRequestPayloadAllFields(t *testing.T) {
	fields := map[string]string{"a": "1", "b": "2"}
	got := BuildRequestPayload(fields, "dest", "*")
	if got.Destination != "dest" {
		t.Fatalf("dest: %q", got.Destination)
	}
	if len(got.Fields) != 2 {
		t.Fatalf("fields: %#v", got.Fields)
	}
}
