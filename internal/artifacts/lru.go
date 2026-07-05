package artifacts

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DiskLRU is a small content-addressed on-disk LRU cache. Keys are R2 object
// keys; values are the raw bytes. Used as L1 between the execution engines and
// the object store so repeat fetches don't pay R2 latency.
type DiskLRU struct {
	mu       sync.Mutex
	root     string
	capacity int64
	curSize  int64
	ll       *list.List              // most-recent at front
	index    map[string]*list.Element // key -> list element
}

type lruEntry struct {
	key  string
	size int64
}

// NewDiskLRU creates a disk LRU rooted at dir with a max byte size. dir is
// created if missing.
func NewDiskLRU(dir string, capacity int64) (*DiskLRU, error) {
	if capacity <= 0 {
		capacity = 512 * 1024 * 1024 // 512 MiB default
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("artifacts: mkdir %s: %w", dir, err)
	}
	c := &DiskLRU{
		root:     dir,
		capacity: capacity,
		ll:       list.New(),
		index:    map[string]*list.Element{},
	}
	return c, nil
}

// Get returns a ReadCloser for the cached value, or (nil, ErrMiss) if the key
// is not present. The returned ReadCloser is safe to close.
func (c *DiskLRU) Get(key string) (io.ReadCloser, error) {
	if !safeKey(key) {
		return nil, errors.New("artifacts: invalid cache key")
	}
	c.mu.Lock()
	el, ok := c.index[key]
	if !ok {
		c.mu.Unlock()
		return nil, ErrMiss
	}
	c.ll.MoveToFront(el)
	c.mu.Unlock()

	f, err := os.Open(c.path(key))
	if err != nil {
		c.mu.Lock()
		// File went missing under us; drop the entry.
		if el, ok := c.index[key]; ok {
			c.ll.Remove(el)
			delete(c.index, key)
		}
		c.mu.Unlock()
		if os.IsNotExist(err) {
			return nil, ErrMiss
		}
		return nil, err
	}
	return f, nil
}

// Put stores value at key, evicting older entries until total size <= capacity.
// If the key already exists it is replaced.
func (c *DiskLRU) Put(key string, value io.Reader) error {
	if !safeKey(key) {
		return errors.New("artifacts: invalid cache key")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove any existing entry first; size accounting will be redone.
	if el, ok := c.index[key]; ok {
		c.ll.Remove(el)
		delete(c.index, key)
		old := el.Value.(*lruEntry).size
		c.curSize -= old
		_ = os.Remove(c.path(key))
	}

	full := c.path(key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(full), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	n, err := io.Copy(tmp, value)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, full); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	c.curSize += n
	el := c.ll.PushFront(&lruEntry{key: key, size: n})
	c.index[key] = el
	c.evictUntil()
	return nil
}

// ErrMiss is returned by Get when the key is not cached.
var ErrMiss = errors.New("artifacts: cache miss")

func (c *DiskLRU) path(key string) string {
	return filepath.Join(c.root, filepath.FromSlash(key))
}

func (c *DiskLRU) evictUntil() {
	for c.curSize > c.capacity {
		back := c.ll.Back()
		if back == nil {
			return
		}
		e := back.Value.(*lruEntry)
		c.ll.Remove(back)
		delete(c.index, e.key)
		c.curSize -= e.size
		_ = os.Remove(c.path(e.key))
	}
}

func safeKey(k string) bool {
	if k == "" {
		return false
	}
	if strings.Contains(k, "..") {
		return false
	}
	return true
}