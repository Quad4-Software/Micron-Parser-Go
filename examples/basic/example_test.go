// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package basic_test

import (
	"fmt"

	"git.quad4.io/Go-Libs/micron-parser-go/micron"
)

func ExampleFormatNomadnetworkURL() {
	fmt.Println(micron.FormatNomadnetworkURL("node/page.mu"))
	fmt.Println(micron.FormatNomadnetworkURL("https://example.com/a"))
	// Output:
	// nomadnetwork://node/page.mu
	// https://example.com/a
}

func ExampleParseHeaderTags() {
	c := micron.ParseHeaderTags("#!fg=abc\n#!bg=123456\n\nbody")
	fmt.Println(c.FG, c.BG)
	// Output:
	// abc 123456
}

func ExampleParser_ConvertMicronToHTML() {
	p := micron.Parser{DarkTheme: true, ForceMonospace: false}
	fmt.Print(p.ConvertMicronToHTML("hello"))
	// Output:
	// <div style="line-height:1.5;color:#ddd;"><div style="text-align:left;"><span style="color:#ddd;">hello</span></div></div>
}

func ExampleCollectFormFields() {
	out := micron.CollectFormFields([]micron.FieldInput{
		{Type: "text", Name: "user", Value: "alice"},
		{Type: "radio", Name: "plan", Value: "pro", Checked: true},
	})
	fmt.Println(out["user"], out["plan"])
	// Output:
	// alice pro
}

func ExampleBuildRequestPayload() {
	fields := map[string]string{"user": "alice", "x": "y"}
	p := micron.BuildRequestPayload(fields, "dest.mu`a=1", "user")
	fmt.Println(p.Destination, p.Fields["user"], p.RequestVars["a"])
	// Output:
	// dest.mu alice 1
}
