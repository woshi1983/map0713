package submodule

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestBrandingNameValidation(t *testing.T) {
	for _, name := range []string{"应用现代化平台", "Modern Application Platform", "ACME 云平台"} {
		if _, err := validateBrandingName(name); err != nil {
			t.Errorf("valid name %q: %v", name, err)
		}
	}
	for _, name := range []string{"", "  ", "abc\nxyz", "<script>", strings.Repeat("x", 61)} {
		if _, err := validateBrandingName(name); err == nil {
			t.Errorf("invalid name %q accepted", name)
		}
	}
}

func TestBrandingLogoValidation(t *testing.T) {
	valid := image.NewRGBA(image.Rect(0, 0, 1, 1))
	valid.Set(0, 0, color.RGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, valid); err != nil {
		t.Fatal(err)
	}
	if _, mime, err := parseBrandingLogo("data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())); err != nil || mime != "image/png" {
		t.Fatalf("valid PNG rejected: %s, %v", mime, err)
	}
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if _, _, err := parseBrandingLogo("data:image/png;base64," + base64.StdEncoding.EncodeToString(png)); err == nil {
		t.Fatal("truncated PNG accepted")
	}
	if _, _, err := parseBrandingLogo("data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte("<svg/>"))); err == nil {
		t.Fatal("SVG accepted")
	}
	if _, _, err := parseBrandingLogo("data:image/png;base64," + strings.Repeat("A", 400000)); err == nil {
		t.Fatal("oversized image accepted")
	}
}
