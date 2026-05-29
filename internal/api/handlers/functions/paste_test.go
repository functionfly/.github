package functions

import (
	"testing"
)

func TestValidateFunctionName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: "helloWorld", wantErr: false},
		{name: "valid with underscore", input: "hello_world", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "starts with number", input: "1hello", wantErr: true},
		{name: "contains space", input: "hello world", wantErr: true},
		{name: "contains slash", input: "hello/world", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFunctionName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestSanitizeFunctionName(t *testing.T) {
	got := sanitizeFunctionName("  hello\x00world\n")
	if got != "helloworld" {
		t.Fatalf("sanitizeFunctionName() = %q, want %q", got, "helloworld")
	}
}

func TestSanitizeCode(t *testing.T) {
	got := sanitizeCode("print('hi')\x00")
	if got != "print('hi')" {
		t.Fatalf("sanitizeCode() = %q, want %q", got, "print('hi')")
	}
}

func TestValidateProviders(t *testing.T) {
	if err := validateProviders([]string{"cloud"}); err != nil {
		t.Fatalf("expected valid providers, got %v", err)
	}
	if err := validateProviders([]string{}); err == nil {
		t.Fatal("expected error for empty providers")
	}
	if err := validateProviders([]string{"cloud", "invalid"}); err == nil {
		t.Fatal("expected error for invalid provider")
	}
}

func TestValidateRegion(t *testing.T) {
	if err := validateRegion("us-east-1"); err != nil {
		t.Fatalf("expected valid region, got %v", err)
	}
	if err := validateRegion("invalid-region"); err == nil {
		t.Fatal("expected error for invalid region")
	}
}

func TestLanguageTags(t *testing.T) {
	tags, err := languageTags("python")
	if err != nil {
		t.Fatalf("languageTags() error = %v", err)
	}
	if string(tags) != `["python"]` {
		t.Fatalf("languageTags() = %s, want %q", tags, `["python"]`)
	}

	maliciousTags, err := languageTags(`python","injected`)
	if err != nil {
		t.Fatalf("languageTags() error = %v", err)
	}
	if string(maliciousTags) != `["python\",\"injected"]` {
		t.Fatalf("languageTags() should escape malicious input, got %s", maliciousTags)
	}
}
