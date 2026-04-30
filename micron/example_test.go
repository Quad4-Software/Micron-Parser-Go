// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron_test

import (
	"fmt"

	"git.quad4.io/Go-Libs/micron-parser-go/micron"
)

// Additional runnable examples live in examples/basic (go test ./examples/...).
func ExampleColorToCSS() {
	fmt.Printf("%q\n%q\n", micron.ColorToCSS("abc"), micron.ColorToCSS("default"))
	// Output:
	// "#abc"
	// ""
}
