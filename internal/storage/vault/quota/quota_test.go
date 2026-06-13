package quota

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetPlanQuota_AllPlansHaveAllResources(t *testing.T) {
	resources := []Resource{
		ResourceSecrets, ResourceDynamicCreds,
		ResourceTokensPerSecret, ResourceAuditExports,
	}
	plans := []Plan{PlanFree, PlanPro, PlanTeam, PlanEnterprise}
	for _, p := range plans {
		for _, r := range resources {
			q := GetPlanQuota(p, r)
			if q.Limit <= 0 {
				t.Errorf("%s / %s: limit must be > 0, got %d", p, r, q.Limit)
			}
			if q.Resource != r {
				t.Errorf("%s / %s: resource mismatch in result", p, r)
			}
		}
	}
}

func TestGetPlanQuota_HigherPlanHasMoreSecrets(t *testing.T) {
	free := GetPlanQuota(PlanFree, ResourceSecrets).Limit
	pro := GetPlanQuota(PlanPro, ResourceSecrets).Limit
	team := GetPlanQuota(PlanTeam, ResourceSecrets).Limit
	ent := GetPlanQuota(PlanEnterprise, ResourceSecrets).Limit
	if !(free < pro && pro < team && team < ent) {
		t.Fatalf("expected free<pro<team<ent, got %d %d %d %d", free, pro, team, ent)
	}
}

func TestGetPlanQuota_UnknownPlanReturnsZero(t *testing.T) {
	q := GetPlanQuota(Plan("unknown"), ResourceSecrets)
	if q.Limit != 0 {
		t.Fatalf("unknown plan: limit=%d, want 0", q.Limit)
	}
}

func TestGetPlanQuota_DynamicCredsHasWindow(t *testing.T) {
	q := GetPlanQuota(PlanPro, ResourceDynamicCreds)
	if q.Window == 0 {
		t.Fatal("dynamic creds must have a 30d window")
	}
	if q.Window != 30*24*time.Hour {
		t.Fatalf("window=%s, want 30d", q.Window)
	}
}

// fakeStore is a hand-rolled in-memory quota.Store for unit tests.
type fakeStore struct {
	plan      Plan
	overrides map[Resource]int64
	secretN   int64
	tokens    int64
	dynCreds  int64
	auditExp  int64
}

func (f *fakeStore) GetTenantPlan(_ context.Context, _ uuid.UUID) (Plan, error) {
	return f.plan, nil
}
func (f *fakeStore) GetOverride(_ context.Context, _ uuid.UUID, r Resource) (int64, time.Duration, bool, error) {
	if v, ok := f.overrides[r]; ok {
		return v, 0, true, nil
	}
	return 0, 0, false, nil
}
func (f *fakeStore) CountSecrets(_ context.Context, _ uuid.UUID) (int64, error) {
	return f.secretN, nil
}
func (f *fakeStore) CountActiveTokens(_ context.Context, _ uuid.UUID) (int64, error) {
	return f.tokens, nil
}
func (f *fakeStore) CountDynamicCredsSince(_ context.Context, _ uuid.UUID, _ time.Time) (int64, error) {
	return f.dynCreds, nil
}
func (f *fakeStore) CountAuditExportsSince(_ context.Context, _ uuid.UUID, _ time.Time) (int64, error) {
	return f.auditExp, nil
}

func TestEnforcer_SecretCount_BelowLimit(t *testing.T) {
	store := &fakeStore{plan: PlanFree, secretN: 10}
	e := NewEnforcer(store)
	d, err := e.CheckSecretCount(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed {
		t.Fatalf("should be allowed: %+v", d)
	}
	if d.Limit != 25 {
		t.Fatalf("free plan secret limit: got %d want 25", d.Limit)
	}
	if d.Remaining != 15 {
		t.Fatalf("remaining: got %d want 15", d.Remaining)
	}
}

func TestEnforcer_SecretCount_AtLimit(t *testing.T) {
	store := &fakeStore{plan: PlanFree, secretN: 25}
	e := NewEnforcer(store)
	d, _ := e.CheckSecretCount(context.Background(), uuid.New())
	if d.Allowed {
		t.Fatal("at-limit must not be allowed")
	}
	if d.Remaining != 0 {
		t.Fatalf("remaining=%d, want 0", d.Remaining)
	}
}

func TestEnforcer_SecretCount_OverrideWins(t *testing.T) {
	store := &fakeStore{
		plan:      PlanFree,
		secretN:   50, // already past free default
		overrides: map[Resource]int64{ResourceSecrets: 100},
	}
	e := NewEnforcer(store)
	d, _ := e.CheckSecretCount(context.Background(), uuid.New())
	if !d.Allowed {
		t.Fatalf("override 100 should permit 50: %+v", d)
	}
	if d.Limit != 100 {
		t.Fatalf("limit=%d, want 100 (override)", d.Limit)
	}
}

func TestEnforcer_HeadersAlwaysPresent(t *testing.T) {
	store := &fakeStore{plan: PlanPro, secretN: 5}
	e := NewEnforcer(store)
	d, _ := e.CheckSecretCount(context.Background(), uuid.New())
	if d.Headers["X-RateLimit-Limit"] == "" {
		t.Fatal("missing X-RateLimit-Limit")
	}
	if d.Headers["X-RateLimit-Remaining"] == "" {
		t.Fatal("missing X-RateLimit-Remaining")
	}
}

func TestEnforcer_DynamicCredsHasReset(t *testing.T) {
	store := &fakeStore{plan: PlanPro, dynCreds: 100}
	e := NewEnforcer(store)
	d, _ := e.CheckDynamicCreds(context.Background(), uuid.New())
	if !d.Allowed {
		t.Fatal("100 < 5000 should be allowed")
	}
	if d.Headers["X-RateLimit-Reset"] == "" {
		t.Fatal("rolling-window resource must set X-RateLimit-Reset")
	}
}

func TestEnforcer_TokensPerSecret(t *testing.T) {
	store := &fakeStore{plan: PlanFree, tokens: 4}
	e := NewEnforcer(store)
	d, _ := e.CheckTokensPerSecret(context.Background(), uuid.New())
	if !d.Allowed {
		t.Fatal("4 < 5 free should be allowed")
	}
	if d.Limit != 5 {
		t.Fatalf("free plan token limit: %d, want 5", d.Limit)
	}
}

func TestDecision_ClampsNegativeRemaining(t *testing.T) {
	d := Decision{Limit: 10, Current: 12}
	e := &Enforcer{}
	out := e.decide(d.Current, d.Limit, 0, time.Time{})
	if out.Remaining != 0 {
		t.Fatalf("remaining=%d, want 0 (clamped)", out.Remaining)
	}
	if out.Allowed {
		t.Fatal("12 > 10 must not be allowed")
	}
}

func TestMarshalDecision_StableJSON(t *testing.T) {
	d := Decision{Allowed: true, Limit: 25, Current: 5, Remaining: 20}
	got := MarshalDecision(d)
	if got == "" {
		t.Fatal("marshal returned empty")
	}
	// Spot-check that key fields appear in the output.
	for _, want := range []string{`"allowed":true`, `"limit":25`, `"remaining":20`} {
		if !contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
