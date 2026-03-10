package registry

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// ── isOriginAllowed ───────────────────────────────────────────────────────────

func TestIsOriginAllowed_Wildcard(t *testing.T) {
	if !isOriginAllowed("https://example.com", []string{"*"}) {
		t.Error("wildcard should allow any origin")
	}
}

func TestIsOriginAllowed_ExactMatch(t *testing.T) {
	if !isOriginAllowed("https://example.com", []string{"https://example.com"}) {
		t.Error("exact match should be allowed")
	}
}

func TestIsOriginAllowed_CaseInsensitive(t *testing.T) {
	if !isOriginAllowed("https://EXAMPLE.COM", []string{"https://example.com"}) {
		t.Error("origin matching should be case-insensitive")
	}
}

func TestIsOriginAllowed_NotInList(t *testing.T) {
	if isOriginAllowed("https://evil.com", []string{"https://example.com", "https://app.example.com"}) {
		t.Error("origin not in list should be denied")
	}
}

func TestIsOriginAllowed_EmptyList(t *testing.T) {
	if isOriginAllowed("https://example.com", []string{}) {
		t.Error("empty allowed list should deny all origins")
	}
}

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

// ── Phase 3: X-Embed-Origin header in generated script ───────────────────────

func TestGenerateEmbedScript_ContainsEmbedOriginHeader(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	if !strings.Contains(script, "X-Embed-Origin") {
		t.Error("expected script to send X-Embed-Origin header for analytics tracking")
	}
	if !strings.Contains(script, "window.location.origin") {
		t.Error("expected script to use window.location.origin as embed origin value")
	}
}

// ── Phase 2: Form & Widget Tests ─────────────────────────────────────────

func TestGenerateEmbedScript_FormUsesFormData(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Form should use FormData API for serialization
	if !strings.Contains(script, "new FormData(formEl)") {
		t.Error("expected form function to use FormData API")
	}
	// Form should return cleanup function
	if !strings.Contains(script, "return function ()") {
		t.Error("expected form function to return cleanup function")
	}
	if !strings.Contains(script, "removeEventListener") {
		t.Error("expected cleanup function to remove event listener")
	}
}

func TestGenerateEmbedScript_WidgetReturnsCleanup(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Widget should be a function
	if !strings.Contains(script, "function widget(") {
		t.Error("expected widget to be a function")
	}
	// Check that container.innerHTML = '' is used in cleanup
	if !strings.Contains(script, "container.innerHTML = ''") {
		t.Error("expected widget cleanup to clear container")
	}
}

func TestGenerateEmbedScript_WidgetHasDefaultOptions(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Widget should have default title using author/name
	if !strings.Contains(script, "AUTHOR + \"/\" + NAME") {
		t.Error("expected widget to have default title with author/name")
	}
	// Widget should have default placeholder
	if !strings.Contains(script, "Enter input (JSON)") {
		t.Error("expected widget to have default placeholder")
	}
	// Widget should have default button text
	if !strings.Contains(script, "|| \"Run\"") {
		t.Error("expected widget to have default button text")
	}
}

func TestGenerateEmbedScript_WidgetHasLoadingState(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Widget should show loading state
	if !strings.Contains(script, "Running…") {
		t.Error("expected widget to show running state")
	}
	// Widget should handle JSON parse errors
	if !strings.Contains(script, "Invalid JSON input") {
		t.Error("expected widget to handle JSON parse errors")
	}
}

// ── Theme Support Tests ───────────────────────────────────────────────────

func TestGenerateEmbedScript_ThemeSupport_Light(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "light"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Script should contain theme config
	if !strings.Contains(script, "theme: \"light\"") {
		t.Error("expected theme 'light' in script config")
	}
	// Should have theme detection function
	if !strings.Contains(script, "getTheme") {
		t.Error("expected getTheme function for theme detection")
	}
	if !strings.Contains(script, "prefers-color-scheme") {
		t.Error("expected system preference detection for auto theme")
	}
}

func TestGenerateEmbedScript_ThemeSupport_Dark(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "dark"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Script should contain theme config
	if !strings.Contains(script, "theme: \"dark\"") {
		t.Error("expected theme 'dark' in script config")
	}
	// Widget should apply dark theme colors
	if !strings.Contains(script, "#1e1e1e") {
		t.Error("expected dark theme background color in widget")
	}
}

func TestGenerateEmbedScript_ThemeSupport_Auto(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Script should contain theme config
	if !strings.Contains(script, "theme: \"auto\"") {
		t.Error("expected theme 'auto' in script config")
	}
	// Widget should handle auto theme via getTheme
	if !strings.Contains(script, "getTheme(options.theme") {
		t.Error("expected widget to use getTheme for auto detection")
	}
}

func TestGenerateEmbedScript_HTMLEntityEscaping(t *testing.T) {
	fn := makeTestFunction("acme", "slugify", "")
	ver := makeTestVersion("1.0.0")
	opts := EmbedOptions{Namespace: "ff", Autoload: true, UI: false, Theme: "auto"}

	script := generateEmbedScript(fn, ver, "", opts)

	// Should have HTML escape utility
	if !strings.Contains(script, "function escHtml") {
		t.Error("expected escHtml function for XSS prevention")
	}
	// escHtml should escape common XSS vectors
	if !strings.Contains(script, "&lt;") || !strings.Contains(script, "&gt;") {
		t.Error("expected escHtml to escape HTML entities")
	}
}
