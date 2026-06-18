package client

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RouterConfig configures WasmPoolRouter.
type RouterConfig struct {
	// ExternalPercent: 0–100. If 0, every request is Local (default).
	ExternalPercent int

	// ExternalTenants: always-routed-to-external tenant IDs.
	ExternalTenants []string
	// LocalTenants: always-routed-to-local tenant IDs.
	// Wins on conflict with ExternalTenants (per the plan).
	LocalTenants []string

	// DryRun: when true, route to Local (authoritative) and fire-and-forget
	// the same request to External. Divergences are counted in metrics.
	DryRun bool

	// HashSeed: optional FNV seed for percentage routing. Changing this
	// reshuffles which tenants fall into the cohort. Defaults to 0.
	HashSeed uint32
}

// WasmPoolRouter chooses between External and Local on a per-request basis.
// It implements:
//   - Percentage rollout (ExternalPercent)
//   - Per-tenant overrides (ExternalTenants / LocalTenants, Local wins)
//   - Circuit-breaker fallback to Local
//   - Dry-run mode (Local authoritative, External best-effort)
type WasmPoolRouter struct {
	external WasmPoolClient
	local    WasmPoolClient
	breaker  *CircuitBreaker
	cfg      RouterConfig

	extSet map[string]struct{}
	locSet map[string]struct{}
	mu     sync.RWMutex
}

// NewRouter constructs a router. The external client may be nil in dev or
// when ExternalPercent=0; the router treats nil-external as "always Local".
func NewRouter(local, external WasmPoolClient, cfg RouterConfig) *WasmPoolRouter {
	r := &WasmPoolRouter{
		external: external,
		local:    local,
		breaker:  NewCircuitBreaker(),
		cfg:      cfg,
		extSet:   stringSet(cfg.ExternalTenants),
		locSet:   stringSet(cfg.LocalTenants),
	}
	go r.breakerMetricsLoop()
	return r
}

// Execute runs a request through the router. It is safe for concurrent use.
func (r *WasmPoolRouter) Execute(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	start := time.Now()
	decision, reason, target := r.decide(req)
	r.recordDecision(decision, reason)

	// Dry-run: Local is authoritative; External is fire-and-forget.
	if r.cfg.DryRun && target == r.local {
		go r.fireDryRun(req)
		return r.executeLocal(ctx, req, start)
	}

	// Local-only: ExternalPercent=0 or tenant override or circuit open.
	if target == r.local {
		return r.executeLocal(ctx, req, start)
	}

	// External path (wrapped in circuit breaker).
	if !r.breaker.Allow() {
		RoutingDecisions.WithLabelValues("local", "circuit_open").Inc()
		return r.executeLocal(ctx, req, start)
	}
	resp, err := r.external.Execute(ctx, req)
	ClientLatency.WithLabelValues("external", req.Runtime).Observe(time.Since(start).Seconds())
	if err != nil {
		r.breaker.OnFailure()
		logrus.WithError(err).WithField("tenant", req.TenantID).Warn("wasm pool: external failed, falling back to local")
		return r.executeLocal(ctx, req, start)
	}
	r.breaker.OnSuccess()
	return resp, nil
}

func (r *WasmPoolRouter) decide(req *Request) (decision, reason string, target WasmPoolClient) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.external == nil {
		return "local", "no_external", r.local
	}
	if _, ok := r.locSet[req.TenantID]; ok {
		return "local", "override", r.local
	}
	if _, ok := r.extSet[req.TenantID]; ok {
		return "external", "override", r.external
	}
	if r.cfg.ExternalPercent <= 0 {
		return "local", "percentage", r.local
	}
	if r.cfg.ExternalPercent >= 100 {
		return "external", "percentage", r.external
	}
	if r.cfg.DryRun {
		return "local", "dry_run", r.local
	}
	if hashInPercent(req.TenantID, r.cfg.ExternalPercent, r.cfg.HashSeed) {
		return "external", "percentage", r.external
	}
	return "local", "percentage", r.local
}

func (r *WasmPoolRouter) executeLocal(ctx context.Context, req *Request, start time.Time) (*Response, error) {
	resp, err := r.local.Execute(ctx, req)
	ClientLatency.WithLabelValues("local", req.Runtime).Observe(time.Since(start).Seconds())
	return resp, err
}

// fireDryRun sends a copy of the request to External and records any
// divergence from the Local result. The External response is discarded.
func (r *WasmPoolRouter) fireDryRun(req *Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	resp, err := r.external.Execute(ctx, req)
	latency := time.Since(start)
	diverges := func(field string, local, remote interface{}) {
		if local != remote {
			DryRunDivergences.WithLabelValues(field).Inc()
		}
	}
	if err != nil {
		diverges("error_local", "", err.Error())
		return
	}
	diverges("error", "", resp.Error)
	diverges("output_bytes", len(req.Input), len(resp.Output))
	diverges("latency_ms", "", latency.Milliseconds())
	diverges("cold_started", false, resp.ColdStarted)
}

func (r *WasmPoolRouter) recordDecision(decision, reason string) {
	RoutingDecisions.WithLabelValues(decision, reason).Inc()
}

func (r *WasmPoolRouter) breakerMetricsLoop() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s := r.breaker.State()
			for _, st := range []BreakerState{BreakerClosed, BreakerOpen, BreakerHalfOpen} {
				v := 0.0
				if st == s {
					v = 1.0
				}
				BreakerStateGauge.WithLabelValues(st.String()).Set(v)
			}
		}
	}
}

// hashInPercent returns true if tenantID's hash falls in the first Percent
// values [0, Percent). The plan says: `hash(tenantID) % 100 < percent`.
func hashInPercent(tenantID string, percent int, seed uint32) bool {
	if percent <= 0 {
		return false
	}
	if percent >= 100 {
		return true
	}
	h := fnv.New32a()
	if seed != 0 {
		_, _ = h.Write([]byte{byte(seed), byte(seed >> 8), byte(seed >> 16), byte(seed >> 24)})
	}
	_, _ = h.Write([]byte(tenantID))
	return int(h.Sum32()%100) < percent
}

func stringSet(items []string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m[s] = struct{}{}
	}
	return m
}

// LoadRouterConfigFromEnv reads RouterConfig from the env vars defined in
// the plan. The function is also useful for tests, which call it with
// t.Setenv.
func LoadRouterConfigFromEnv() RouterConfig {
	return RouterConfig{
		ExternalPercent:  envInt("WASM_POOL_EXTERNAL_PERCENT", 0),
		ExternalTenants: splitCSV(os.Getenv("WASM_POOL_EXTERNAL_TENANTS")),
		LocalTenants:    splitCSV(os.Getenv("WASM_POOL_LOCAL_TENANTS")),
		DryRun:          envBool("WASM_POOL_EXTERNAL_DRY_RUN", false),
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v == "true" || v == "1"
}
