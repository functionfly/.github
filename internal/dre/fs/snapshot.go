// Package fs implements filesystem snapshot handling for the DCC Protocol.
// It provides read-only snapshot layers and ephemeral /tmp tracking.
package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"sync"
)

// FilesystemSnapshot represents a read-only filesystem snapshot for the capsule.
type FilesystemSnapshot struct {
	// Base hash of the snapshot (dependencies, compiled artifacts, static assets)
	hash string
	
	// Virtual filesystem tree
	tree map[string]*FileNode
	
	// Ephemeral /tmp tracking
	tmpFiles map[string][]byte
	
	// Ephemeral hash (included in OutputHash)
	ephemeralHash string
	
	mu sync.RWMutex
}

// FileNode represents a node in the virtual filesystem.
type FileNode struct {
	Name    string       `json:"name"`
	IsDir   bool         `json:"is_dir"`
	Mode    fs.FileMode  `json:"mode"`
	Size    int64        `json:"size"`
	Content []byte       `json:"content,omitempty"`
	Children []*FileNode `json:"children,omitempty"`
}

// New creates a new FilesystemSnapshot with the given base hash.
func New(baseHash string) *FilesystemSnapshot {
	return &FilesystemSnapshot{
		hash:      baseHash,
		tree:      make(map[string]*FileNode),
		tmpFiles:  make(map[string][]byte),
	}
}

// Hash returns the snapshot hash (FSSnapshotHash).
func (s *FilesystemSnapshot) Hash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hash
}

// SetHash sets the snapshot hash.
func (s *FilesystemSnapshot) SetHash(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash = hash
}

// AddFile adds a file to the snapshot.
func (s *FilesystemSnapshot) AddFile(filePath string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Normalize path
	filePath = path.Clean(filePath)
	
	// Create the file node
	node := &FileNode{
		Name:    path.Base(filePath),
		IsDir:   false,
		Mode:    0644,
		Size:    int64(len(content)),
		Content: content,
	}
	
	// Add to tree
	s.tree[filePath] = node
	
	return nil
}

// AddDirectory adds a directory to the snapshot.
func (s *FilesystemSnapshot) AddDirectory(dirPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Normalize path
	dirPath = path.Clean(dirPath)
	
	node := &FileNode{
		Name:    path.Base(dirPath),
		IsDir:   true,
		Mode:    0755,
		Children: make([]*FileNode, 0),
	}
	
	s.tree[dirPath] = node
	
	return nil
}

// GetFile returns the content of a file in the snapshot.
func (s *FilesystemSnapshot) GetFile(filePath string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	filePath = path.Clean(filePath)
	
	node, ok := s.tree[filePath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	
	if node.IsDir {
		return nil, fmt.Errorf("is a directory: %s", filePath)
	}
	
	return node.Content, nil
}

// ListDirectory lists the contents of a directory.
func (s *FilesystemSnapshot) ListDirectory(dirPath string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	dirPath = path.Clean(dirPath)
	
	// If root
	if dirPath == "." || dirPath == "/" {
		var files []string
		for p := range s.tree {
			files = append(files, p)
		}
		sort.Strings(files)
		return files, nil
	}
	
	node, ok := s.tree[dirPath]
	if !ok {
		return nil, fmt.Errorf("directory not found: %s", dirPath)
	}
	
	if !node.IsDir {
		return nil, fmt.Errorf("not a directory: %s", dirPath)
	}
	
	var files []string
	for p := range s.tree {
		if path.Dir(p) == dirPath {
			files = append(files, path.Base(p))
		}
	}
	
	sort.Strings(files)
	return files, nil
}

// WriteTmp writes to the ephemeral /tmp directory.
func (s *FilesystemSnapshot) WriteTmp(fileName string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	tmpPath := "/tmp/" + fileName
	s.tmpFiles[tmpPath] = content
	
	return nil
}

// ReadTmp reads from the ephemeral /tmp directory.
func (s *FilesystemSnapshot) ReadTmp(fileName string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tmpPath := "/tmp/" + fileName
	
	content, ok := s.tmpFiles[tmpPath]
	if !ok {
		return nil, fmt.Errorf("tmp file not found: %s", fileName)
	}
	
	return content, nil
}

// EphemeralHash returns the hash of all ephemeral /tmp files.
// This is included in the OutputHash.
func (s *FilesystemSnapshot) EphemeralHash() string {
	s.mu.RLock()
	
	if s.ephemeralHash != "" {
		s.mu.RUnlock()
		return s.ephemeralHash
	}
	s.mu.RUnlock()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	h := sha256.New()
	
	// Hash all tmp files
	var files []string
	for f := range s.tmpFiles {
		files = append(files, f)
	}
	sort.Strings(files)
	
	for _, f := range files {
		h.Write([]byte(f))
		h.Write(s.tmpFiles[f])
	}
	
	s.ephemeralHash = hex.EncodeToString(h.Sum(nil))
	return s.ephemeralHash
}

// FullHash returns the full filesystem hash including both snapshot and ephemeral.
func (s *FilesystemSnapshot) FullHash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	h := sha256.New()
	h.Write([]byte(s.hash))
	h.Write([]byte(s.EphemeralHash()))
	
	return hex.EncodeToString(h.Sum(nil))
}

// JSON returns the snapshot as JSON.
func (s *FilesystemSnapshot) JSON() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, err := json.MarshalIndent(s.tree, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}
	
	return string(data), nil
}

// SnapshotHash computes the snapshot hash from files.
func SnapshotHash(files map[string][]byte) string {
	h := sha256.New()
	
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	
	for _, p := range paths {
		h.Write([]byte(p))
		h.Write(files[p])
	}
	
	return hex.EncodeToString(h.Sum(nil))
}
