package network

import (
	"testing"
)

func TestNew(t *testing.T) {
	h := New(ModeRecord)
	
	if h == nil {
		t.Fatal("New() returned nil")
	}
	
	if h.Mode() != ModeRecord {
		t.Errorf("Expected mode 'record', got '%s'", h.Mode())
	}
}

func TestNewWithManifest(t *testing.T) {
	h, err := NewWithManifest(ModeStub, "")
	
	if err != nil {
		t.Errorf("NewWithManifest() returned error: %v", err)
	}
	
	if h == nil {
		t.Fatal("NewWithManifest() returned nil")
	}
	
	if h.Mode() != ModeStub {
		t.Errorf("Expected mode 'stub', got '%s'", h.Mode())
	}
}

func TestSetMode(t *testing.T) {
	h := New(ModeRecord)
	
	h.SetMode(ModeDisabled)
	
	if h.Mode() != ModeDisabled {
		t.Errorf("Expected mode 'disabled', got '%s'", h.Mode())
	}
}

func TestModeDisabled(t *testing.T) {
	h := New(ModeDisabled)
	
	_, err := h.HandleRequest("https://example.com", nil)
	
	if err == nil {
		t.Error("Expected error for disabled mode")
	}
}

func TestRecordMode(t *testing.T) {
	h := New(ModeRecord)
	
	// First, record a response
	h.Record("https://api.example.com/data", []byte(`{"result": "success"}`))
	
	// Now request should return the recording
	resp, err := h.HandleRequest("https://api.example.com/data", nil)
	
	if err != nil {
		t.Errorf("HandleRequest() returned error: %v", err)
	}
	
	if string(resp) != `{"result": "success"}` {
		t.Errorf("Expected response, got %s", string(resp))
	}
}

func TestStubMode(t *testing.T) {
	h := New(ModeStub)
	
	// Add a stub
	h.AddStub("https://api.example.com/data", []byte(`{"stubbed": true}`))
	
	// Request should return the stub
	resp, err := h.HandleRequest("https://api.example.com/data", nil)
	
	if err != nil {
		t.Errorf("HandleRequest() returned error: %v", err)
	}
	
	if string(resp) != `{"stubbed": true}` {
		t.Errorf("Expected stubbed response, got %s", string(resp))
	}
}

func TestStubModeNotFound(t *testing.T) {
	h := New(ModeStub)
	
	_, err := h.HandleRequest("https://unknown.example.com", nil)
	
	if err == nil {
		t.Error("Expected error for unknown stub")
	}
}

func TestDependencyHash(t *testing.T) {
	h := New(ModeRecord)
	
	// Add recordings
	h.Record("https://api.example.com/data1", []byte("response1"))
	h.Record("https://api.example.com/data2", []byte("response2"))
	
	hash1 := h.DependencyHash()
	
	// Hash should be consistent
	hash2 := h.DependencyHash()
	
	if hash1 != hash2 {
		t.Errorf("Hash should be consistent: %s vs %s", hash1, hash2)
	}
	
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
}

func TestDeterministicHash(t *testing.T) {
	// First handler
	h1 := New(ModeRecord)
	h1.Record("https://api.example.com/data", []byte("response"))
	hash1 := h1.DependencyHash()
	
	// Second handler with same data
	h2 := New(ModeRecord)
	h2.Record("https://api.example.com/data", []byte("response"))
	hash2 := h2.DependencyHash()
	
	// Hashes should be identical
	if hash1 != hash2 {
		t.Errorf("Hashes should be identical: %s vs %s", hash1, hash2)
	}
}

func TestDifferentResponsesDifferentHash(t *testing.T) {
	h1 := New(ModeRecord)
	h1.Record("https://api.example.com/data", []byte("response1"))
	hash1 := h1.DependencyHash()
	
	h2 := New(ModeRecord)
	h2.Record("https://api.example.com/data", []byte("response2"))
	hash2 := h2.DependencyHash()
	
	// Hashes should be different
	if hash1 == hash2 {
		t.Error("Hashes should be different with different responses")
	}
}

func TestRecordings(t *testing.T) {
	h := New(ModeRecord)
	
	h.Record("https://api.example.com/data", []byte("response"))
	
	recordings := h.Recordings()
	
	if len(recordings) != 1 {
		t.Errorf("Expected 1 recording, got %d", len(recordings))
	}
	
	if string(recordings["https://api.example.com/data"]) != "response" {
		t.Error("Recording mismatch")
	}
}

func TestStubs(t *testing.T) {
	h := New(ModeStub)
	
	h.AddStub("https://api.example.com/data", []byte("stub"))
	
	stubs := h.Stubs()
	
	if len(stubs) != 1 {
		t.Errorf("Expected 1 stub, got %d", len(stubs))
	}
}

func TestIsAllowed(t *testing.T) {
	h := New(ModeDisabled)
	
	if h.IsAllowed() {
		t.Error("IsAllowed() should return false for disabled mode")
	}
	
	h.SetMode(ModeRecord)
	
	if !h.IsAllowed() {
		t.Error("IsAllowed() should return true for record mode")
	}
	
	h.SetMode(ModeStub)
	
	if !h.IsAllowed() {
		t.Error("IsAllowed() should return true for stub mode")
	}
}
