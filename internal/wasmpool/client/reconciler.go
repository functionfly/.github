package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PrewarmClient is the minimal HTTP surface the reconciler needs from
// wasm-pool-service. The plan calls for POST /admin/tenants/:id/prewarm.
type PrewarmClient interface {
	Prewarm(ctx context.Context, endpoint, tenantID, runtime string, count int) error
}

// HTTPPrewarmClient posts to the /admin/tenants/:id/prewarm endpoint over
// plain HTTP. The orchestrator's pod has direct L3 access to the headless
// service's pod IPs, so a one-shot HTTP call is the simplest path.
type HTTPPrewarmClient struct {
	client *http.Client
}

// NewHTTPPrewarmClient constructs a client with a short timeout per call.
func NewHTTPPrewarmClient() *HTTPPrewarmClient {
	return &HTTPPrewarmClient{client: &http.Client{Timeout: 10 * time.Second}}
}

// Prewarm calls the per-tenant prewarm endpoint. endpoint is a host:port
// (e.g. "10.0.5.23:8085") for a specific pool replica.
func (h *HTTPPrewarmClient) Prewarm(ctx context.Context, endpoint, tenantID, runtime string, count int) error {
	url := fmt.Sprintf("http://%s/admin/tenants/%s/prewarm", endpoint, tenantID)
	body, _ := json.Marshal(map[string]any{"runtime": runtime, "count": count})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("prewarm %s: status %d", tenantID, resp.StatusCode)
	}
	return nil
}

// Reconciler detects changes in the consistent-hash ring and proactively
// prewarms tenants whose owning replica changed, so the new owner has a
// warm instance when the orchestrator starts routing to it.
type Reconciler struct {
	prewarm    PrewarmClient
	external   *ExternalPoolClient
	tenants    []string // known tenant IDs to consider
	runtimes   []string // runtimes to prewarm
	maxConc    int
	tickPeriod time.Duration

	mu     sync.Mutex
	last   map[string]string // tenantID → last-known owner
	stop   chan struct{}
	once   sync.Once
}

// ReconcilerConfig configures the Reconciler.
type ReconcilerConfig struct {
	Tenants    []string      // tenant IDs to reconcile (from orchestrator's tenant list)
	Runtimes   []string      // runtimes to prewarm (e.g. ["python"])
	MaxConc    int           // max simultaneous prewarm calls (plan: 10)
	TickPeriod time.Duration // how often to re-check the ring (plan: 30s)
}

// NewReconciler constructs a reconciler.
func NewReconciler(cfg ReconcilerConfig, prewarm PrewarmClient, ext *ExternalPoolClient) *Reconciler {
	if cfg.MaxConc <= 0 {
		cfg.MaxConc = 10
	}
	if cfg.TickPeriod <= 0 {
		cfg.TickPeriod = 30 * time.Second
	}
	return &Reconciler{
		prewarm:    prewarm,
		external:   ext,
		tenants:    cfg.Tenants,
		runtimes:   cfg.Runtimes,
		maxConc:    cfg.MaxConc,
		tickPeriod: cfg.TickPeriod,
		last:       make(map[string]string),
		stop:       make(chan struct{}),
	}
}

// Start runs the reconciler loop. Returns immediately; the loop runs in a
// goroutine. Call Stop() to terminate.
func (r *Reconciler) Start(ctx context.Context) {
	// Seed with the current owners so the first tick only fires for
	// *changes*, not the initial state.
	for _, t := range r.tenants {
		r.last[t] = r.external.OwnerOf(t)
	}
	go r.loop(ctx)
}

// Stop terminates the reconciler loop.
func (r *Reconciler) Stop() {
	r.once.Do(func() { close(r.stop) })
}

func (r *Reconciler) loop(ctx context.Context) {
	t := time.NewTicker(r.tickPeriod)
	defer t.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ctx.Done():
			return
		case <-t.C:
			if err := r.reconcile(ctx); err != nil {
				logrus.WithError(err).Warn("pool reconciler: tick failed")
			}
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	// Refresh the external client's ring first so OwnerOf reflects the
	// latest endpoints. Errors are logged but don't stop reconciliation.
	if err := r.external.RefreshEndpoints(ctx); err != nil {
		logrus.WithError(err).Debug("pool reconciler: refresh endpoints")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	type job struct{ tenant, runtime string }
	var jobs []job
	for _, t := range r.tenants {
		owner := r.external.OwnerOf(t)
		if owner == "" {
			continue
		}
		if prev, ok := r.last[t]; !ok || prev != owner {
			for _, runtime := range r.runtimes {
				jobs = append(jobs, job{tenant: t, runtime: runtime})
			}
		}
		r.last[t] = owner
	}
	if len(jobs) == 0 {
		return nil
	}
	logrus.WithField("jobs", len(jobs)).Info("pool reconciler: prewarming shifted tenants")

	sem := make(chan struct{}, r.maxConc)
	var wg sync.WaitGroup
	for _, j := range jobs {
		sem <- struct{}{}
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			defer func() { <-sem }()
			owner := r.external.OwnerOf(j.tenant)
			if owner == "" {
				return
			}
			if err := r.prewarm.Prewarm(ctx, owner, j.tenant, j.runtime, 1); err != nil {
				logrus.WithError(err).WithField("tenant", j.tenant).
					Warn("pool reconciler: prewarm failed")
			}
		}(j)
	}
	wg.Wait()
	return nil
}
