package execution

import (
	"encoding/json"
	"testing"
)

func TestParseMicroVMManifest_Empty(t *testing.T) {
	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(nil)
	if pkgs != nil || net != nil || pkgCache || strictNet {
		t.Errorf("expected all zero values for nil manifest, got pkgs=%v net=%v pkgCache=%v strictNet=%v", pkgs, net, pkgCache, strictNet)
	}
}

func TestParseMicroVMManifest_EmptyJSON(t *testing.T) {
	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(json.RawMessage(`{}`))
	if len(pkgs) != 0 || len(net) != 0 || pkgCache || strictNet {
		t.Errorf("expected empty for empty JSON, got pkgs=%v net=%v pkgCache=%v strictNet=%v", pkgs, net, pkgCache, strictNet)
	}
}

func TestParseMicroVMManifest_InvalidJSON(t *testing.T) {
	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(json.RawMessage(`{invalid`))
	if pkgs != nil || net != nil || pkgCache || strictNet {
		t.Errorf("expected all zero values for invalid JSON, got pkgs=%v net=%v pkgCache=%v strictNet=%v", pkgs, net, pkgCache, strictNet)
	}
}

func TestParseMicroVMManifest_NestedFunctionFormat(t *testing.T) {
	manifest := json.RawMessage(`{
		"function": {
			"python": {
				"packages": ["numpy", "pandas"]
			},
			"enterprise": {
				"network_allowlist": ["api.example.com", "db.internal"],
				"package_cache_enabled": true,
				"strict_network_whitelist": true
			}
		}
	}`)

	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(manifest)

	if len(pkgs) != 2 || pkgs[0] != "numpy" || pkgs[1] != "pandas" {
		t.Errorf("packages = %v, want [numpy pandas]", pkgs)
	}
	if len(net) != 2 || net[0] != "api.example.com" || net[1] != "db.internal" {
		t.Errorf("network = %v, want [api.example.com db.internal]", net)
	}
	if !pkgCache {
		t.Error("expected pkgCache=true")
	}
	if !strictNet {
		t.Error("expected strictNet=true")
	}
}

func TestParseMicroVMManifest_FlatFormat(t *testing.T) {
	manifest := json.RawMessage(`{
		"python": {
			"packages": ["requests"]
		},
		"enterprise": {
			"network_allowlist": ["10.0.0.1"],
			"package_cache_enabled": false,
			"strict_network_whitelist": false
		}
	}`)

	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(manifest)

	if len(pkgs) != 1 || pkgs[0] != "requests" {
		t.Errorf("packages = %v, want [requests]", pkgs)
	}
	if len(net) != 1 || net[0] != "10.0.0.1" {
		t.Errorf("network = %v, want [10.0.0.1]", net)
	}
	if pkgCache {
		t.Error("expected pkgCache=false")
	}
	if strictNet {
		t.Error("expected strictNet=false")
	}
}

func TestParseMicroVMManifest_MergedFormats(t *testing.T) {
	manifest := json.RawMessage(`{
		"function": {
			"python": {
				"packages": ["numpy"]
			},
			"enterprise": {
				"network_allowlist": ["api.example.com"]
			}
		},
		"python": {
			"packages": ["pandas"]
		},
		"enterprise": {
			"network_allowlist": ["db.internal"]
		}
	}`)

	pkgs, net, _, _ := parseMicroVMManifest(manifest)

	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages merged, got %d: %v", len(pkgs), pkgs)
	}
	if len(net) != 2 {
		t.Errorf("expected 2 network entries merged, got %d: %v", len(net), net)
	}
}

func TestParseMicroVMManifest_PartialData(t *testing.T) {
	manifest := json.RawMessage(`{
		"function": {
			"python": {
				"packages": ["numpy"]
			}
		}
	}`)

	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(manifest)

	if len(pkgs) != 1 || pkgs[0] != "numpy" {
		t.Errorf("packages = %v, want [numpy]", pkgs)
	}
	if len(net) != 0 {
		t.Errorf("network = %v, want empty", net)
	}
	if pkgCache {
		t.Error("expected pkgCache=false")
	}
	if strictNet {
		t.Error("expected strictNet=false")
	}
}

func TestParseMicroVMManifest_EmptyArrays(t *testing.T) {
	manifest := json.RawMessage(`{
		"function": {
			"python": { "packages": [] },
			"enterprise": { "network_allowlist": [] }
		}
	}`)

	pkgs, net, _, _ := parseMicroVMManifest(manifest)

	if len(pkgs) != 0 {
		t.Errorf("expected empty packages, got %v", pkgs)
	}
	if len(net) != 0 {
		t.Errorf("expected empty network, got %v", net)
	}
}
