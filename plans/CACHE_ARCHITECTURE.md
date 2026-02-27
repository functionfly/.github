# FunctionFly Execution Caching System Architecture

> **TL;DR**: This document specifies a 3-layer caching system that transforms FunctionFly from a serverless platform into a **global behavior CDN**. The key insight: `(function version + normalized input) → output` becomes a mathematical lookup table.

---

## 1. System Overview

### 1.1 The Problem

Every function call on a tiny server consumes CPU. Without caching:
- 10,000 calls to `slugify("hello world")` = 10,000 code executions
- Your $5 server dies instantly under traffic

### 1.2 The Solution

Cache **function results as mathematical mappings**:

```
(function_id + version + normalized_input) → output
```

A function becomes a lookup table, not compute. This document specifies how.

### 1.3 Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Registry as Source of Truth** | Cache eligibility determined at publish time, not runtime guesswork |
| **Immutable Cache Entries** | Never mutate; always replace. Enables safe concurrent reads |
| **Layered Architecture** | Progressive fallback: memory (μs) → disk (ms) → CDN (free) |
| **Canonical Hashing** | Normalized inputs ensure cache hits even with JSON key reordering |
| **Defense in Depth** | Cache poisoning protection at entry and exit |

---

## 2. Cache Eligibility (Registry Integration)

### 2.1 Decision Flow

```
Request arrives → Check manifest → Eligibility decision
```

### 2.2 Manifest Fields (Already Existing)

The registry already stores these on [`RegistryFunctionVersion`](internal/storage/registry_models.go:33):

```go
type RegistryFunctionVersion struct {
    Deterministic bool   // Required: function must be deterministic
    CacheTTL      int    // Required: > 0 for caching
    Version       string // Required: cache namespace
    SideEffects   string // Required: must be "none"
    Idempotent    bool   // Optional: enables safe retry
}
```

### 2.3 Eligibility Rules

| Condition | Cache Eligible | Reason |
|-----------|---------------|--------|
| `Deterministic == true` AND `CacheTTL > 0` AND `SideEffects == "none"` | ✅ Yes | Safe to cache |
| `Deterministic == false` | ❌ No | Non-deterministic output |
| `CacheTTL == 0` | ❌ No | Explicitly disabled |
| `SideEffects == "network"` | ❌ No | External state may change |
| `SideEffects == "external_state"` | ❌ No | Files/DB may change |

### 2.4 Implementation: Cache Eligibility Checker

Location: `internal/cache/eligibility.go`

```go
package cache

import "github.com/functionfly/functionfly/internal/storage"

// EligibilityResult contains cache eligibility decision
type EligibilityResult struct {
    Eligible      bool
    TTL           int           // seconds
    Version       string        // cache namespace
    FunctionID    string        // for cache key
    CanUseCDN     bool          // public functions only
}

// CheckEligibility determines if a function version is cache-eligible
func CheckEligibility(v *storage.RegistryFunctionVersion) EligibilityResult {
    // Must be deterministic with TTL > 0
    if !v.Deterministic || v.CacheTTL <= 0 {
        return EligibilityResult{Eligible: false}
    }
    
    // Must have no side effects
    if v.SideEffects != "" && v.SideEffects != "none" {
        return EligibilityResult{Eligible: false}
    }
    
    return EligibilityResult{
        Eligible:  true,
        TTL:       v.CacheTTL,
        Version:   v.Version,
        FunctionID: v.FunctionID.String(),
        CanUseCDN: true, // Set based on function visibility
    }
}
```

---

## 3. Canonical Input Hashing

### 3.1 The Problem

These inputs should produce identical cache keys:

```json
{ "b": 2, "a": 1 }
{ "a": 1, "b": 2 }
{ "a":1,"b":2 }
```

Without normalization, cache hit rate collapses.

### 3.2 Normalization Rules

| Type | Rule |
|------|------|
| **Object keys** | Sort alphabetically (deep sort) |
| **Strings** | Trim whitespace, preserve Unicode |
| **Numbers** | Canonical JSON representation (no trailing zeros) |
| **Booleans** | Strict `true`/`false` (no "1"/"0") |
| **Nulls** | Preserve as `null` |
| **Arrays** | Process each element recursively |
| **Whitespace** | Trim, collapse interior whitespace to single space |

### 3.3 Cache Key Format

```
fx:cache:{function_id}:{version}:{input_hash}
```

Example:
```
fx:cache:a1b2c3d4:v1.0.0:9f31ab2e...
```

### 3.4 Hash Algorithm

- **Function**: SHA-256 (fast, secure, widely available)
- **Key components**: `function_id + "::" + version + "::" + normalized_input`

### 3.5 Implementation: Input Normalizer

Location: `internal/cache/normalizer.go`

```go
package cache

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "sort"
    "strings"
)

// NormalizeInput canonicalizes JSON input for consistent hashing
func NormalizeInput(input []byte) ([]byte, error) {
    // First pass: unmarshal to interface{}
    var raw interface{}
    if err := json.Unmarshal(input, &raw); err != nil {
        return nil, err
    }
    
    // Recursively normalize
    normalized := normalizeValue(raw)
    
    // Marshal with sorted keys
    return json.Marshal(normalized)
}

// normalizeValue recursively normalizes a JSON value
func normalizeValue(v interface{}) interface{} {
    switch val := v.(type) {
    case map[string]interface{}:
        // Sort keys alphabetically
        sorted := make(map[string]interface{})
        keys := make([]string, 0, len(val))
        for k := range val {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for _, k := range keys {
            sorted[k] = normalizeValue(val[k])
        }
        return sorted
        
    case []interface{}:
        // Normalize each array element
        result := make([]interface{}, len(val))
        for i, item := range val {
            result[i] = normalizeValue(item)
        }
        return result
        
    case string:
        // Trim whitespace, collapse interior spaces
        trimmed := strings.TrimSpace(val)
        return strings.Join(strings.Fields(trimmed), " ")
        
    case float64:
        // Canonical number representation
        // JSON unmarshals all numbers as float64
        // This preserves canonical form
        return val
        
    case bool:
        return val  // Already canonical
        
    case nil:
        return nil
        
    default:
        return v
    }
}

// GenerateCacheKey creates a cache key from function metadata and input
func GenerateCacheKey(functionID, version string, normalizedInput []byte) string {
    hasher := sha256.New()
    hasher.Write([]byte(functionID))
    hasher.Write([]byte("::"))
    hasher.Write([]byte(version))
    hasher.Write([]byte("::"))
    hasher.Write(normalizedInput)
    
    hash := hex.EncodeToString(hasher.Sum(nil))
    return "fx:cache:" + functionID + ":" + version + ":" + hash[:16]
}
```

---

## 4. Multi-Layer Cache Architecture

### 4.1 Layer Overview

| Layer | Storage | Latency | Capacity | Persistence |
|-------|---------|---------|----------|-------------|
| **L1 Memory** | In-process LRU | ~1-10μs | 50-500MB | ❌ Restart clears |
| **L2 Disk** | SQLite | ~1-10ms | 1-50GB | ✅ Survives restart |
| **L3 CDN** | HTTP Headers | ~50-200ms | Unlimited | ✅ Global edge |

### 4.2 L1: In-Memory Cache

**Purpose**: Handle hot traffic bursts, sub-microsecond lookups

**Implementation Options**:

| Library | Pros | Cons |
|---------|------|------|
| [groupcache](https://github.com/golang/groupcache) | Google's standard, LRU | No pure LRU |
| [bigcache](https://github.com/allegro/bigcache) | High performance, shards | Memory overhead |
| [freecache](https://github.com/coocood/freecache) | Pure Go, zero GC | No TTL by default |
| [ristretto](https://github.com/dgraph-io/ristretto) | Best Go LRU, TTL, metrics | Newer, less battle-tested |

**Recommended**: `ristretto` - best combination of features, performance, and active maintenance.

**Configuration**:

```go
// internal/cache/memory.go
package cache

import (
    "github.com/dgraph-io/ristretto"
    "time"
)

type MemoryCache struct {
    cache *ristretto.Cache
}

func NewMemoryCache(maxMemoryMB int64, metricsEnabled bool) (*MemoryCache, error) {
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: maxMemoryMB * 10,  // 10x for evictions
        MaxCost:     maxMemoryMB * 1024 * 1024,
        BufferItems: 64,
        Metrics:     metricsEnabled,
    })
    if err != nil {
        return nil, err
    }
    
    return &MemoryCache{cache: cache}, nil
}

// Get retrieves from memory cache
func (m *MemoryCache) Get(key string) ([]byte, bool) {
    return m.cache.Get(key)
}

// Set stores in memory cache
func (m *MemoryCache) Set(key string, value []byte, ttl time.Duration) {
    cost := int64(len(value))
    m.cache.SetWithTTL(key, value, cost, ttl)
}
```

**L1 Configuration Guidelines**:

| Server Tier | Max Memory | Recommended L1 Size |
|-------------|------------|---------------------|
| $5/tiny | 512MB | 50MB |
| $10/small | 1GB | 100MB |
| $20/medium | 2GB | 250MB |
| $50/large | 4GB | 500MB |

### 4.3 L2: Disk Cache (SQLite)

**Purpose**: Persist across restarts, handle repeated calls

**Schema Design**:

Location: `migrations/` (new file)

```sql
-- migrations/YYYYMMDD_cache_schema.sql

CREATE TABLE IF NOT EXISTS function_cache (
    -- Primary key components
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key       VARCHAR(255) NOT NULL UNIQUE,
    
    -- Cache key components (for invalidation)
    function_id     UUID NOT NULL,
    version         VARCHAR(50) NOT NULL,
    input_hash      VARCHAR(64) NOT NULL,
    
    -- Cached data
    output_json     JSONB NOT NULL,
    output_size     INTEGER NOT NULL,
    
    -- Metadata
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    hit_count       INTEGER NOT NULL DEFAULT 0,
    last_hit_at     TIMESTAMPTZ,
    
    -- Indexes
    CONSTRAINT unique_function_version_input UNIQUE (function_id, version, input_hash)
);

-- Index for fast lookups
CREATE INDEX idx_cache_key ON function_cache(cache_key);
CREATE INDEX idx_function_version ON function_cache(function_id, version);
CREATE INDEX idx_expires_at ON function_cache(expires_at);

-- Index for invalidation queries
CREATE INDEX idx_function_version_expires ON function_cache(function_id, version, expires_at);
```

**Implementation**:

Location: `internal/cache/disk.go`

```go
package cache

import (
    "database/sql"
    "encoding/json"
    "time"
    
    "github.com/google/uuid"
)

type DiskCache struct {
    db *sql.DB
}

type CacheRecord struct {
    ID          uuid.UUID
    CacheKey    string
    FunctionID  uuid.UUID
    Version     string
    InputHash   string
    OutputJSON  json.RawMessage
    OutputSize  int
    CreatedAt   time.Time
    ExpiresAt   time.Time
    HitCount    int
    LastHitAt   *time.Time
}

func NewDiskCache(db *sql.DB) *DiskCache {
    return &DiskCache{db: db}
}

// Get retrieves from disk cache
func (d *DiskCache) Get(cacheKey string) (*CacheRecord, error) {
    query := `
        SELECT id, cache_key, function_id, version, input_hash, 
               output_json, output_size, created_at, expires_at, 
               hit_count, last_hit_at
        FROM function_cache
        WHERE cache_key = $1 AND expires_at > NOW()
    `
    
    var record CacheRecord
    err := d.db.QueryRow(query, cacheKey).Scan(
        &record.ID, &record.CacheKey, &record.FunctionID, &record.Version,
        &record.InputHash, &record.OutputJSON, &record.OutputSize,
        &record.CreatedAt, &record.ExpiresAt, &record.HitCount, &record.LastHitAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    // Update hit count (async to avoid blocking)
    go d.incrementHitCount(record.ID)
    
    return &record, nil
}

// Set stores in disk cache
func (d *DiskCache) Set(record *CacheRecord) error {
    query := `
        INSERT INTO function_cache (
            cache_key, function_id, version, input_hash,
            output_json, output_size, expires_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (cache_key) DO UPDATE SET
            output_json = EXCLUDED.output_json,
            output_size = EXCLUDED.output_size,
            expires_at = EXCLUDED.expires_at,
            hit_count = function_cache.hit_count + 1,
            last_hit_at = NOW()
    `
    
    _, err := d.db.Exec(query,
        record.CacheKey, record.FunctionID, record.Version,
        record.InputHash, record.OutputJSON, record.OutputSize,
        record.ExpiresAt,
    )
    
    return err
}

func (d *DiskCache) incrementHitCount(id uuid.UUID) {
    // Async update - don't block
    d.db.Exec("UPDATE function_cache SET hit_count = hit_count + 1, last_hit_at = NOW() WHERE id = $1", id)
}

// Cleanup expired entries (run periodically)
func (d *DiskCache) Cleanup() error {
    _, err := d.db.Exec("DELETE FROM function_cache WHERE expires_at < NOW()")
    return err
}
```

### 4.4 L3: CDN Cache (HTTP Headers)

**Purpose**: Free compute for public deterministic functions at edge

**Strategy**: Set `Cache-Control` headers on responses for cache-eligible functions

**Cache-Control Header Values**:

| Scenario | Header | Rationale |
|----------|--------|-----------|
| Public deterministic function | `public, max-age={CacheTTL}` | CDN can cache |
| Private function | `private, max-age={CacheTTL}` | Only browser/CDN with auth can cache |
| Non-deterministic | `no-store, no-cache` | Never cache |
| Error response | `no-store` | Don't cache errors |

**Implementation**:

Location: `internal/cache/cdn.go`

```go
package cache

import (
    "net/http"
    "strconv"
)

type CDNConfig struct {
    EnableCDNCaching bool   // Toggle for CDN caching
    CDNMaxAge        int    // Default CDN max-age (seconds)
}

func SetCDNHeaders(w http.ResponseWriter, eligibility EligibilityResult, isPublic bool) {
    if !eligibility.Eligible {
        // Never cache non-eligible responses
        w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        return
    }
    
    if !isPublic || !eligibility.CanUseCDN {
        // Private cache - can be cached by browser, not CDN
        w.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(eligibility.TTL))
        return
    }
    
    // Public CDN cache
    w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(eligibility.TTL))
    w.Header().Set("X-Cache-Status", "MISS")  // Will be "HIT" at edge
}
```

---

## 5. Execution Flow (Complete)

### 5.1 Request Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           REQUEST ARRIVES                                    │
│                    POST /trase/{author}/{name}                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. LOOKUP FUNCTION & VERSION                                               │
│    • Get function from registry                                             │
│    • Get version (or latest)                                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. CHECK CACHE ELIGIBILITY                                                 │
│    • deterministic?                                                          │
│    • cache_ttl > 0?                                                         │
│    • side_effects == "none"?                                               │
│    IF NOT ELIGIBLE → Skip to execution                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                         ┌──────────┴──────────┐
                         ▼                      ▼
                 ┌─────────────┐        ┌──────────────┐
                 │   ELIGIBLE  │        │ NOT ELIGIBLE │
                 └─────────────┘        └──────────────┘
                         │                      │
                         ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. CANONICALIZE INPUT                                                      │
│    • Normalize JSON (sort keys, trim whitespace)                           │
│    • Generate SHA-256 hash                                                 │
│    • Build cache key: fx:cache:{fn_id}:{ver}:{hash}                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 4. CHECK L1 CACHE (Memory)                                                 │
│    • In-process LRU lookup                                                  │
│    • IF HIT → Return cached result                                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                         ┌──────────┴──────────┐
                         ▼                      ▼
                   ┌─────────┐           ┌──────────┐
                   │  HIT    │           │   MISS   │
                   └─────────┘           └──────────┘
                         │                      │
                         ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 5. CHECK L2 CACHE (Disk)                                                   │
│    • SQLite lookup                                                          │
│    • IF HIT → Return + populate L1                                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                         ┌──────────┴──────────┐
                         ▼                      ▼
                   ┌─────────┐           ┌──────────┐
                   │  HIT    │           │   MISS   │
                   └─────────┘           └──────────┘
                         │                      │
                         ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 6. EXECUTE FUNCTION                                                        │
│    • Run in sandbox                                                        │
│    • Record execution                                                      │
│    • Store result in L1 + L2                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 7. RETURN RESPONSE                                                          │
│    • Set Cache-Control headers                                             │
│    • Include cached=true/false in response                                │
│    • Broadcast execution (real-time)                                      │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Implementation: Cache Service

Location: `internal/cache/service.go`

```go
package cache

import (
    "database/sql"
    "encoding/json"
    "errors"
    "time"
    
    "github.com/dgraph-io/ristretto"
    "github.com/functionfly/functionfly/internal/storage"
)

var (
    ErrNotEligible = errors.New("function not eligible for caching")
    ErrNotFound    = errors.New("cache entry not found")
)

type CacheService struct {
    memory *MemoryCache
    disk   *DiskCache
    config *CacheConfig
}

type CacheConfig struct {
    MaxMemoryMB      int64
    EnableDiskCache  bool
    EnableCDNCaching bool
    DefaultTTL       int  // seconds
}

type CacheResult struct {
    Output    json.RawMessage
    FromCache bool
    Layer     string  // "memory", "disk", "none"
    Hit       bool
}

func NewCacheService(db *sql.DB, config *CacheConfig) (*CacheService, error) {
    memory, err := NewMemoryCache(config.MaxMemoryMB, true)
    if err != nil {
        return nil, err
    }
    
    disk := NewDiskCache(db)
    
    return &CacheService{
        memory: memory,
        disk:   disk,
        config: config,
    }, nil
}

// GetOrExecute checks cache, executes if needed
func (c *CacheService) GetOrExecute(
    eligibility EligibilityResult,
    input []byte,
    executeFn func() (json.RawMessage, error),
) (*CacheResult, error) {
    if !eligibility.Eligible {
        // Not eligible - just execute
        output, err := executeFn()
        if err != nil {
            return nil, err
        }
        return &CacheResult{
            Output:    output,
            FromCache: false,
            Layer:     "none",
            Hit:       false,
        }, nil
    }
    
    // Normalize input
    normalizedInput, err := NormalizeInput(input)
    if err != nil {
        return nil, err
    }
    
    // Generate cache key
    cacheKey := GenerateCacheKey(eligibility.FunctionID, eligibility.Version, normalizedInput)
    
    // Check L1
    if cached, found := c.memory.Get(cacheKey); found {
        return &CacheResult{
            Output:    cached,
            FromCache: true,
            Layer:     "memory",
            Hit:       true,
        }, nil
    }
    
    // Check L2
    if c.config.EnableDiskCache {
        record, err := c.disk.Get(cacheKey)
        if err == nil && record != nil {
            // Populate L1
            ttl := time.Duration(eligibility.TTL) * time.Second
            c.memory.Set(cacheKey, record.OutputJSON, ttl)
            
            return &CacheResult{
                Output:    record.OutputJSON,
                FromCache: true,
                Layer:     "disk",
                Hit:       true,
            }, nil
        }
    }
    
    // Execute function
    output, err := executeFn()
    if err != nil {
        return nil, err
    }
    
    // Store in cache (both layers)
    ttl := time.Duration(eligibility.TTL) * time.Second
    c.memory.Set(cacheKey, output, ttl)
    
    if c.config.EnableDiskCache {
        record := &CacheRecord{
            CacheKey:   cacheKey,
            FunctionID: eligibility.FunctionID,
            Version:    eligibility.Version,
            OutputJSON: output,
            OutputSize: len(output),
            ExpiresAt:  time.Now().Add(ttl),
        }
        c.disk.Set(record)
    }
    
    return &CacheResult{
        Output:    output,
        FromCache: false,
        Layer:     "none",
        Hit:       false,
    }, nil
}

// InvalidateFunction invalidates all cache entries for a function
func (c *CacheService) InvalidateFunction(functionID string) error {
    // For memory: we'd need to track keys per function
    // This is a simplification - in production, track keys
    
    // For disk: delete by function_id
    _, err := c.disk.db.Exec(
        "DELETE FROM function_cache WHERE function_id = $1",
        functionID,
    )
    return err
}

// InvalidateVersion invalidates cache for a specific version
func (c *CacheService) InvalidateVersion(functionID, version string) error {
    _, err := c.disk.db.Exec(
        "DELETE FROM function_cache WHERE function_id = $1 AND version = $2",
        functionID, version,
    )
    return err
}
```

---

## 6. Cache Invalidation Rules

### 6.1 Automatic Invalidation Events

| Event | Effect | Rationale |
|-------|--------|-----------|
| **New version published** | Clear cache for that function+version | Different code = different outputs |
| **TTL expired** | Entry removed on next access | Fresh data required |
| **Function set non-deterministic** | Skip cache for future requests | Can't trust old results |
| **CacheTTL set to 0** | Skip cache for future requests | Explicitly disabled |
| **SideEffects changed** | Clear cache | May affect behavior |
| **Manual purge** | Clear specific/all cache | Admin operation |

### 6.2 Implementation: Publisher Invalidation Hook

Location: `internal/cache/invalidator.go`

```go
package cache

import (
    "github.com/functionfly/functionfly/internal/storage"
)

type CacheInvalidator struct {
    service *CacheService
}

func NewCacheInvalidator(service *CacheService) *CacheInvalidator {
    return &CacheInvalidator{service: service}
}

// OnFunctionPublished - called when new version is published
func (i *CacheInvalidator) OnFunctionPublished(fn *storage.RegistryFunction, version *storage.RegistryFunctionVersion) error {
    // Version changed - invalidate old version cache
    return i.service.InvalidateVersion(fn.ID.String(), version.Version)
}

// OnFunctionUpdated - called when function metadata changes
func (i *CacheInvalidator) OnFunctionUpdated(fn *storage.RegistryFunction, oldVersion *storage.RegistryFunctionVersion) error {
    // Check if cache-affecting fields changed
    if oldVersion.Deterministic != oldVersion.Deterministic ||
       oldVersion.CacheTTL != oldVersion.CacheTTL ||
       oldVersion.SideEffects != oldVersion.SideEffects {
        return i.service.InvalidateFunction(fn.ID.String())
    }
    return nil
}

// PurgeAll - admin operation to clear all cache
func (i *CacheInvalidator) PurgeAll() error {
    // Truncate disk cache, clear memory
    // Implementation depends on storage
    return nil
}
```

### 6.3 Manual Invalidation API

```
DELETE /api/v1/cache/{function_id}
DELETE /api/v1/cache/{function_id}/{version}
DELETE /api/v1/cache  (admin only - clear all)
```

---

## 7. Cache Poisoning Protection

### 7.1 The Threat

A malicious function could return:
- Extremely large outputs (fill disk)
- Malformed data (break clients)
- Serialized objects with pointers (fail on other machines)
- Time bombs (expired data that looks fresh)

### 7.2 Protection Strategy

**Rule**: Never trust raw function output. Always re-serialize using platform serializer.

### 7.3 Implementation

Location: `internal/cache/validator.go`

```go
package cache

import (
    "encoding/json"
    "fmt"
    "time"
)

type OutputValidator struct {
    MaxOutputSize   int           // bytes
    MaxDepth        int           // JSON nesting
    AllowedTypes    []string      // allowed JSON types
}

func NewOutputValidator() *OutputValidator {
    return &OutputValidator{
        MaxOutputSize: 1024 * 1024, // 1MB max
        MaxDepth:      10,
        AllowedTypes:  []string{"object", "array", "string", "number", "boolean", "null"},
    }
}

// ValidateAndReSerialize ensures output is safe to cache
func (v *OutputValidator) ValidateAndReSerialize(output json.RawMessage) (json.RawMessage, error) {
    // Check size
    if len(output) > v.MaxOutputSize {
        return nil, fmt.Errorf("output exceeds max size: %d > %d", len(output), v.MaxOutputSize)
    }
    
    // Unmarshal to validate structure
    var raw interface{}
    if err := json.Unmarshal(output, &raw); err != nil {
        return nil, fmt.Errorf("invalid JSON output: %w", err)
    }
    
    // Validate depth
    if err := v.validateDepth(raw, 0); err != nil {
        return nil, err
    }
    
    // Re-serialize using canonical formatter (removes any weirdness)
    canonical, err := json.Marshal(raw)
    if err != nil {
        return nil, fmt.Errorf("failed to re-serialize: %w", err)
    }
    
    return canonical, nil
}

func (v *OutputValidator) validateDepth(val interface{}, depth int) error {
    if depth > v.MaxDepth {
        return fmt.Errorf("max nesting depth exceeded: %d", depth)
    }
    
    switch v := val.(type) {
    case map[string]interface{}:
        for _, val := range v {
            if err := v.validateDepth(val, depth+1); err != nil {
                return err
            }
        }
    case []interface{}:
        for _, item := range v {
            if err := v.validateDepth(item, depth+1); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### 7.4 Additional Protections

| Protection | Implementation |
|------------|-----------------|
| **Max output size** | Reject >1MB (configurable) |
| **Max depth** | Reject >10 levels nesting |
| **Serialization check** | Must be valid JSON |
| **Time bomb prevention** | Store only data, not timestamps |
| **Pointer rejection** | JSON has no pointers - safe |

---

## 8. Integration Points

### 8.1 Modified Execution Handler

The existing [`HandleExecute`](internal/api/handlers/registry/execution.go:20) needs modification:

```go
// In HandleExecute, after getting function version:
// Add cache service parameter

func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    // ... existing code to get function and version ...
    
    // Check cache eligibility
    eligibility := cache.CheckEligibility(fnVersion)
    
    // Read input
    body, err := io.ReadAll(r.Body)
    // ... handle error ...
    
    // Get from cache or execute
    result, err := h.cacheService.GetOrExecute(
        eligibility,
        body,
        func() (json.RawMessage, error) {
            // Execute function logic here
            return executeFunction(...)
        },
    )
    // ... handle error ...
    
    // Set CDN headers
    isPublic := fn.Visibility == "public"
    cache.SetCDNHeaders(w, eligibility, isPublic)
    
    // Return response with cached flag
    response := functionregistry.ExecutionResponse{
        OK:      true,
        Data:    result.Output,
        Cached:  result.Hit,
        DurationMs: durationMs,
    }
    // ...
}
```

### 8.2 Required Dependencies

```go
// go.mod additions
require (
    github.com/dgraph-io/ristretto v0.1.1
    github.com/google/uuid v1.5.0
)
```

### 8.3 Database Migration

```sql
-- Run this migration before deploying cache code
CREATE TABLE IF NOT EXISTS function_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key VARCHAR(255) NOT NULL UNIQUE,
    function_id UUID NOT NULL,
    version VARCHAR(50) NOT NULL,
    input_hash VARCHAR(64) NOT NULL,
    output_json JSONB NOT NULL,
    output_size INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    hit_count INTEGER NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMPTZ
);

CREATE INDEX idx_cache_key ON function_cache(cache_key);
CREATE INDEX idx_function_version ON function_cache(function_id, version);
CREATE INDEX idx_expires_at ON function_cache(expires_at);
```

---

## 9. Metrics & Observability

### 9.1 Cache-Specific Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `cache_hit_total` | Counter | Total cache hits by layer |
| `cache_miss_total` | Counter | Total cache misses |
| `cache_hit_ratio` | Gauge | Hits / (Hits + Misses) |
| `cache_size_bytes` | Gauge | Current cache size |
| `cache_memory_cost` | Gauge | Memory used by cache |
| `cache_eviction_total` | Counter | Entries evicted |
| `cache_error_total` | Counter | Cache errors |

### 9.2 Implementation

Location: `internal/cache/metrics.go`

```go
package cache

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    cacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "functionfly_cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"layer"}, // "memory", "disk"
    )
    
    cacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "functionfly_cache_misses_total",
            Help: "Total number of cache misses",
        },
    )
    
    cacheSize = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "functionfly_cache_size_bytes",
            Help: "Current cache size in bytes",
        },
    )
)
```

---

## 10. Cost Analysis

### 10.1 Without Cache

| Traffic Pattern | Compute Cost |
|-----------------|--------------|
| 10K calls/day | $0.01/day |
| 100K calls/day | $0.10/day |
| 1M calls/day | $1.00/day |

### 10.2 With Cache (90% hit rate)

| Traffic Pattern | Cache Cost | Compute Cost | Total Savings |
|-----------------|------------|--------------|----------------|
| 10K calls/day | ~$0 | $0.001 | 90% |
| 100K calls/day | ~$0.01 | $0.01 | 90% |
| 1M calls/day | ~$0.10 | $0.10 | 90% |

### 10.3 CDN Integration (Free)

For public deterministic functions with CDN headers:
- Cache hit at edge = $0 compute cost
- Only CDN bandwidth costs (often free tier)

---

## 11. File Structure Summary

```
internal/
├── cache/
│   ├── cache.go           # Main cache service interface
│   ├── eligibility.go     # Cache eligibility checker
│   ├── normalizer.go      # JSON normalization + hashing
│   ├── memory.go          # L1 memory cache (ristretto)
│   ├── disk.go            # L2 disk cache (SQLite)
│   ├── cdn.go             # L3 CDN headers
│   ├── invalidator.go     # Cache invalidation
│   ├── validator.go       # Poisoning protection
│   └── metrics.go         # Cache metrics
└── api/
    └── handlers/
        └── registry/
            └── execution.go  # Modified execution handler

migrations/
└── YYYYMMDD_cache_schema.sql
```

---

## 12. Summary

This caching system transforms FunctionFly:

| Before | After |
|--------|-------|
| Every call = compute | 90%+ calls = lookup |
| Cost scales with traffic | Cost approaches zero |
| Slow cold starts | Instant cache hits |
| Fragile under load | CDN absorbs traffic |

The key insight: `(function + version + input) → output` is a mathematical mapping. By normalizing inputs and storing results, we convert compute into lookups.

This is how a $5 server handles 1M requests/day.
