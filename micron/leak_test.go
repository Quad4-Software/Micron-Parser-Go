// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"runtime"
	"testing"
	"time"
)

func TestNoGoroutineLeakAcrossRepeatedConversions(t *testing.T) {
	p := Parser{DarkTheme: true, ForceMonospace: false}
	base := runtime.NumGoroutine()
	for i := 0; i < 2000; i++ {
		_ = p.ConvertMicronToHTML("`!hello\n`B123world\n`b")
	}
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > base+2 {
		t.Fatalf("possible goroutine leak: base=%d after=%d", base, after)
	}
}
