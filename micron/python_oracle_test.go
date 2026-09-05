// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPythonOracleFingerprint(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not found")
	}
	script := filepath.Join("testdata", "python_oracle.py")
	cases := []string{
		"> Title\nBody\n",
		"`[Go`page.mu]\n",
		"`<user`alice>\n",
		"`{part.mu}\n",
		"#!fg=abc\n#!bg=123\n> Hi\n",
	}
	p := Parser{DarkTheme: true, ForceMonospace: false}
	for _, src := range cases {
		cmd := exec.Command("python3", script)
		cmd.Stdin = nil
		// pass via temp file for simplicity
		tmp := filepath.Join(t.TempDir(), "in.mu")
		if err := os.WriteFile(tmp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("python3", script, tmp).Output()
		if err != nil {
			t.Fatalf("python oracle: %v", err)
		}
		var py SemanticFingerprint
		if err := json.Unmarshal(out, &py); err != nil {
			t.Fatalf("json: %v raw=%s", err, out)
		}
		goFP := p.Parse(src).Fingerprint()
		if len(goFP.Sections) != len(py.Sections) {
			t.Fatalf("sections len go=%v py=%v for %q", goFP.Sections, py.Sections, src)
		}
		for i := range goFP.Sections {
			if goFP.Sections[i] != py.Sections[i] {
				t.Fatalf("section[%d] go=%d py=%d for %q", i, goFP.Sections[i], py.Sections[i], src)
			}
		}
		if len(goFP.Links) != len(py.Links) {
			t.Fatalf("links go=%d py=%d for %q\ngo=%#v\npy=%#v", len(goFP.Links), len(py.Links), src, goFP.Links, py.Links)
		}
		if len(goFP.Fields) != len(py.Fields) {
			t.Fatalf("fields go=%d py=%d for %q", len(goFP.Fields), len(py.Fields), src)
		}
		if len(goFP.Partials) != len(py.Partials) {
			t.Fatalf("partials go=%d py=%d for %q", len(goFP.Partials), len(py.Partials), src)
		}
		if goFP.Headers.FG != py.Headers.FG || goFP.Headers.BG != py.Headers.BG {
			t.Fatalf("headers go=%+v py=%+v for %q", goFP.Headers, py.Headers, src)
		}
	}
}

func TestPythonOracleGuideOptional(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not found")
	}
	path := filepath.Join("testdata", "nomadnet_guide_official.mu")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	script := filepath.Join("testdata", "python_oracle.py")
	out, err := exec.Command("python3", script, path).Output()
	if err != nil {
		t.Fatalf("python oracle: %v", err)
	}
	var py SemanticFingerprint
	if err := json.Unmarshal(out, &py); err != nil {
		t.Fatal(err)
	}
	goFP := (&Parser{DarkTheme: true}).Parse(string(src)).Fingerprint()
	if len(goFP.Sections) == 0 || len(py.Sections) == 0 {
		t.Fatalf("expected sections in guide fingerprint go=%d py=%d", len(goFP.Sections), len(py.Sections))
	}
	// Guide is large; require same section depth sequence when lengths match,
	// otherwise require Go section count within 10% of Python (scanner vs IR).
	if len(goFP.Sections) == len(py.Sections) {
		for i := range goFP.Sections {
			if goFP.Sections[i] != py.Sections[i] {
				t.Fatalf("guide section[%d] go=%d py=%d", i, goFP.Sections[i], py.Sections[i])
			}
		}
	} else {
		diff := len(goFP.Sections) - len(py.Sections)
		if diff < 0 {
			diff = -diff
		}
		if diff*10 > len(py.Sections)+1 {
			t.Fatalf("guide section count diverge go=%d py=%d", len(goFP.Sections), len(py.Sections))
		}
	}
}
