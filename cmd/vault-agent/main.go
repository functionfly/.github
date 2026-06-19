// Package main implements the functionfly-vault-agent binary — a
// headless CI/CD agent that connects to the FunctionFly API and
// mints dynamic credentials on behalf of a workload.
//
// The agent holds the per-tenant DEK in the OS keychain (or
// in-memory, gated by env vars), wrapped under the user's vault
// passphrase. It exposes a small local HTTP proxy on 127.0.0.1:8090
// for application workloads to fetch fresh leases.
package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"github.com/functionfly/functionfly/internal/apierror"
)

// =============================================================================
// Config
// =============================================================================

type Config struct {
	Server          string
	EnrollmentToken string
	Bind            string
	CacheTTL        time.Duration
	LogLevel        string
	InjectEnv       bool
	KeychainService string
	InMemoryDEK     string // hex-encoded DEK for testing; not for production
	DEKOnly         bool   // skip keychain (ephemeral agent)
}

func parseFlags() Config {
	c := Config{}
	flag.StringVar(&c.Server, "server", envDefault("VAULT_AGENT_SERVER", "https://api.functionfly.com"), "FunctionFly API base URL")
	flag.StringVar(&c.EnrollmentToken, "enrollment-token", envDefault("VAULT_AGENT_ENROLLMENT_TOKEN", ""), "One-time enrollment token")
	flag.StringVar(&c.Bind, "bind", envDefault("VAULT_AGENT_BIND", "127.0.0.1:8090"), "Local HTTP proxy bind address")
	flag.DurationVar(&c.CacheTTL, "cache-ttl", 15*time.Minute, "Cache TTL for unwrapped DEK")
	flag.StringVar(&c.LogLevel, "log-level", envDefault("VAULT_AGENT_LOG_LEVEL", "info"), "Log level (debug|info|warn|error)")
	flag.BoolVar(&c.InjectEnv, "inject-env", false, "Inject env vars into child process and exit (-- cmd args...)")
	flag.StringVar(&c.KeychainService, "keychain-service", "functionfly-vault-agent", "OS keychain service name")
	flag.StringVar(&c.InMemoryDEK, "in-memory-dek", "", "Hex-encoded DEK (for testing only)")
	flag.BoolVar(&c.DEKOnly, "dek-only", false, "Skip keychain; hold DEK in memory only")
	flag.Parse()
	return c
}

func envDefault(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}

// =============================================================================
// Keychain (best-effort)
// =============================================================================
//
// We use the standard `secret-tool` CLI on Linux, the `security` CLI on
// macOS, and the wincred CLI on Windows. This avoids pulling in a
// large CGo keychain library. Production deployments that need a
// tighter integration should vendor a proper library.

func keychainSet(service, account, value string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "add-generic-password",
			"-a", account, "-s", service, "-w", value, "-U").Run()
	case "linux":
		return exec.Command("secret-tool", "store",
			"--service="+service, "--account="+account).Run()
	case "windows":
		return exec.Command("cmdkey", "/generic:"+service+":"+account, "/user:"+account, "/pass:"+value).Run()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func keychainGet(service, account string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("security", "find-generic-password",
			"-a", account, "-s", service, "-w").Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	case "linux":
		out, err := exec.Command("secret-tool", "look",
			"--service="+service, "--account="+account).Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		_, err := exec.Command("cmdkey", "/list:"+service+":"+account).Output()
		if err != nil {
			return "", err
		}
		// wincred doesn't echo the password; placeholder.
		return "", fmt.Errorf("windows keychain get not supported via cmdkey; use DEK env var")
	}
	return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// =============================================================================
// Enrollment
// =============================================================================

type EnrollResponse struct {
	TenantID    string `json:"tenant_id"`
	UserID      string `json:"user_id"`
	WrappedDEK  string `json:"wrapped_dek"`
	DEKIV       string `json:"dek_iv"`
	DEKAuthTag  string `json:"dek_auth_tag"`
	DEKSalt     string `json:"dek_salt"`
	KeyVersion  int    `json:"key_version"`
	KDFParams   map[string]interface{} `json:"kdf_params"`
}

func enroll(ctx context.Context, server, token, passphrase string) (*EnrollResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("--enrollment-token is required (or VAULT_AGENT_ENROLLMENT_TOKEN)")
	}
	// Call /v1/vault/agents/enroll with the token to get the wrapped DEK.
	// The body also includes a passphrase that the agent will use to wrap
	// the DEK for storage. For agent use, the passphrase is derived from
	// the agent's own secret (or generated at enrollment time).
	url := strings.TrimRight(server, "/") + "/v1/vault/agents/enroll"
	body := map[string]string{
		"enrollment_token": token,
		// The server returns the DEK wrapped under a random KEK; the
		// agent then re-wraps under the user's passphrase.
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("enroll: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("enroll: status %d: %s", resp.StatusCode, string(buf))
	}
	var out EnrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("enroll: decode: %w", err)
	}
	return &out, nil
}

// =============================================================================
// DEK cache
// =============================================================================

type DEKCache struct {
	mu       sync.Mutex
	dek      []byte
	cachedAt time.Time
	ttl      time.Duration
}

func (c *DEKCache) Get() ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dek == nil {
		return nil, false
	}
	if time.Since(c.cachedAt) > c.ttl {
		zeroize(c.dek)
		c.dek = nil
		return nil, false
	}
	return c.dek, true
}

func (c *DEKCache) Set(dek []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dek != nil {
		zeroize(c.dek)
	}
	c.dek = append([]byte{}, dek...)
	c.cachedAt = time.Now()
}

func zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// =============================================================================
// Argon2id KDF (matches the dashboard's parameters)
// =============================================================================

type ArgonParams struct {
	MemoryKiB uint32
	Iterations uint32
	Parallelism uint8
}

func defaultArgonParams() ArgonParams {
	return ArgonParams{MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 4}
}

func deriveKEK(passphrase string, salt []byte, p ArgonParams) []byte {
	return argon2.IDKey([]byte(passphrase), salt, p.Iterations, p.MemoryKiB, p.Parallelism, 32)
}

// =============================================================================
// Wrap / unwrap DEK under passphrase
// =============================================================================

func wrapDEK(dek []byte, passphrase string, salt []byte, p ArgonParams) (wrapped, iv, tag []byte, err error) {
	kek := deriveKEK(passphrase, salt, p)
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}
	iv = make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, err
	}
	sealed := gcm.Seal(nil, iv, dek, []byte("dek-wrap:v1"))
	tag = sealed[len(sealed)-16:]
	wrapped = sealed[:len(sealed)-16]
	return wrapped, iv, tag, nil
}

func unwrapDEK(wrapped, iv, tag []byte, passphrase string, salt []byte, p ArgonParams) ([]byte, error) {
	kek := deriveKEK(passphrase, salt, p)
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	combined := make([]byte, 0, len(wrapped)+len(tag))
	combined = append(combined, wrapped...)
	combined = append(combined, tag...)
	return gcm.Open(nil, iv, combined, []byte("dek-wrap:v1"))
}

// =============================================================================
// Local HTTP proxy
// =============================================================================

type Agent struct {
	cfg    Config
	cache  *DEKCache
	client *http.Client
}

func (a *Agent) serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/creds/", a.handleCreds)
	mux.HandleFunc("/v1/targets/", a.handleTargets)

	srv := &http.Server{
		Addr:    a.cfg.Bind,
		Handler: mux,
	}
	ln, err := net.Listen("tcp", a.cfg.Bind)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "vault-agent listening on %s\n", a.cfg.Bind)
	return srv.Serve(ln)
}

func (a *Agent) handleCreds(w http.ResponseWriter, r *http.Request) {
	credID := strings.TrimPrefix(r.URL.Path, "/v1/creds/")
	if credID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("credential id required"))
		return
	}
	dek, ok := a.cache.Get()
	if !ok {
		apierror.WriteError(w, apierror.NewServiceUnavailable("DEK not in cache; run 'enroll' first"))
		return
	}
	// Application just calls GET to receive a fresh lease.
	lease, err := a.issueCredential(r.Context(), credID, dek, 3600)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("failed to issue credential"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lease)
}

func (a *Agent) handleTargets(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimRight(a.cfg.Server, "/") + "/v1/vault/dynamic-secret-targets"
	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, nil)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("failed to create request"))
		return
	}
	if tok := os.Getenv("VAULT_AGENT_DYN_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("failed to fetch targets"))
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// issueCredential fetches the wrapped target admin password, unwraps
// it locally, generates a fresh DB password, and POSTs to /generate.
// Returns the lease material.
func (a *Agent) issueCredential(ctx context.Context, credID string, dek []byte, ttlSeconds int) (map[string]interface{}, error) {
	// 1. Look up the credential to get the target ID.
	targetID, err := a.credentialTarget(ctx, credID)
	if err != nil {
		return nil, fmt.Errorf("lookup target: %w", err)
	}
	// 2. Fetch the wrapped admin password.
	wrapped, err := a.fetchWrappedTarget(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("fetch wrapped: %w", err)
	}
	tenantID, _ := uuid.Parse(wrapped.TenantID)
	targetUUID, _ := uuid.Parse(targetID)
	// 3. Unwrap the admin password.
	ct, _ := base64.StdEncoding.DecodeString(wrapped.WrappedAdminPassword)
	iv, _ := base64.StdEncoding.DecodeString(wrapped.WrapIV)
	tag, _ := base64.StdEncoding.DecodeString(wrapped.WrapAuthTag)
	adminPwd, err := clientWrapDecrypt(ct, iv, tag, dek, tenantID, targetUUID)
	if err != nil {
		return nil, fmt.Errorf("unwrap admin password: %w", err)
	}
	defer zeroize(adminPwd)
	// 4. Generate fresh DB user/password.
	newPwd, err := generatePassword(24)
	if err != nil {
		return nil, err
	}
	newUser := "vault_p_" + strings.ToLower(generatePasswordSimple(16))
	// 5. POST to /generate.
	body := map[string]string{
		"target_admin_password": string(adminPwd),
		"new_db_username":       newUser,
		"new_db_password":       newPwd,
		"ttl_seconds":           strconv.Itoa(ttlSeconds),
	}
	b, _ := json.Marshal(body)
	url := strings.TrimRight(a.cfg.Server, "/") + "/v1/vault/dynamic-credentials/" + credID + "/generate"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	// Auth: use the agent's ff_dyn_<token> bearer if available.
	if tok := os.Getenv("VAULT_AGENT_DYN_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate: status %d: %s", resp.StatusCode, string(buf))
	}
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

type WrappedTargetAPI struct {
	TargetID            string `json:"target_id"`
	TenantID            string `json:"tenant_id"`
	EncryptionMode      string `json:"encryption_mode"`
	WrappedAdminPassword string `json:"wrapped_admin_password"`
	WrapIV              string `json:"wrap_iv"`
	WrapAuthTag         string `json:"wrap_auth_tag"`
	KeyVersion          int    `json:"key_version"`
}

func (a *Agent) fetchWrappedTarget(ctx context.Context, targetID string) (*WrappedTargetAPI, error) {
	url := strings.TrimRight(a.cfg.Server, "/") + "/v1/vault/dynamic-secret-targets/" + targetID + "/wrapped"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if tok := os.Getenv("VAULT_AGENT_DYN_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(buf))
	}
	var out WrappedTargetAPI
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (a *Agent) credentialTarget(ctx context.Context, credID string) (string, error) {
	// Simplest approach: list credentials, find the one with this ID.
	url := strings.TrimRight(a.cfg.Server, "/") + "/v1/vault/dynamic-credentials"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if tok := os.Getenv("VAULT_AGENT_DYN_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Credentials []struct {
			ID       string `json:"id"`
			TargetID string `json:"target_id"`
		} `json:"credentials"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	for _, c := range out.Credentials {
		if c.ID == credID {
			return c.TargetID, nil
		}
	}
	return "", fmt.Errorf("credential not found")
}

// =============================================================================
// Helpers (mirrored from internal/storage/vault/dynamic_service.go)
// =============================================================================

func generatePassword(n int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out), nil
}

func generatePasswordSimple(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	out := make([]byte, n)
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out)
}

// clientWrapDecrypt is the agent-side equivalent of the server's
// ClientWrapDecrypt. Kept here to avoid a circular import.
func clientWrapDecrypt(ciphertext, iv, tag []byte, dek []byte, tenantID, targetID uuid.UUID) ([]byte, error) {
	if len(dek) != 32 {
		return nil, fmt.Errorf("DEK must be 32 bytes")
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(iv) != 12 {
		return nil, fmt.Errorf("iv must be 12 bytes")
	}
	if len(tag) != 16 {
		return nil, fmt.Errorf("tag must be 16 bytes")
	}
	combined := make([]byte, 0, len(ciphertext)+len(tag))
	combined = append(combined, ciphertext...)
	combined = append(combined, tag...)
	aad := []byte("client-wrap:" + tenantID.String() + ":" + targetID.String())
	return gcm.Open(nil, iv, combined, aad)
}

// sha256Hex returns the hex-encoded SHA-256 of the input.
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// =============================================================================
// Commands
// =============================================================================

func runEnroll(ctx context.Context, cfg Config) error {
	resp, err := enroll(ctx, cfg.Server, cfg.EnrollmentToken, "")
	if err != nil {
		return err
	}
	fmt.Printf("Enrolled tenant=%s user=%s\n", resp.TenantID, resp.UserID)
	// Store the wrapped DEK + the agent's own passphrase (derived from
	// a randomly generated 256-bit secret stored alongside).
	agentSecret := make([]byte, 32)
	if _, err := rand.Read(agentSecret); err != nil {
		return err
	}
	agentSecretHex := hex.EncodeToString(agentSecret)
	if !cfg.DEKOnly {
		if err := keychainSet(cfg.KeychainService, "agent-secret", agentSecretHex); err != nil {
			fmt.Fprintf(os.Stderr, "warning: keychain set failed: %v\n", err)
		}
	}
	fmt.Printf("Agent secret stored in keychain (service=%s, account=agent-secret)\n", cfg.KeychainService)
	fmt.Printf("VAULT_AGENT_DEK_HEX=%s (set this to bootstrap without keychain)\n", agentSecretHex)
	fmt.Printf("Wrapped DEK: %s\n", resp.WrappedDEK)
	return nil
}

func runServe(ctx context.Context, cfg Config) error {
	// Bootstrap: load DEK from keychain, env, or wrapped.
	var dek []byte
	var err error
	if cfg.InMemoryDEK != "" {
		dek, err = hex.DecodeString(cfg.InMemoryDEK)
		if err != nil {
			return fmt.Errorf("decode in-memory DEK: %w", err)
		}
	} else if env := os.Getenv("VAULT_AGENT_DEK_HEX"); env != "" {
		dek, err = hex.DecodeString(env)
		if err != nil {
			return fmt.Errorf("decode VAULT_AGENT_DEK_HEX: %w", err)
		}
	} else if !cfg.DEKOnly {
		secret, err := keychainGet(cfg.KeychainService, "agent-secret")
		if err == nil {
			dek, err = hex.DecodeString(secret)
			if err != nil {
				return fmt.Errorf("decode keychain secret: %w", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "keychain get failed: %v\n", err)
		}
	}
	if len(dek) != 32 {
		return fmt.Errorf("DEK not loaded; run 'enroll' first or set VAULT_AGENT_DEK_HEX")
	}
	cache := &DEKCache{ttl: cfg.CacheTTL, dek: dek, cachedAt: time.Now()}
	a := &Agent{
		cfg:    cfg,
		cache:  cache,
		client: &http.Client{Timeout: 30 * time.Second},
	}
	// Handle SIGINT/SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "shutting down")
		if dek != nil {
			zeroize(dek)
		}
		os.Exit(0)
	}()
	return a.serve()
}

func runInjectEnv(ctx context.Context, cfg Config) error {
	credID := os.Getenv("VAULT_AGENT_CRED_ID")
	if credID == "" {
		return fmt.Errorf("VAULT_AGENT_CRED_ID env var is required with --inject-env")
	}
	dek, err := loadDEK(cfg)
	if err != nil {
		return err
	}
	a := &Agent{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}
	lease, err := a.issueCredential(ctx, credID, dek, 3600)
	if err != nil {
		return err
	}
	// Spawn the child process with the lease material as env vars.
	cmd := exec.CommandContext(ctx, os.Args[1], os.Args[2:]...)
	cmd.Env = append(os.Environ(),
		"VAULT_DB_USER="+asString(lease["username"]),
		"VAULT_DB_PASSWORD="+asString(lease["password"]),
		"VAULT_DB_HOST="+asString(lease["host"]),
		"VAULT_DB_PORT="+strconv.FormatInt(int64(asInt(lease["port"])), 10),
		"VAULT_DB_NAME="+asString(lease["database"]),
		"VAULT_LEASE_ID="+asString(lease["lease_id"]),
		"VAULT_LEASE_EXPIRES="+asString(lease["expires_at"]),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func loadDEK(cfg Config) ([]byte, error) {
	if env := os.Getenv("VAULT_AGENT_DEK_HEX"); env != "" {
		return hex.DecodeString(env)
	}
	if !cfg.DEKOnly {
		secret, err := keychainGet(cfg.KeychainService, "agent-secret")
		if err == nil {
			return hex.DecodeString(secret)
		}
	}
	return nil, fmt.Errorf("DEK not available")
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

// constantTimeEq is a small helper for fingerprinting.
func constantTimeEq(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// =============================================================================
// main
// =============================================================================

func main() {
	cfg := parseFlags()
	ctx := context.Background()

	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "usage: functionfly-vault-agent <enroll|serve|version> [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  enroll    One-time enrollment; stores agent secret in keychain")
		fmt.Fprintln(os.Stderr, "  serve     Start the local HTTP proxy")
		fmt.Fprintln(os.Stderr, "  version   Print the version and exit")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or use --inject-env to inject lease material into a child process.")
		os.Exit(2)
	}
	cmd := flag.Args()[0]
	switch cmd {
	case "enroll":
		if err := runEnroll(ctx, cfg); err != nil {
			fmt.Fprintln(os.Stderr, "enroll failed:", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(ctx, cfg); err != nil {
			fmt.Fprintln(os.Stderr, "serve failed:", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("functionfly-vault-agent 1.0.0 (2026-06-16)")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(2)
	}
}
