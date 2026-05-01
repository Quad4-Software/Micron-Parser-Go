// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"maps"
	"math/rand/v2"
	"strings"
	"testing"
)

const propertyIterations = 800

func randUTF8Chunk(r *rand.Rand, maxBytes int) string {
	n := r.IntN(maxBytes + 1)
	if n == 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		switch r.IntN(8) {
		case 0:
			v := byte(r.IntN(256))
			if v == 0 {
				v = 1
			}
			b[i] = v
		case 1:
			b[i] = '\n'
		case 2:
			b[i] = '`'
		case 3:
			b[i] = '<'
		case 4:
			b[i] = '>'
		case 5:
			b[i] = '"'
		case 6:
			b[i] = '\''
		default:
			b[i] = byte('a' + r.IntN(26))
		}
	}
	return string(b)
}

func randMicronCorpusLine(r *rand.Rand) string {
	tokens := []string{
		"hello", "world", "> section", ">> sub", "-", "`!bold", "`_u", "`*i",
		"`F123col", "`B456bg", "`f", "`b", "`c", "`l", "`r", "`a",
		"`<16|name`v>", "`<!|pwd`sec>", "`[open`example.com`k=v|f1]",
		"`{p:/x.mu}", "# comment", "",
	}
	parts := 1 + r.IntN(4)
	var b strings.Builder
	for i := range parts {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(tokens[r.IntN(len(tokens))])
	}
	return b.String()
}

func randMicronDocument(r *rand.Rand) string {
	nl := 1 + r.IntN(16)
	lines := make([]string, 0, nl)
	for range nl {
		switch r.IntN(5) {
		case 0:
			lines = append(lines, randMicronCorpusLine(r))
		case 1:
			lines = append(lines, randUTF8Chunk(r, 48))
		default:
			lines = append(lines, randMicronCorpusLine(r)+randUTF8Chunk(r, 24))
		}
	}
	return strings.Join(lines, "\n")
}

func randFieldMap(r *rand.Rand, maxKeys int) map[string]string {
	n := r.IntN(maxKeys + 1)
	m := make(map[string]string, n)
	for range n {
		k := randUTF8Chunk(r, 24)
		if k == "" {
			k = "k"
		}
		m[k] = randUTF8Chunk(r, 64)
	}
	return m
}

func cloneStringMap(m map[string]string) map[string]string {
	return maps.Clone(m)
}

func hexDigit(r *rand.Rand) byte {
	const hexd = "0123456789abcdefABCDEF"
	return hexd[r.IntN(len(hexd))]
}

func randValidHex3(r *rand.Rand) string {
	return string([]byte{hexDigit(r), hexDigit(r), hexDigit(r)})
}

func randValidHex6(r *rand.Rand) string {
	var b [6]byte
	for i := range b {
		b[i] = hexDigit(r)
	}
	return string(b[:])
}

func TestPropertyConvertDeterministicAndSafe(t *testing.T) {
	r := rand.New(rand.NewPCG(0x9e3779b97f4a7c15, 0x517cc1b727220a95))
	parsers := []Parser{
		{DarkTheme: true, ForceMonospace: true},
		{DarkTheme: true, ForceMonospace: false},
		{DarkTheme: false, ForceMonospace: true},
		{DarkTheme: false, ForceMonospace: false},
	}
	for range propertyIterations {
		doc := randMicronDocument(r)
		for _, p := range parsers {
			a := p.ConvertMicronToHTML(doc)
			b := p.ConvertMicronToHTML(doc)
			if a != b {
				t.Fatalf("non-deterministic convert for parser %#v", p)
			}
			assertFuzzOutputHTMLSafety(t, a)
		}
	}
}

func TestPropertyFormatNomadnetworkURLIdempotent(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for range propertyIterations {
		s := randUTF8Chunk(r, 128)
		if r.IntN(3) == 0 {
			s = "https://" + randUTF8Chunk(r, 40)
		}
		once := FormatNomadnetworkURL(s)
		twice := FormatNomadnetworkURL(once)
		if once != twice {
			t.Fatalf("idempotence: %q -> %q -> %q", s, once, twice)
		}
	}
}

func TestPropertyBuildRequestPayloadDoesNotMutateFieldsMap(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for range propertyIterations {
		fields := randFieldMap(r, 12)
		snap := cloneStringMap(fields)
		dest := randUTF8Chunk(r, 40)
		if r.IntN(2) == 0 {
			dest = strings.TrimSpace(dest) + "`a=" + randUTF8Chunk(r, 8) + "|b=x"
		}
		spec := randUTF8Chunk(r, 32)
		if r.IntN(4) == 0 {
			spec = "*"
		}
		_ = BuildRequestPayload(fields, dest, spec)
		if !maps.Equal(snap, fields) {
			t.Fatalf("allFields map mutated: before %#v after %#v", snap, fields)
		}
	}
}

func TestPropertyBuildRequestPayloadStarCopiesAllEntries(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	for range propertyIterations {
		fields := randFieldMap(r, 10)
		got := BuildRequestPayload(fields, randUTF8Chunk(r, 20), "*")
		if !maps.Equal(got.Fields, fields) {
			t.Fatalf("star copy mismatch: in %#v out %#v", fields, got.Fields)
		}
	}
}

func TestPropertyParseHeaderTagColorLengths(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	for range propertyIterations {
		var b strings.Builder
		nh := r.IntN(4)
		for range nh {
			switch r.IntN(3) {
			case 0:
				b.WriteString("#!fg=")
				b.WriteString(randValidHex3(r))
			case 1:
				b.WriteString("#!bg=")
				b.WriteString(randValidHex6(r))
			default:
				b.WriteString("#!fg=")
				b.WriteString(randUTF8Chunk(r, 5))
			}
			b.WriteByte('\n')
		}
		b.WriteString(randMicronDocument(r))
		pc := ParseHeaderTags(b.String())
		if pc.FG != "" && len(pc.FG) != 3 && len(pc.FG) != 6 {
			t.Fatalf("FG length: %q", pc.FG)
		}
		if pc.BG != "" && len(pc.BG) != 3 && len(pc.BG) != 6 {
			t.Fatalf("BG length: %q", pc.BG)
		}
	}
}

func TestPropertyCollectFormFieldsRadioLastCheckedWins(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 10))
	for range propertyIterations {
		name := randUTF8Chunk(r, 16)
		if name == "" {
			name = "n"
		}
		nRadio := 2 + r.IntN(8)
		inputs := make([]FieldInput, 0, nRadio)
		want := ""
		for i := range nRadio {
			val := randUTF8Chunk(r, 12)
			if val == "" {
				val = "v"
			}
			checked := r.IntN(3) == 0
			if i == nRadio-1 {
				checked = true
				want = val
			}
			inputs = append(inputs, FieldInput{Type: "radio", Name: name, Value: val, Checked: checked})
		}
		got := CollectFormFields(inputs)
		if got[name] != want {
			t.Fatalf("radio name %q want %q got %#v", name, want, got)
		}
	}
}

func TestPropertyColorToCSSValidHex(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 12))
	for range propertyIterations {
		h3 := randValidHex3(r)
		css3 := ColorToCSS(h3)
		if len(css3) != 4 || css3[0] != '#' {
			t.Fatalf("hex3 %q -> %q", h3, css3)
		}
		h6 := randValidHex6(r)
		css6 := ColorToCSS(h6)
		if len(css6) != 7 || css6[0] != '#' {
			t.Fatalf("hex6 %q -> %q", h6, css6)
		}
	}
}
