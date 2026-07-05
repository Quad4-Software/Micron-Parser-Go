// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"os"
	"path/filepath"
	"strings"
)

func regressionMarkupSeedsStatic() []string {
	dir := filepath.Join("testdata", "regressions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mu") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, string(raw))
	}
	return out
}
