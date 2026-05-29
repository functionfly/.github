package studio

import (
	"testing"
)

func TestHashLicenseKeyDeterministic(t *testing.T) {
	a := hashLicenseKey("FFLIC-ABCDEF1234567890")
	b := hashLicenseKey("FFLIC-ABCDEF1234567890")
	if a != b {
		t.Fatalf("expected stable hash, got %q and %q", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(a))
	}
}

func TestGenerateLicenseKeyFormat(t *testing.T) {
	key, hash, prefix, err := generateLicenseKey()
	if err != nil {
		t.Fatalf("generateLicenseKey: %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("expected hash length 64, got %d", len(hash))
	}
	if prefix != key[:12] {
		t.Fatalf("expected prefix %q, got %q", key[:12], prefix)
	}
	if key[:6] != "FFLIC-" {
		t.Fatalf("expected FFLIC- prefix, got %q", key)
	}
}

func TestValidSPDXLicense(t *testing.T) {
	for _, lic := range []string{"mit", "apache", "gpl", "proprietary", "custom"} {
		if !validSPDXLicense(lic) {
			t.Fatalf("expected %q to be valid", lic)
		}
	}
	if validSPDXLicense("bsd") {
		t.Fatal("expected bsd to be invalid")
	}
}

func TestValidCommercialType(t *testing.T) {
	for _, typ := range []string{"open", "restricted", "commercial"} {
		if !validCommercialType(typ) {
			t.Fatalf("expected %q to be valid", typ)
		}
	}
}

func TestMaskLicenseKey(t *testing.T) {
	masked := maskLicenseKey("FFLIC-ABCD")
	if masked != "FFLIC-ABCD****" {
		t.Fatalf("unexpected mask: %q", masked)
	}
}
