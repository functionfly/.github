package trace

import (
	"testing"
)

func TestNew(t *testing.T) {
	tr := New(100)
	
	if tr == nil {
		t.Fatal("New() returned nil")
	}
	
	if tr.chunkSize != 100 {
		t.Errorf("Expected chunk size 100, got %d", tr.chunkSize)
	}
	
	if !tr.IsEnabled() {
		t.Error("Trace should be enabled by default")
	}
}

func TestNewDefaultChunkSize(t *testing.T) {
	tr := New(0)
	
	if tr.chunkSize != 1000 {
		t.Errorf("Expected default chunk size 1000, got %d", tr.chunkSize)
	}
}

func TestRecord(t *testing.T) {
	tr := New(10)
	
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	tr.Record("memory", 101, []byte(`{"addr": 0x1000}`))
	
	if tr.EntryCount() != 2 {
		t.Errorf("Expected 2 entries, got %d", tr.EntryCount())
	}
}

func TestDisable(t *testing.T) {
	tr := New(10)
	
	tr.Disable()
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	
	if tr.EntryCount() != 0 {
		t.Errorf("Expected 0 entries (disabled), got %d", tr.EntryCount())
	}
}

func TestEnable(t *testing.T) {
	tr := New(10)
	
	tr.Disable()
	tr.Enable()
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	
	if tr.EntryCount() != 1 {
		t.Errorf("Expected 1 entry, got %d", tr.EntryCount())
	}
}

func TestHash(t *testing.T) {
	tr := New(10)
	
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	
	hash := tr.Hash()
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
}

func TestDeterministicHash(t *testing.T) {
	// First trace
	tr1 := New(10)
	tr1.Record("syscall", 100, []byte(`{"id": 1}`))
	tr1.Record("memory", 101, []byte(`{"addr": 0x1000}`))
	hash1 := tr1.Hash()
	
	// Second trace with same data
	tr2 := New(10)
	tr2.Record("syscall", 100, []byte(`{"id": 1}`))
	tr2.Record("memory", 101, []byte(`{"addr": 0x1000}`))
	hash2 := tr2.Hash()
	
	// Hashes should be identical
	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: %s vs %s", hash1, hash2)
	}
}

func TestDifferentEntriesDifferentHash(t *testing.T) {
	tr1 := New(10)
	tr1.Record("syscall", 100, []byte(`{"id": 1}`))
	hash1 := tr1.Hash()
	
	tr2 := New(10)
	tr2.Record("syscall", 100, []byte(`{"id": 2}`))
	hash2 := tr2.Hash()
	
	// Hashes should be different
	if hash1 == hash2 {
		t.Error("Hash should differ with different entries")
	}
}

func TestChunking(t *testing.T) {
	tr := New(3) // Chunk size of 3
	
	// Fill first chunk
	tr.Record("syscall", 1, []byte("1"))
	tr.Record("syscall", 2, []byte("2"))
	tr.Record("syscall", 3, []byte("3"))
	
	chunks := tr.Chunks()
	
	// Should have 1 chunk
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk, got %d", len(chunks))
	}
	
	// Add one more to trigger new chunk
	tr.Record("syscall", 4, []byte("4"))
	
	chunks = tr.Chunks()
	
	// Should now have 2 chunks
	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
	
	// Second chunk should have prev hash
	if chunks[1].PrevHash == "" {
		t.Error("Second chunk should have prev hash")
	}
}

func TestClear(t *testing.T) {
	tr := New(10)
	
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	
	if tr.EntryCount() != 1 {
		t.Errorf("Expected 1 entry, got %d", tr.EntryCount())
	}
	
	tr.Clear()
	
	if tr.EntryCount() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", tr.EntryCount())
	}
	
	// Hash should be recalculated
	hash := tr.Hash()
	if hash != "" {
		t.Error("Hash should be empty after clear")
	}
}

func TestReset(t *testing.T) {
	tr := New(10)
	
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	tr.Reset(20) // New chunk size
	
	if tr.chunkSize != 20 {
		t.Errorf("Expected chunk size 20, got %d", tr.chunkSize)
	}
	
	if tr.EntryCount() != 0 {
		t.Errorf("Expected 0 entries after reset, got %d", tr.EntryCount())
	}
}

func TestCompactHash(t *testing.T) {
	tr := New(10)
	
	tr.Record("syscall", 100, []byte(`{"id": 1}`))
	
	hash := tr.Hash()
	compact := tr.CompactHash()
	
	// Compact should be prefix of full hash
	if len(hash) < len(compact) {
		t.Error("Compact hash should be shorter than full hash")
	}
	
	if hash[:len(compact)] != compact {
		t.Error("Compact hash should be prefix of full hash")
	}
}

func TestEmptyTrace(t *testing.T) {
	tr := New(10)
	
	hash := tr.Hash()
	
	// Empty trace should have empty hash
	if hash != "" {
		t.Errorf("Empty trace should have empty hash, got %s", hash)
	}
}
