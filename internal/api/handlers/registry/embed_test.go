package registry

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func makeTestFunction(author, name, description string) *registry.RegistryFunction {
	return &registry.RegistryFunction{
		ID:          uuid.New(),
		Author:      author,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	}
}

func makeTestVersion(version string) *registry.RegistryFunctionVersion {
	return &registry.RegistryFunctionVersion{
		ID:      uuid.New(),
		Version: version,
	}
}

// ── generateEmbedScript ───────────────────────────────────────────────────────

func TestGenerateEmbedScript_ContainsAuthorAndName(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "Converts text to a URL slug")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, `"acme"`) {
		t.Error("expected script to contain author \"acme\"")
	}
	if !strings.Contains(script, `"slugify"`) {
		t.Error("expected script to contain name \"slugify\"")
	}
}

func TestGenerateEmbedScript_PinnedVersionInScript(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("2.3.1")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "2.3.1", opts)

	if !strings.Contains(script, "2.3.1") {
		t.Error("expected pinned version 2.3.1 to appear in script")
	}
}

func TestGenerateEmbedScript_DefaultNamespace(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, `"ff"`) {
		t.Error("expected default namespace \"ff\" in script")
	}
}

func TestGenerateEmbedScript_CustomNamespace(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "myApp", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, `"myApp"`) {
		t.Error("expected custom namespace \"myApp\" in script")
	}
}

func TestGenerateEmbedScript_AutoloadFalse(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: false, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, "autoload: false") {
		t.Error("expected autoload: false in script config")
	}
}

func TestGenerateEmbedScript_UIEnabled(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: true, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, "ui: true") {
		t.Error("expected ui: true in script config")
	}
}

func TestGenerateEmbedScript_ContainsRunFunction(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	for _, symbol := range []string{"function run(", "function form(", "function on(", "function widget("} {
		if !strings.Contains(script, symbol) {
			t.Errorf("expected script to contain %q", symbol)
		}
	}
}

func TestGenerateEmbedScript_ContainsPublicAPI(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, "global[NAMESPACE] = api") {
		t.Error("expected global namespace assignment in script")
	}
}

func TestGenerateEmbedScript_NilVersion(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	// Should not panic when fnVersion is nil
	script := generateEmbedScript(fn, nil, "", opts)

	if !strings.Contains(script, "latest") {
		t.Error("expected 'latest' version when fnVersion is nil")
	}
}

// ── parseEmbedOptions ─────────────────────────────────────────────────────────

func TestParseEmbedOptions_Defaults(t *testing.T) {
	opts := parseEmbedOptions("", "", "", "")

	if opts.Namespace != "ff" {
		t.Errorf("expected default namespace 'ff', got %q", opts.Namespace)
	}
	if !opts.Autoload {
		t.Error("expected Autoload to be true by default")
	}
	if opts.UI {
		t.Error("expected UI to be false by default")
	}
	if opts.Theme != "auto" {
		t.Errorf("expected default theme 'auto', got %q", opts.Theme)
	}
}

func TestParseEmbedOptions_CustomValues(t *testing.T) {
	opts := parseEmbedOptions("myNS", "false", "true", "dark")

	if opts.Namespace != "myNS" {
		t.Errorf("expected namespace 'myNS', got %q", opts.Namespace)
	}
	if opts.Autoload {
		t.Error("expected Autoload to be false")
	}
	if !opts.UI {
		t.Error("expected UI to be true")
	}
	if opts.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %q", opts.Theme)
	}
}

func TestParseEmbedOptions_SanitizesNamespace(t *testing.T) {
	// Namespace with invalid characters should be stripped
	opts := parseEmbedOptions("my-app<script>", "", "", "")

	if strings.ContainsAny(opts.Namespace, "-<>") {
		t.Errorf("namespace should not contain special chars, got %q", opts.Namespace)
	}
}

func TestParseEmbedOptions_InvalidThemeFallsBackToAuto(t *testing.T) {
	opts := parseEmbedOptions("", "", "", "rainbow")

	if opts.Theme != "auto" {
		t.Errorf("invalid theme should fall back to 'auto', got %q", opts.Theme)
	}
}
