package artifacts

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LocalStore is a filesystem-backed Store used for local development and
// `--skip-migrations` workflows where R2 is not configured.
//
// Layout mirrors the R2 key scheme so swapping the backend in production is a
// pure configuration change.
type LocalStore struct {
	root      string
	presignMu sync.Mutex
	presigned map[string]presignedToken
}

type presignedToken struct {
	key         string
	contentType string
	maxBytes    int64
	expiresAt   time.Time
}

// NewLocalStore creates (or reuses) a LocalStore rooted at path. Missing
// parent directories are created.
func NewLocalStore(path string) (*LocalStore, error) {
	if path == "" {
		path = filepath.Join(os.TempDir(), "functionfly-artifacts")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("artifacts: create local root %s: %w", path, err)
	}
	logrus.WithField("path", path).Info("artifacts: local store initialised")
	return &LocalStore{
		root:      path,
		presigned: map[string]presignedToken{},
	}, nil
}

// Backend implements Store.
func (s *LocalStore) Backend() Backend { return BackendLocal }

// Put implements Store.
func (s *LocalStore) Put(_ context.Context, kind Kind, body io.Reader, contentType string) (Meta, error) {
	if body == nil {
		return Meta{}, errors.New("artifacts: nil body")
	}
	hasher := sha256.New()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(io.MultiWriter(hasher, buf), body); err != nil {
		return Meta{}, fmt.Errorf("artifacts: read body: %w", err)
	}
	data := buf.Bytes()
	sha := hex.EncodeToString(hasher.Sum(nil))
	if kind == "" {
		kind = kindFromContentType(contentType)
	}
	ext := extFromContentType(contentType)
	key := KeyFor(kind, sha, ext)
	return s.commit(data, sha, key, contentType)
}

// PutForKey stores the bytes at an explicit key (no content-hash derivation).
// Used by the local-upload HTTP handler after the presign-PUT token validates.
func (s *LocalStore) PutForKey(_ context.Context, key string, body io.Reader, contentType string) (Meta, error) {
	if key == "" {
		return Meta{}, errors.New("artifacts: empty key")
	}
	if body == nil {
		return Meta{}, errors.New("artifacts: nil body")
	}
	hasher := sha256.New()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(io.MultiWriter(hasher, buf), body); err != nil {
		return Meta{}, fmt.Errorf("artifacts: read body: %w", err)
	}
	data := buf.Bytes()
	sha := hex.EncodeToString(hasher.Sum(nil))
	return s.commit(data, sha, key, contentType)
}

// commit writes the bytes to disk at key. Shared by Put (content-addressed) and
// PutForKey (token-bound).
func (s *LocalStore) commit(data []byte, sha, key, contentType string) (Meta, error) {
	full := filepath.Join(s.root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return Meta{}, fmt.Errorf("artifacts: mkdir %s: %w", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return Meta{}, fmt.Errorf("artifacts: write %s: %w", full, err)
	}
	return Meta{
		Key:         key,
		Backend:     BackendLocal,
		ContentHash: sha,
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

// Get implements Store.
func (s *LocalStore) Get(_ context.Context, key string) (io.ReadCloser, error) {
	full, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("artifacts: local object %s not found", key)
		}
		return nil, err
	}
	return f, nil
}

// Head implements Store.
func (s *LocalStore) Head(_ context.Context, key string) (Meta, error) {
	full, err := s.resolve(key)
	if err != nil {
		return Meta{}, err
	}
	fi, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return Meta{}, fmt.Errorf("artifacts: local object %s not found", key)
		}
		return Meta{}, err
	}
	return Meta{
		Key:     key,
		Backend: BackendLocal,
		Size:    fi.Size(),
	}, nil
}

// Delete implements Store.
func (s *LocalStore) Delete(_ context.Context, key string) error {
	full, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists implements Store.
func (s *LocalStore) Exists(_ context.Context, key string) (bool, error) {
	full, err := s.resolve(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(full)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// PresignPut implements Store. The local backend doesn't support real
// presigned URLs, so we mint a one-shot upload token that the local upload
// handler accepts.
func (s *LocalStore) PresignPut(_ context.Context, key, contentType string, maxBytes int64) (string, error) {
	token := randomToken()
	exp := time.Now().Add(PresignTTL)
	s.presignMu.Lock()
	s.presigned[token] = presignedToken{key: key, contentType: contentType, maxBytes: maxBytes, expiresAt: exp}
	s.presignMu.Unlock()

	v := url.Values{}
	v.Set("token", token)
	v.Set("key", key)
	return "/api/artifacts/local-upload?" + v.Encode(), nil
}

// PresignGet implements Store. The local backend serves the file directly via
// an authenticated handler so the dashboard can preview without proxying
// bytes through Go. The token is single-use and bound to the requested key.
func (s *LocalStore) PresignGet(_ context.Context, key string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = PresignTTL
	}
	token := randomToken()
	exp := time.Now().Add(ttl)
	s.presignMu.Lock()
	s.presigned[token] = presignedToken{key: key, expiresAt: exp}
	s.presignMu.Unlock()

	v := url.Values{}
	v.Set("token", token)
	v.Set("key", key)
	return "/api/artifacts/local-download?" + v.Encode(), nil
}

// LocalDownloadToken looks up a presigned-GET token without consuming it so
// the HTTP handler can validate and stream the bytes.
func (s *LocalStore) LocalDownloadToken(token string) (string, error) {
	s.presignMu.Lock()
	defer s.presignMu.Unlock()
	t, ok := s.presigned[token]
	if !ok {
		return "", errors.New("artifacts: unknown download token")
	}
	if time.Now().After(t.expiresAt) {
		delete(s.presigned, token)
		return "", errors.New("artifacts: download token expired")
	}
	return t.key, nil
}

// ConsumePresignedToken validates and removes a presigned upload token. The
// local HTTP handler uses this to commit bytes uploaded via the
// presigned-PUT simulation.
func (s *LocalStore) ConsumePresignedToken(token string) (presignedToken, error) {
	s.presignMu.Lock()
	defer s.presignMu.Unlock()
	t, ok := s.presigned[token]
	if !ok {
		return presignedToken{}, errors.New("artifacts: unknown or consumed presigned token")
	}
	delete(s.presigned, token)
	if time.Now().After(t.expiresAt) {
		return presignedToken{}, errors.New("artifacts: presigned token expired")
	}
	return t, nil
}

// PresignedTokenKey returns the target object key for a presigned upload
// token without consuming it.
func (s *LocalStore) PresignedTokenKey(token string) (string, error) {
	s.presignMu.Lock()
	defer s.presignMu.Unlock()
	t, ok := s.presigned[token]
	if !ok {
		return "", errors.New("artifacts: unknown presigned token")
	}
	if time.Now().After(t.expiresAt) {
		delete(s.presigned, token)
		return "", errors.New("artifacts: presigned token expired")
	}
	return t.key, nil
}

// PresignedTokenContentType returns the declared content type for a presigned
// upload token without consuming it.
func (s *LocalStore) PresignedTokenContentType(token string) (string, error) {
	s.presignMu.Lock()
	defer s.presignMu.Unlock()
	t, ok := s.presigned[token]
	if !ok {
		return "", errors.New("artifacts: unknown presigned token")
	}
	if time.Now().After(t.expiresAt) {
		delete(s.presigned, token)
		return "", errors.New("artifacts: presigned token expired")
	}
	return t.contentType, nil
}

// PresignedTokenMaxBytes returns the upload-size cap for a presigned upload
// token without consuming it.
func (s *LocalStore) PresignedTokenMaxBytes(token string) (int64, error) {
	s.presignMu.Lock()
	defer s.presignMu.Unlock()
	t, ok := s.presigned[token]
	if !ok {
		return 0, errors.New("artifacts: unknown presigned token")
	}
	if time.Now().After(t.expiresAt) {
		delete(s.presigned, token)
		return 0, errors.New("artifacts: presigned token expired")
	}
	return t.maxBytes, nil
}

func (s *LocalStore) resolve(key string) (string, error) {
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("artifacts: invalid key %q", key)
	}
	return filepath.Join(s.root, filepath.FromSlash(key)), nil
}

func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}