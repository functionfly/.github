package fs

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New("base-hash-123")
	
	if s == nil {
		t.Fatal("New() returned nil")
	}
	
	if s.Hash() != "base-hash-123" {
		t.Errorf("Expected hash 'base-hash-123', got '%s'", s.Hash())
	}
}

func TestAddFile(t *testing.T) {
	s := New("")
	
	err := s.AddFile("/app/main.py", []byte("print('hello')"))
	if err != nil {
		t.Errorf("AddFile() returned error: %v", err)
	}
	
	content, err := s.GetFile("/app/main.py")
	if err != nil {
		t.Errorf("GetFile() returned error: %v", err)
	}
	
	if string(content) != "print('hello')" {
		t.Errorf("Expected content 'print(hello)', got '%s'", string(content))
	}
}

func TestGetFileNotFound(t *testing.T) {
	s := New("")
	
	_, err := s.GetFile("/nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestAddDirectory(t *testing.T) {
	s := New("")
	
	err := s.AddDirectory("/app")
	if err != nil {
		t.Errorf("AddDirectory() returned error: %v", err)
	}
	
	files, err := s.ListDirectory("/app")
	if err != nil {
		t.Errorf("ListDirectory() returned error: %v", err)
	}
	
	if len(files) != 0 {
		t.Errorf("Expected empty directory, got %d files", len(files))
	}
}

func TestListDirectory(t *testing.T) {
	s := New("")
	
	s.AddFile("/app/main.py", []byte("code"))
	s.AddFile("/app/utils.py", []byte("code"))
	s.AddFile("/app/data.json", []byte("{}"))
	
	files, err := s.ListDirectory("/app")
	if err != nil {
		t.Errorf("ListDirectory() returned error: %v", err)
	}
	
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
	
	// Should be sorted
	if files[0] != "data.json" || files[1] != "main.py" || files[2] != "utils.py" {
		t.Errorf("Files not sorted: %v", files)
	}
}

func TestWriteTmp(t *testing.T) {
	s := New("")
	
	err := s.WriteTmp("temp.txt", []byte("temporary data"))
	if err != nil {
		t.Errorf("WriteTmp() returned error: %v", err)
	}
	
	content, err := s.ReadTmp("temp.txt")
	if err != nil {
		t.Errorf("ReadTmp() returned error: %v", err)
	}
	
	if string(content) != "temporary data" {
		t.Errorf("Expected content 'temporary data', got '%s'", string(content))
	}
}

func TestReadTmpNotFound(t *testing.T) {
	s := New("")
	
	_, err := s.ReadTmp("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tmp file")
	}
}

func TestEphemeralHash(t *testing.T) {
	s := New("")
	
	s.WriteTmp("file1.txt", []byte("content1"))
	s.WriteTmp("file2.txt", []byte("content2"))
	
	hash1 := s.EphemeralHash()
	
	// Should be consistent
	hash2 := s.EphemeralHash()
	
	if hash1 != hash2 {
		t.Errorf("EphemeralHash() should be consistent: %s vs %s", hash1, hash2)
	}
	
	if hash1 == "" {
		t.Error("EphemeralHash() should not be empty")
	}
}

func TestDeterministicEphemeralHash(t *testing.T) {
	s1 := New("")
	s1.WriteTmp("file.txt", []byte("content"))
	hash1 := s1.EphemeralHash()
	
	s2 := New("")
	s2.WriteTmp("file.txt", []byte("content"))
	hash2 := s2.EphemeralHash()
	
	// Hashes should be identical
	if hash1 != hash2 {
		t.Errorf("EphemeralHash() should be deterministic: %s vs %s", hash1, hash2)
	}
}

func TestDifferentTmpDifferentHash(t *testing.T) {
	s1 := New("")
	s1.WriteTmp("file.txt", []byte("content1"))
	hash1 := s1.EphemeralHash()
	
	s2 := New("")
	s2.WriteTmp("file.txt", []byte("content2"))
	hash2 := s2.EphemeralHash()
	
	// Hashes should be different
	if hash1 == hash2 {
		t.Error("EphemeralHash() should differ with different content")
	}
}

func TestFullHash(t *testing.T) {
	s := New("base-hash")
	s.WriteTmp("temp.txt", []byte("temp"))
	
	hash := s.FullHash()
	
	if hash == "" {
		t.Error("FullHash() should not be empty")
	}
	
	// Full hash should include both base and ephemeral
	if hash == s.Hash() {
		t.Error("FullHash() should differ from base hash")
	}
}

func TestSnapshotHash(t *testing.T) {
	files := map[string][]byte{
		"/app/main.py": []byte("print('hello')"),
		"/app/utils.py": []byte("def helper(): pass"),
	}
	
	hash := SnapshotHash(files)
	
	if hash == "" {
		t.Error("SnapshotHash() should not be empty")
	}
}

func TestSnapshotHashDeterministic(t *testing.T) {
	files1 := map[string][]byte{
		"/app/main.py": []byte("code"),
	}
	
	files2 := map[string][]byte{
		"/app/main.py": []byte("code"),
	}
	
	hash1 := SnapshotHash(files1)
	hash2 := SnapshotHash(files2)
	
	if hash1 != hash2 {
		t.Errorf("SnapshotHash() should be deterministic: %s vs %s", hash1, hash2)
	}
}
