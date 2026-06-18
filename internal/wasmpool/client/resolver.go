package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ResolverConfig controls periodic DNS resolution of a headless K8s Service.
// The plan recommends a 30s cache TTL.
type ResolverConfig struct {
	Addr    string        // headless service DNS name, e.g. "wasm-pool-service:8084"
	CacheTTL time.Duration // how long a successful resolution is reused
	OnChange func(addrs []string) // optional callback when the endpoint set changes
}

// Resolver periodically resolves a headless K8s Service and tracks changes.
// The first resolution happens synchronously in Start(); subsequent ones run
// on the background ticker.
type Resolver struct {
	cfg ResolverConfig

	mu       sync.RWMutex
	cached   []string
	cachedAt time.Time

	stop chan struct{}
	once sync.Once
}

// NewResolver constructs a resolver. Call Start to begin.
func NewResolver(cfg ResolverConfig) *Resolver {
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 30 * time.Second
	}
	return &Resolver{cfg: cfg, stop: make(chan struct{})}
}

// Start performs the first resolution synchronously, then starts the background
// loop. Returns an error if the first resolution fails (which is fatal at
// startup — the plan says "pre-populate the ring from a config map or DNS
// lookup at startup, before the first Execute call").
func (r *Resolver) Start(ctx context.Context) error {
	addrs, err := resolveHeadless(ctx, r.cfg.Addr)
	if err != nil {
		return err
	}
	r.setCached(addrs, time.Now())
	go r.loop()
	return nil
}

// Stop terminates the background loop.
func (r *Resolver) Stop() {
	r.once.Do(func() { close(r.stop) })
}

// Endpoints returns the most recently resolved endpoints (cached).
func (r *Resolver) Endpoints() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.cached))
	copy(out, r.cached)
	return out
}

func (r *Resolver) setCached(addrs []string, at time.Time) {
	r.mu.Lock()
	r.cached = addrs
	r.cachedAt = at
	r.mu.Unlock()
	if r.cfg.OnChange != nil {
		r.cfg.OnChange(addrs)
	}
}

func (r *Resolver) loop() {
	t := time.NewTicker(r.cfg.CacheTTL)
	defer t.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-t.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			addrs, err := resolveHeadless(ctx, r.cfg.Addr)
			cancel()
			if err != nil {
				continue
			}
			r.setCached(addrs, time.Now())
		}
	}
}

// netLookupHost is a package-level indirection so tests can stub it.
var netLookupHost = func(ctx context.Context, host string) ([]string, error) {
	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

func splitHostPort(addr string) (string, string, error) {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("missing port in %s", addr)
	}
	host := addr[:idx]
	port := addr[idx+1:]
	if _, err := strconv.Atoi(port); err != nil {
		return "", "", fmt.Errorf("invalid port %s: %w", port, err)
	}
	return host, port, nil
}

func joinHostPort(host, port string) string {
	return net.JoinHostPort(host, port)
}

// EnvAddr returns the headless service DNS name from WASM_POOL_SERVICE_ADDR
// or the well-known default.
func EnvAddr() string {
	if v := os.Getenv("WASM_POOL_SERVICE_ADDR"); v != "" {
		return v
	}
	return "wasm-pool-service:8084"
}
