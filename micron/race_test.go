// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"sync"
	"testing"
)

func TestConcurrentConvertMicronToHTML(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: true}
	inputs := []string{
		"> a\n`!x",
		"`B123line\n`breset",
		"`[x`host`a=b|f]",
		"`<24|name`v>",
	}
	var wg sync.WaitGroup
	for n := 0; n < 128; n++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out := p.ConvertMicronToHTML(inputs[i%len(inputs)])
			if strings.TrimSpace(out) == "" {
				t.Errorf("empty output for case %d", i)
			}
		}(n)
	}
	wg.Wait()
}
