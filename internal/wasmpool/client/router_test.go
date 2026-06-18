package client

import (
	"context"
	"testing"
	"time"
)

// TestCircuitBreaker_TripsAfterThreshold verifies the 5-failure / 30s window
// from the plan: after 5 consecutive failures the breaker opens and stays
// open until the window elapses.
func TestCircuitBreaker_TripsAfterThreshold(t *testing.T) {
	b := NewCircuitBreaker()
	for i := 0; i < b.Threshold; i++ {
		if !b.Allow() {
			t.Fatalf("Allow() returned false at failure %d while still closed", i)
		}
		b.OnFailure()
	}
	if b.State() != BreakerOpen {
		t.Fatalf("expected breaker open after %d failures, got %s", b.Threshold, b.State())
	}
	if b.Allow() {
		t.Fatal("Allow() returned true while open")
	}
}

// TestCircuitBreaker_HalfOpenAfterWindow verifies that after the 30s window
// elapses, the breaker moves to half-open and allows exactly one probe.
func TestCircuitBreaker_HalfOpenAfterWindow(t *testing.T) {
	b := &CircuitBreaker{Threshold: 5, Window: 10 * time.Millisecond}
	for i := 0; i < 5; i++ {
		b.OnFailure()
	}
	if b.State() != BreakerOpen {
		t.Fatalf("expected open, got %s", b.State())
	}
	time.Sleep(20 * time.Millisecond)
	if b.State() != BreakerHalfOpen {
		t.Fatalf("expected half_open after window, got %s", b.State())
	}
	if !b.Allow() {
		t.Fatal("first half-open call should be allowed")
	}
	if b.Allow() {
		t.Fatal("second half-open call should be blocked")
	}
	b.OnSuccess()
	if b.State() != BreakerClosed {
		t.Fatalf("expected closed after probe success, got %s", b.State())
	}
}

// TestCircuitBreaker_HalfOpenFailureReopens verifies that a failed probe
// in half-open immediately re-opens the breaker.
func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	b := &CircuitBreaker{Threshold: 5, Window: 5 * time.Millisecond}
	for i := 0; i < 5; i++ {
		b.OnFailure()
	}
	time.Sleep(10 * time.Millisecond)
	if !b.Allow() {
		t.Fatal("half-open probe should be allowed")
	}
	b.OnFailure()
	if b.State() != BreakerOpen {
		t.Fatalf("expected re-open after half-open failure, got %s", b.State())
	}
}

// TestCircuitBreaker_SuccessResetsFailureCount verifies that a success
// resets the failure counter, so intermittent failures don't trip the
// breaker unless 5 happen in a row.
func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	b := NewCircuitBreaker()
	for i := 0; i < 4; i++ {
		b.Allow()
		b.OnFailure()
	}
	b.OnSuccess()
	for i := 0; i < 4; i++ {
		if !b.Allow() {
			t.Fatalf("Allow() returned false after success reset (iter %d)", i)
		}
		b.OnFailure()
	}
	if b.State() != BreakerClosed {
		t.Fatalf("expected closed after reset + 4 failures, got %s", b.State())
	}
}

// --- Router decision matrix ---

type stubClient struct {
	name string
	err  error
	data []byte
}

func (s *stubClient) Name() string                                     { return s.name }
func (s *stubClient) Close() error                                     { return nil }
func (s *stubClient) Execute(ctx context.Context, req *Request) (*Response, error) {
	return &Response{Output: s.data, Error: errString(s.err)}, s.err
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func newRouterPair(extErr error) (*WasmPoolRouter, *stubClient, *stubClient) {
	ext := &stubClient{name: "external", err: extErr, data: []byte("ext")}
	loc := &stubClient{name: "local", data: []byte("loc")}
	return NewRouter(loc, ext, RouterConfig{}), ext, loc
}

func TestRouter_DefaultsToLocal(t *testing.T) {
	r, _, _ := newRouterPair(nil)
	resp, err := r.Execute(context.Background(), &Request{TenantID: "t1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if string(resp.Output) != "loc" {
		t.Fatalf("expected local, got %q", resp.Output)
	}
}

func TestRouter_PercentageBoundary(t *testing.T) {
	// 0% → all local
	r := NewRouter(&stubClient{name: "l", data: []byte("loc")}, &stubClient{name: "e", data: []byte("ext")}, RouterConfig{ExternalPercent: 0})
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		resp, _ := r.Execute(context.Background(), &Request{TenantID: id})
		if string(resp.Output) != "loc" {
			t.Fatalf("0%%: expected local for %s, got external", id)
		}
	}
	// 100% → all external
	r = NewRouter(&stubClient{name: "l", data: []byte("loc")}, &stubClient{name: "e", data: []byte("ext")}, RouterConfig{ExternalPercent: 100})
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		resp, _ := r.Execute(context.Background(), &Request{TenantID: id})
		if string(resp.Output) != "ext" {
			t.Fatalf("100%%: expected external for %s, got local", id)
		}
	}
	// 50% → roughly half external
	r = NewRouter(&stubClient{name: "l", data: []byte("loc")}, &stubClient{name: "e", data: []byte("ext")}, RouterConfig{ExternalPercent: 50})
	ext, loc := 0, 0
	for i := 0; i < 200; i++ {
		tenant := "t-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i%10))
		resp, _ := r.Execute(context.Background(), &Request{TenantID: tenant})
		if string(resp.Output) == "ext" {
			ext++
		} else {
			loc++
		}
	}
	if ext < 70 || ext > 130 {
		t.Fatalf("50%%: expected ~100 external, got %d", ext)
	}
}

func TestRouter_LocalTenantWinsOverExternal(t *testing.T) {
	cfg := RouterConfig{
		ExternalPercent:  100,
		ExternalTenants:  []string{"force-ext"},
		LocalTenants:     []string{"force-ext"}, // conflict
	}
	r := NewRouter(&stubClient{name: "l"}, &stubClient{name: "e"}, cfg)
	resp, _ := r.Execute(context.Background(), &Request{TenantID: "force-ext"})
	if string(resp.Output) != "" {
		t.Fatal("local-tenants list should win on conflict")
	}
}

func TestRouter_CircuitOpenForcesLocal(t *testing.T) {
	r, _, _ := newRouterPair(nil)
	r.cfg.ExternalPercent = 100
	// Trip the breaker.
	for i := 0; i < r.breaker.Threshold; i++ {
		r.breaker.OnFailure()
	}
	resp, err := r.Execute(context.Background(), &Request{TenantID: "t1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if string(resp.Output) != "loc" {
		t.Fatalf("expected local fallback when breaker open, got external")
	}
}

func TestRouter_DryRunAlwaysLocal(t *testing.T) {
	r := NewRouter(&stubClient{name: "l"}, &stubClient{name: "e"}, RouterConfig{
		ExternalPercent: 100,
		DryRun:          true,
	})
	resp, _ := r.Execute(context.Background(), &Request{TenantID: "t1"})
	if string(resp.Output) != "" {
		t.Fatal("dry-run must always use local as authoritative")
	}
}

func TestRouter_ExternalFailureFallsBackToLocal(t *testing.T) {
	r, _, _ := newRouterPair(context.DeadlineExceeded)
	r.cfg.ExternalPercent = 100
	// Single failure: fallback to local, breaker still closed (1 < threshold).
	resp, err := r.Execute(context.Background(), &Request{TenantID: "t1"})
	if err != nil {
		t.Fatalf("Execute should not return error on fallback: %v", err)
	}
	if string(resp.Output) != "loc" {
		t.Fatal("expected local fallback on external error")
	}
	if r.breaker.State() != BreakerClosed {
		t.Fatalf("expected breaker closed after 1 failure, got %s", r.breaker.State())
	}
	// Trip the breaker with 5 failures and confirm fallback keeps working.
	for i := 0; i < r.breaker.Threshold; i++ {
		r.breaker.OnFailure()
	}
	if r.breaker.State() != BreakerOpen {
		t.Fatalf("expected breaker open after %d failures, got %s", r.breaker.Threshold, r.breaker.State())
	}
}

// --- hashInPercent ---

func TestHashInPercent(t *testing.T) {
	for _, c := range []struct {
		id  string
		pct int
		want bool
	}{
		// 0% and 100% are short-circuited before the hash is computed.
		{"x", 0, false},
		{"x", 100, true},
	} {
		if got := hashInPercent(c.id, c.pct, 0); got != c.want {
			t.Errorf("hashInPercent(%q, %d) = %v, want %v", c.id, c.pct, got, c.want)
		}
	}
}
