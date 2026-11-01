package submodule

import (
	"encoding/base64"
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
