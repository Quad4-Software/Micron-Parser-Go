// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"testing"
)

func convert(t *testing.T, src string) string {
	t.Helper()
	p := Parser{DarkTheme: true, ForceMonospace: false}
	return p.ConvertMicronToHTML(src)
}

func requireImage(t *testing.T, src string) string {
	t.Helper()
	out := convert(t, src)
	if !strings.Contains(out, `class="mu-image"`) {
		t.Fatalf("expected mu-image placeholder for %q, got %s", src, out)
	}
	if strings.Contains(out, `class="Mu-nl"`) {
		t.Fatalf("expected no Mu-nl anchor for %q, got %s", src, out)
	}
	return out
}

func requirePlainLink(t *testing.T, src string) string {
	t.Helper()
	out := convert(t, src)
	if strings.Contains(out, `class="mu-image"`) {
		t.Fatalf("expected normal link for %q, got %s", src, out)
	}
	if !strings.Contains(out, `class="Mu-nl"`) {
		t.Fatalf("expected Mu-nl anchor for %q, got %s", src, out)
	}
	return out
}

func TestImageLinkBasic(t *testing.T) {
	out := requireImage(t, "`[River valley`:/media/harbour.png`w=400]")
	if !strings.Contains(out, `data-mu-image-url=":/media/harbour.png"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `data-mu-image-alt="River valley"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `data-mu-image-path=":/media/harbour.png"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `data-mu-image-w="400"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `role="img"`) || !strings.Contains(out, `aria-label="River valley"`) {
		t.Fatal(out)
	}
}

func TestImageLinkMediaExtensions(t *testing.T) {
	for _, ext := range []string{"webp", "png", "jpg", "jpeg", "bmp", "gif", "tiff"} {
		requireImage(t, "`[a`:/media/x."+ext+"`img=1]")
	}
	// Uppercase extension still matches.
	requireImage(t, "`[a`:/media/x.PNG`img=1]")
	// Media path alone detects an image even without img=1.
	requireImage(t, "`[a`:/media/x.png`w=10]")
	// file/ paths only allow webp.
	requireImage(t, "`[a`:/file/x.webp`img=1]")
}

func TestImageLinkRejectedPaths(t *testing.T) {
	requirePlainLink(t, "`[a`:/file/x.png`img=1]")
	requirePlainLink(t, "`[a`:/media/x.exe`img=1]")
	requirePlainLink(t, "`[a`:/other/x.png`img=1]")
	requirePlainLink(t, "`[a`:/media/../etc/pass.webp`img=1]")
	requirePlainLink(t, "`[a`:/media/x*.png`img=1]")
}

func TestImageLinkEmptyAlt(t *testing.T) {
	requirePlainLink(t, "`[`:/media/x.png]")
	requirePlainLink(t, "`[`:/media/x.png`img=1]")
	requirePlainLink(t, "`[   `:/media/x.png`img=1]")
}

func TestImageLinkHashURL(t *testing.T) {
	out := requireImage(t, "`[a`0123456789abcdef0123456789abcdef:/media/x.png`img=1]")
	if !strings.Contains(out, `data-mu-image-path="0123456789abcdef0123456789abcdef:/media/x.png"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `data-mu-image-url="0123456789abcdef0123456789abcdef:/media/x.png"`) {
		t.Fatal(out)
	}
	// Bad hash falls back to a normal link.
	requirePlainLink(t, "`[a`notahash:/media/x.png`img=1]")
}

func TestImageLinkOptions(t *testing.T) {
	out := requireImage(t, "`[a`:/media/x.png`img=1;w=400;h=267;s=18088;k=abc-1;a=right;profile=fast]")
	for _, want := range []string{
		`data-mu-image-w="400"`,
		`data-mu-image-h="267"`,
		`data-mu-image-s="18088"`,
		`data-mu-image-k="abc-1"`,
		`data-mu-image-a="right"`,
		`data-mu-image-profile="fast"`,
		`width:400px`,
		`min-height:267px`,
		`class="mu-image-size">17.7 kB`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestImageLinkDefaults(t *testing.T) {
	out := requireImage(t, "`[a`:/media/x.png`img=1]")
	// Align defaults to left and is always emitted.
	if !strings.Contains(out, `data-mu-image-a="left"`) {
		t.Fatal(out)
	}
	for _, absent := range []string{
		"data-mu-image-w=",
		"data-mu-image-h=",
		"data-mu-image-s=",
		"data-mu-image-k=",
		"data-mu-image-profile=",
		"mu-image-size",
	} {
		if strings.Contains(out, absent) {
			t.Fatalf("unexpected %q in %s", absent, out)
		}
	}
	// No inline style on the image div when w and h are unset.
	if !strings.Contains(out, `role="img" aria-label="a">`) {
		t.Fatal(out)
	}
}

func TestImageLinkClamping(t *testing.T) {
	out := requireImage(t, "`[a`:/media/x.png`img=1;w=99999;h=99999]")
	if !strings.Contains(out, `data-mu-image-w="8192"`) || !strings.Contains(out, `data-mu-image-h="8192"`) {
		t.Fatal(out)
	}
	for _, bad := range []string{"w=-5", "w=abc", "w=0", "w="} {
		out := requireImage(t, "`[a`:/media/x.png`img=1;"+bad+"]")
		if strings.Contains(out, "data-mu-image-w=") {
			t.Fatalf("expected no w attribute for %q, got %s", bad, out)
		}
	}
	// Fractional values floor before clamping.
	out = requireImage(t, "`[a`:/media/x.png`img=1;w=400.9]")
	if !strings.Contains(out, `data-mu-image-w="400"`) {
		t.Fatal(out)
	}
	// Size hint clamps to 100 MiB.
	out = requireImage(t, "`[a`:/media/x.png`img=1;s=999999999]")
	if !strings.Contains(out, `data-mu-image-s="104857600"`) {
		t.Fatal(out)
	}
}

func TestImageLinkKeyAndProfileSanitize(t *testing.T) {
	out := requireImage(t, "`[a`:/media/x.png`img=1;k=bad key!;profile=bad*profile]")
	if strings.Contains(out, "data-mu-image-k=") {
		t.Fatal(out)
	}
	if strings.Contains(out, "data-mu-image-profile=") {
		t.Fatal(out)
	}
	// Key is truncated to 64 chars.
	long := strings.Repeat("a", 80)
	out = requireImage(t, "`[a`:/media/x.png`img=1;k="+long+"]")
	if !strings.Contains(out, `data-mu-image-k="`+strings.Repeat("a", 64)+`"`) {
		t.Fatal(out)
	}
}

func TestImageLinkURLSuffixStripped(t *testing.T) {
	requireImage(t, "`[a`:/media/x.png?dl=1`img=1]")
	requireImage(t, "`[a`:/media/x.png#frag`img=1]")
	requireImage(t, "`[a`:/media/x.png?dl=1#frag`img=1]")
	// Suffix that hides a bad extension still rejects.
	requirePlainLink(t, "`[a`:/media/x.exe?dl=1`img=1]")
}

func TestImageLinkEscaping(t *testing.T) {
	out := requireImage(t, "`[a<b\"c`:/media/x.png`img=1]")
	if !strings.Contains(out, `data-mu-image-alt="a&lt;b&#34;c"`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `<span class="mu-image-alt">a&lt;b&#34;c</span>`) {
		t.Fatal(out)
	}
	if strings.Contains(out, `alt="a<b"c"`) {
		t.Fatal(out)
	}
}

func TestImageLinkStructure(t *testing.T) {
	out := requireImage(t, "`[a`:/media/x.png`img=1]")
	if !strings.Contains(out, `<span class="mu-image-actions"><a class="mu-image-action" data-mu-image-action="load" role="button" tabindex="0">Load image</a></span>`) {
		t.Fatal(out)
	}
	if !strings.Contains(out, `<img class="mu-image-output" alt="a" hidden="">`) {
		t.Fatal(out)
	}
}

func TestImageAltTruncation(t *testing.T) {
	long := strings.Repeat("x", 300)
	out := requireImage(t, "`["+long+"`:/media/x.png`img=1]")
	want := `data-mu-image-alt="` + strings.Repeat("x", 240) + `"`
	if !strings.Contains(out, want) {
		t.Fatal(out)
	}
}
