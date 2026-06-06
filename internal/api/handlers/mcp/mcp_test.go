package mcp

import (
	"encoding/json"
	"testing"
)

func TestParseFrame_SingleRequest(t *testing.T) {
	raw := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	reqs, errResp := ParseFrame(raw)
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	if len(reqs) != 1 {
		t.Fatalf("want 1 request, got %d", len(reqs))
	}
	if reqs[0].Method != "tools/list" {
		t.Errorf("method = %q, want %q", reqs[0].Method, "tools/list")
	}
	if !reqs[0].HasID() {
		t.Errorf("expected request to have an id")
	}
}

func TestParseFrame_Notification(t *testing.T) {
	raw := []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	reqs, errResp := ParseFrame(raw)
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	if len(reqs) != 1 {
		t.Fatalf("want 1 request, got %d", len(reqs))
	}
	if reqs[0].HasID() {
		t.Errorf("expected notification to have no id")
	}
}

func TestParseFrame_ParseError(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"jsonrpc":"1.0","method":"x"}`),  // wrong version
		[]byte(`{"method":"x"}`),                   // no jsonrpc field
		[]byte(`not json at all`),                   // invalid JSON
		[]byte(`[]`),                                // empty batch
		[]byte(`{"jsonrpc":"2.0"}`),                 // no method
	}
	for i, c := range cases {
		_, errResp := ParseFrame(c)
		if errResp == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}
	}
}

func TestParseFrame_Batch(t *testing.T) {
	raw := []byte(`[
		{"jsonrpc":"2.0","id":1,"method":"tools/list"},
		{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"x__y","arguments":{}}}
	]`)
	reqs, errResp := ParseFrame(raw)
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	if len(reqs) != 2 {
		t.Fatalf("want 2 requests, got %d", len(reqs))
	}
}

func TestMakeError_PreservesID(t *testing.T) {
	resp := MakeError(json.RawMessage(`42`), ErrCodeToolNotFound, "missing", nil)
	if string(resp.ID) != "42" {
		t.Errorf("id = %s, want 42", string(resp.ID))
	}
	if resp.Error.Code != ErrCodeToolNotFound {
		t.Errorf("code = %d, want %d", resp.Error.Code, ErrCodeToolNotFound)
	}
}

func TestSanitizeToolName(t *testing.T) {
	cases := map[string]string{
		"":                                "fn",
		"alice__summarize_pdf":            "alice__summarize_pdf",
		"alice/summarize":                 "alice_summarize",
		"hello world":                     "hello_world",
		"___":                             "fn",
		"a-very-long-tool-name-that-exceeds-the-sixty-four-character-limit-and-should-be-truncated-cleanly": "",
	}
	for in, want := range cases {
		got := SanitizeToolName(in)
		if in != "" && len(in) > ToolNameMaxLength {
			if len(got) > ToolNameMaxLength {
				t.Errorf("SanitizeToolName(%q) = %q (len %d) exceeds limit", in, got, len(got))
			}
			continue
		}
		if got != want {
			t.Errorf("SanitizeToolName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseToolName(t *testing.T) {
	author, name, override := ParseToolName("alice__summarize_pdf")
	if author != "alice" || name != "summarize_pdf" || override != "" {
		t.Errorf("got (%q,%q,%q), want (alice,summarize_pdf,\"\")", author, name, override)
	}

	author, name, override = ParseToolName("custom_alias")
	if author != "" || name != "" || override != "custom_alias" {
		t.Errorf("override path failed: got (%q,%q,%q)", author, name, override)
	}
}

func TestEncodeDecodeCursor_Roundtrip(t *testing.T) {
	original := "550e8400-e29b-41d4-a716-446655440000"
	enc := EncodeCursor(original)
	if enc == "" {
		t.Fatal("empty encode")
	}
	dec := DecodeCursor(enc)
	if dec != original {
		t.Errorf("decode = %q, want %q", dec, original)
	}
}

func TestEncodeCursor_Empty(t *testing.T) {
	if got := EncodeCursor(""); got != "" {
		t.Errorf("EncodeCursor(\"\") = %q, want \"\"", got)
	}
	if got := DecodeCursor(""); got != "" {
		t.Errorf("DecodeCursor(\"\") = %q, want \"\"", got)
	}
}

func TestMaxDepth(t *testing.T) {
	cases := map[string]int{
		`{}`:            1,
		`{"a":1}`:       1,
		`{"a":{"b":1}}`: 2,
		`[1,2,3]`:       1,
		`[[1]]`:         2,
		`1`:             0,
		`null`:          0,
	}
	for in, want := range cases {
		if got := MaxDepth(json.RawMessage(in)); got != want {
			t.Errorf("MaxDepth(%s) = %d, want %d", in, got, want)
		}
	}
}

func TestMaxDepth_InvalidJSON(t *testing.T) {
	if got := MaxDepth(json.RawMessage(`{not json`)); got != -1 {
		t.Errorf("MaxDepth(invalid) = %d, want -1", got)
	}
}

func TestBuildToolDefinition_Basic(t *testing.T) {
	s := ToMCPSettings{
		FunctionID: "abc",
		Author:     "alice",
		Name:       "summarize_pdf",
		Version:    "1.2.0",
		Title:      "Summarize PDF",
		Description: "Extracts and summarizes the key points of any PDF document.",
		Category:   "document-processing",
		Tags:       []string{"pdf", "summary"},
		TrustScore: 92.0,
		TrustTier:  "verified",
		VerifiedMCP: true,
		Runtime:    "node20",
		Manifest: []byte(`{
			"input": {
				"schema": {
					"type": "object",
					"properties": {"url": {"type": "string", "format": "uri"}},
					"required": ["url"]
				}
			}
		}`),
	}
	td := BuildToolDefinition(s)
	if td.Name != "alice__summarize_pdf" {
		t.Errorf("name = %q, want %q", td.Name, "alice__summarize_pdf")
	}
	if td.Title != "Summarize PDF" {
		t.Errorf("title = %q", td.Title)
	}
	if !td.Annotations.ReadOnlyHint {
		t.Errorf("verified_mcp should imply readOnlyHint")
	}
	if td.Annotations.Category != "document-processing" {
		t.Errorf("category = %q", td.Annotations.Category)
	}
	if td.Meta.FunctionFly.TrustScore != 92.0 {
		t.Errorf("trust score = %f", td.Meta.FunctionFly.TrustScore)
	}
	if !td.Meta.FunctionFly.VerifiedMCP {
		t.Errorf("verified_mcp not propagated")
	}
	if len(td.InputSchema) == 0 {
		t.Errorf("input schema empty")
	}
}

func TestBuildToolDefinition_Override(t *testing.T) {
	s := ToMCPSettings{
		Author:           "alice",
		Name:             "x",
		ToolNameOverride: "my-custom-alias",
	}
	td := BuildToolDefinition(s)
	if td.Name != "my-custom-alias" {
		t.Errorf("name = %q, want override", td.Name)
	}
}
