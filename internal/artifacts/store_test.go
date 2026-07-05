package artifacts

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeyFor(t *testing.T) {
	got := KeyFor(KindWASM, "abcdef0123456789", "wasm")
	want := "wasm/ab/abcdef0123456789.wasm"
	if got != want {
		t.Fatalf("KeyFor wasm: got %q want %q", got, want)
	}
	got = KeyFor(KindSource, "0123456789ABCDEF", "py")
	want = "source/01/0123456789abcdef.py"
	if got != want {
		t.Fatalf("KeyFor source: got %q want %q", got, want)
	}
	// Extension-less case.
	got = KeyFor(KindReadme, "deadbeef", "")
	want = "readme/de/deadbeef"
	if got != want {
		t.Fatalf("KeyFor no ext: got %q want %q", got, want)
	}
}

func TestContentHash(t *testing.T) {
	got := ContentHash([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("ContentHash hello: got %q want %q", got, want)
	}
}

func TestLocalStorePutGet(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir)
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	ctx := context.Background()

	data := []byte("hello world")
	hash := ContentHash(data)
	meta, err := store.Put(ctx, KindSource, bytes.NewReader(data), "text/plain")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if meta.ContentHash != hash {
		t.Fatalf("hash mismatch: %s vs %s", meta.ContentHash, hash)
	}
	if meta.Backend != BackendLocal {
		t.Fatalf("backend=%s", meta.Backend)
	}

	// Get roundtrips.
	rc, err := store.Get(ctx, meta.Key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if !bytes.Equal(got, data) {
		t.Fatalf("Get bytes mismatch")
	}

	// Exists.
	exists, err := store.Exists(ctx, meta.Key)
	if err != nil || !exists {
		t.Fatalf("Exists: %v %v", exists, err)
	}

	// Head.
	h, err := store.Head(ctx, meta.Key)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if h.Size != int64(len(data)) {
		t.Fatalf("head size=%d", h.Size)
	}

	// Delete.
	if err := store.Delete(ctx, meta.Key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if exists, _ := store.Exists(ctx, meta.Key); exists {
		t.Fatalf("Exists after Delete: true")
	}
}

func TestLocalStorePresignPutAndDownload(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir)
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	ctx := context.Background()

	url, err := store.PresignPut(ctx, "wasm/ab/abcdef0123456789.wasm", "application/wasm", 1024)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if !strings.Contains(url, "token=") {
		t.Fatalf("url missing token: %s", url)
	}
	tok := tokenFromURL(t, url)
	key := "wasm/ab/abcdef0123456789.wasm"
	if k, err := store.PresignedTokenKey(tok); err != nil || k != key {
		t.Fatalf("token key=%s err=%v", k, err)
	}
	if ct, err := store.PresignedTokenContentType(tok); err != nil || ct != "application/wasm" {
		t.Fatalf("token contentType=%s err=%v", ct, err)
	}
	// Using the same token twice should fail (single-use for PUT).
	if _, err := store.ConsumePresignedToken(tok); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := store.ConsumePresignedToken(tok); err == nil {
		t.Fatalf("expected error on second consume")
	}

	// Presign GET roundtrip.
	data := []byte("compiled wasm bytes")
	full := filepath.Join(dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir seed: %v", err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	dlURL, err := store.PresignGet(ctx, key, 0)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	dlTok := tokenFromURL(t, dlURL)
	gotKey, err := store.LocalDownloadToken(dlTok)
	if err != nil {
		t.Fatalf("LocalDownloadToken: %v", err)
	}
	if gotKey != key {
		t.Fatalf("download key mismatch")
	}
}

func TestDiskLRUPutGetEvict(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewDiskLRU(dir, 10) // 10 byte cap
	if err != nil {
		t.Fatalf("NewDiskLRU: %v", err)
	}
	if err := cache.Put("a/1", bytes.NewReader([]byte("12345"))); err != nil {
		t.Fatalf("put a: %v", err)
	}
	if err := cache.Put("b/2", bytes.NewReader([]byte("67890"))); err != nil {
		t.Fatalf("put b: %v", err)
	}
	// Now total is 10; pushing another byte should evict.
	if err := cache.Put("c/3", bytes.NewReader([]byte("!"))); err != nil {
		t.Fatalf("put c: %v", err)
	}
	// a should be gone.
	if _, err := cache.Get("a/1"); err != ErrMiss {
		t.Fatalf("a should have been evicted: %v", err)
	}
	// b and c should still be present.
	for _, k := range []string{"b/2", "c/3"} {
		if rc, err := cache.Get(k); err != nil {
			t.Fatalf("%s: %v", k, err)
		} else {
			_ = rc.Close()
		}
	}
}

func tokenFromURL(t *testing.T, u string) string {
	t.Helper()
	idx := strings.Index(u, "token=")
	if idx < 0 {
		t.Fatalf("no token in url: %s", u)
	}
	rest := u[idx+len("token="):]
	if amp := strings.Index(rest, "&"); amp >= 0 {
		return rest[:amp]
	}
	return rest
}
