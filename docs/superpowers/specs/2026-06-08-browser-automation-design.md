# Browser Automation: Abstraction Layer with CDP-Native Implementation

**Date:** 2026-06-08
**Status:** Draft
**Author:** System Architect

---

## 1. Overview

Replace the external `agent-browser` CLI with a pure Go CDP (Chrome DevTools Protocol) implementation via `chromedp`. The architecture introduces a `BrowserProvider` interface that allows swapping implementations without changing handler code — enabling future support for `playwright-go` (Firefox/WebKit) behind the same interface.

**Goals:**
- Eliminate external CLI dependency and subprocess spawning overhead
- Reduce action latency from ~50-100ms to ~1-5ms
- Enable feature-rich browser automation (auto-wait, network interception)
- Design for scale (50-200+ concurrent sessions) but launch lean
- Clean break migration — no dual-write, no CLI fallback

**Non-Goals:**
- Keeping CLI as fallback (clean break only)
- Supporting Firefox/WebKit in v1 (can add via interface later)

---

## 2. Architecture

### 2.1 BrowserProvider Interface

**Location:** `internal/agent/browser/iface.go`

```go
// BrowserProvider is the interface for all browser implementations.
// Chromedp, playwright-go, and mock implementations conform to this interface.
type BrowserProvider interface {
    // Lifecycle
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error

    // Session management
    CreateSession(ctx context.Context, agentID string, opts SessionOptions) (*Session, error)
    GetSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)
    ListSessions(ctx context.Context, agentID string) ([]*Session, error)
    CloseSession(ctx context.Context, sessionID uuid.UUID) error

    // Actions
    Navigate(ctx context.Context, sessionID uuid.UUID, url string) (*NavigateResult, error)
    Click(ctx context.Context, sessionID uuid.UUID, elementRef string) error
    Fill(ctx context.Context, sessionID uuid.UUID, elementRef, value string) error
    Extract(ctx context.Context, sessionID uuid.UUID, selector string) ([]map[string]interface{}, error)
    Screenshot(ctx context.Context, sessionID uuid.UUID) (string, error)

    // Credential management
    StoreCredential(ctx context.Context, agentID, name, domain string, data *CredentialData) (*BrowserCredential, error)
    GetCredential(ctx context.Context, agentID, credentialID string) (*BrowserCredential, error)
    ListCredentials(ctx context.Context, agentID string) ([]*BrowserCredential, error)
    DeleteCredential(ctx context.Context, agentID, credentialID string) error

    // Permission management
    GetPermission(ctx context.Context, agentID string) (*BrowserPermission, error)
    UpsertPermission(ctx context.Context, perm *BrowserPermission) error

    // Pool management
    AcquireBrowser(ctx context.Context, agentID string) (BrowserHandle, error)
    ReleaseBrowser(ctx context.Context, agentID string)
    PoolStats() (available, allocated int)

    // Health and diagnostics
    HealthCheck(ctx context.Context) map[string]string
}

// SessionOptions configures a new browser session.
type SessionOptions struct {
    Isolated    bool
    TimeoutMs   int
    Headless    bool
}

// BrowserHandle represents an allocated browser instance.
type BrowserHandle interface {
    Port() int
    PID() int
    IsHealthy() bool
}
```

**Design rationale:** Interface segregation allows implementing only the methods needed per use case. Handler code depends on interface, not concrete implementation.

---

### 2.2 CDP Implementation (Chromedp)

**Location:** `internal/agent/browser/cdp/browser.go`

**Implementation:** `CDPBrowser` struct implementing `BrowserProvider`.

**Key design decisions:**

1. **Persistent WebSocket connections** — Each session maintains a persistent CDP WebSocket connection to Chrome. No subprocess spawn per action. Connection reused across actions within the same session.

2. **Chrome instance pool** — Chrome instances allocated from a configurable pool (default 10, grows to 50+). Ports 9222-9222+PoolSize-1 for shared browsers. Isolated browsers use port range 19222+.

3. **chromedp context per session** — Each `Session` wraps a `*chromedp.Ctx` with its own cancel function. The chromedp context manages the WebSocket connection lifecycle.

```go
type CDPBrowser struct {
    config     Config
    db         *gorm.DB
    redis      *redis.Client
    pool       *CDPPool
    isolated   *IsolatedManager
    sessionMgr *SessionManager
    policy     *PolicyChecker
    creds      *CredentialManager
    recovery   *RecoveryManager
}
```

**Chrome lifecycle:**
- Chrome launched with `--remote-debugging-port={port}` and `--no-sandbox`
- CDP connects via `ws://localhost:{port}/devtools/browser/{browser-id}`
- chromedp manages WebSocket reconnect on connection loss
- Chrome process monitored — restart on crash detection

---

### 2.3 Session Manager

**Location:** `internal/agent/browser/session.go` (adapted)

**Session struct:**

```go
type Session struct {
    ID           uuid.UUID
    AgentID      string
    BrowserPort  int
    BrowserPID   int
    WebSocketURL string
    AuthToken    string
    SessionType  SessionType  // shared | isolated
    Status       SessionStatus  // active | closing | closed
    CDPContext   context.Context  // chromedp context
    CDPcancel    context.CancelFunc
    CreatedAt    time.Time
    ClosedAt     *time.Time
}
```

**Redis keys:**
- `browser:session:{sessionID}` — Session state (TTL: SessionTTL)
- `browser:affinity:{agentID}` — Agent → port affinity (sticky sessions)
- `browser:loop:{agentID}:{sessionID}:{domain}` — Loop detection counter

**Sticky session behavior:**
- On `CreateSession`, check `browser:affinity:{agentID}`
- If exists and browser healthy, reuse same port
- If not, allocate from available pool

---

### 2.4 Pool Manager

**Location:** `internal/agent/browser/pool.go` (adapted)

**CDPPool struct:**

```go
type CDPPool struct {
    mu       sync.RWMutex
    config   Config
    browsers map[int]*BrowserInstance  // port -> instance
    allocated map[string]int  // agentID -> port
    available []int  // ports available for allocation
}

type BrowserInstance struct {
    Port    int
    PID     int
    URL     string  // ws://localhost:{port}/...
    Status  InstanceStatus  // starting | running | stopping | stopped
    AgentID string  // "" if available
}
```

**Pool behavior:**
- **Acquire:** Pop from `available` slice, mark as allocated, return port
- **Release:** Mark as available, clear agent affinity
- **Health check:** Every 30s, send CDP `Page.reload()` to verify Chrome responsive
- **Exhaustion handling:** Log warning when `available <= 2`, queue requests if configured

---

### 2.5 Isolated Manager

**Location:** `internal/agent/browser/isolated.go` (adapted)

Isolated browsers (Premium tier) get dedicated Chrome instances:

```go
type IsolatedManager struct {
    mu        sync.RWMutex
    instances map[string]*IsolatedInstance  // agentID -> instance
    config    Config
    sessionMgr *SessionManager
}

type IsolatedInstance struct {
    ID        uuid.UUID
    AgentID   string
    Port      int
    PID       int
    CDPContext context.Context
    CDPcancel  context.CancelFunc
    Status    string  // starting | running | stopping | stopped
    StartedAt time.Time
}
```

**Behavior:**
- One isolated instance per agent (reused if running)
- Port allocated from 19222+ range
- Background cleanup every 5 minutes — remove stopped instances older than TTL

---

### 2.6 Recovery Manager

**Location:** `internal/agent/browser/recovery.go` (enhanced)

**Error classification:**

```go
type ErrorType int
const (
    ErrorTypeUnknown ErrorType = iota
    ErrorTypeCrash   // Chrome process died
    ErrorTypeTimeout // CDP command timed out
    ErrorTypeNetwork // WebSocket connection lost
    ErrorTypeDomain  // Policy violation
)
```

**Recovery actions:**

```go
type RecoveryAction int
const (
    RecoveryActionRetry RecoveryAction = iota  // Re-execute CDP command
    RecoveryActionRestart                       // Close session, acquire new browser, retry
    RecoveryActionFallback                      // For isolated: failover to shared pool
    RecoveryActionFail                          // Log to dead letter, return error
)
```

**Circuit breaker per agent:**
- Track consecutive failures per agent
- Open after 5 failures in 60 seconds
- Half-open after 30s cooldown
- Return 503 with `Retry-After` when open

---

### 2.7 Policy Checker

**Location:** `internal/agent/browser/policy.go` (existing, adapted)

Domain whitelist enforcement per agent. No changes needed — works the same with CDP as with CLI.

---

### 2.8 Credential Manager

**Location:** `internal/agent/browser/credentials.go` (existing, adapted)

Browser credential storage (zero-knowledge vault integration). No changes needed.

---

## 3. Data Flow

```
Agent Request
     │
     ▼
┌─────────────────────────────────────────────────────┐
│  BrowserService.Navigate()                          │
│  1. Policy check (domain allowed?)                  │
│  2. Get or create session (Redis lookup)           │
│  3. Loop detection check                           │
└─────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────┐
│  CDPBrowser.Navigate(session, url)                   │
│  1. chromedp.Run(session.CDPContext,               │
│       chromedp.Navigate(url),                      │
│       chromedp.Title(&title),                      │
│       chromedp.Status(&status))                    │
└─────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────┐
│  Result → Redis usage record                       │
│  Key: browser:usage:{agent}:{date}                 │
│  Fields: total_actions, total_minutes              │
└─────────────────────────────────────────────────────┘
     │
     ▼
    NavigateResult (JSON response)
```

---

## 4. Configuration

**Environment variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `BROWSER_ENABLED` | `true` | Enable/disable browser service |
| `BROWSER_POOL_SIZE` | `10` | Number of shared Chrome instances |
| `BROWSER_SESSION_TTL` | `3600` | Session TTL in seconds |
| `BROWSER_TIMEOUT_MS` | `30000` | Default action timeout |
| `BROWSER_ISOLATED_PORT_START` | `19222` | Start of isolated port range |
| `BROWSER_HEALTH_CHECK_INTERVAL` | `30` | Health check interval in seconds |
| `BROWSER_CIRCUIT_BREAKER_THRESHOLD` | `5` | Failures before circuit opens |
| `BROWSER_CIRCUIT_BREAKER_COOLDOWN` | `30` | Cooldown in seconds |

---

## 5. Launch Phases

### Phase 1: Design (current)
- Write spec document
- Review and approve

### Phase 2: Interface + CDP Implementation
- Define `BrowserProvider` interface
- Implement `CDPBrowser` with `chromedp`
- Adapt existing managers (Session, Pool, Isolated, Policy, Credentials, Recovery)
- Write unit tests with mock implementation
- Write integration tests with real Chrome

### Phase 3: Shadow Mode Validation
- Deploy CDP implementation alongside existing CLI code
- Log all actions from both, compare results
- Fix any discrepancies

### Phase 4: Cutover
- CDP becomes primary (no CLI calls)
- Remove CLI invocation code from `executeBrowserCommand`
- CLI code remains in repo (one-line revert if catastrophic failure)

### Phase 5: Future Enhancements
- Add `PlaywrightBrowser` implementation behind same interface
- Firefox/WebKit support
- Multi-region pool management

---

## 6. File Structure

```
internal/agent/browser/
├── iface.go              # BrowserProvider interface
├── cdp/
│   └── browser.go        # CDPBrowser implementation
├── service.go            # BrowserService (orchestrator)
├── config.go             # Configuration
├── session.go            # SessionManager (Redis-backed)
├── pool.go               # CDPPool (browser instance pool)
├── isolated.go           # IsolatedManager (per-agent browsers)
├── policy.go             # PolicyChecker (domain whitelist)
├── credentials.go        # CredentialManager
├── recovery.go           # RecoveryManager + CircuitBreaker
└── browser.go            # Database models + types
```

---

## 7. Testing Strategy

**Unit tests:**
- `MockBrowserProvider` implementing `BrowserProvider` interface
- Test each method in isolation
- No Chrome required

**Integration tests:**
- Real Chrome via `chromedp.NewRunner()` with `--no-sandbox`
- Test full action flow: navigate → click → fill → extract → screenshot
- Test session lifecycle: create → use → close

**Performance tests:**
- Benchmark CDP vs CLI latency
- Verify pool correctly handles concurrent sessions

---

## 8. Dependencies

**go.mod additions:**
```go
github.com/chromedp/chromedp v0.52.0
```

**Removed:**
- `agent-browser` CLI (no longer required)

**Existing dependencies:**
- `github.com/redis/go-redis/v9` — Redis client
- `gorm.io/gorm` — Database ORM

---

## 9. Migration

**Clean break strategy:**
1. Implement CDP in parallel with CLI (Phase 2)
2. Shadow validate results match (Phase 3)
3. Cutover to CDP as primary (Phase 4)
4. CLI code retained but never called

**Revert path:** Single line change in `service.go` — revert to CLI `executeBrowserCommand` call.

---

## 10. Success Metrics

- **Latency reduction:** CDP actions < 10ms average vs CLI 50-100ms
- **Reliability:** < 0.1% action failures after circuit breaker warmup
- **Pool utilization:** Available browsers > 2 at all times during normal operation
- **Zero external dependencies:** No CLI binary required to run browser automation