package dna

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func testRepo(t *testing.T) *Repository {
	t.Helper()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("TEST_DB_HOST", "localhost"),
		getenv("TEST_DB_PORT", "5432"),
		getenv("TEST_DB_USER", "postgres"),
		getenv("TEST_DB_PASSWORD", "postgres"),
		getenv("TEST_DB_NAME", "functionfly"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration test: cannot open DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping integration test: cannot ping DB: %v", err)
	}
	return NewRepository(db)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func cleanupTestData(t *testing.T, repo *Repository, functionID string) {
	t.Helper()
	ctx := context.Background()
	repo.db.ExecContext(ctx, "DELETE FROM function_dna_analysis_queue WHERE function_id = $1", functionID)
	repo.db.ExecContext(ctx, "DELETE FROM function_dna_mutations WHERE function_id = $1", functionID)
	repo.db.ExecContext(ctx, "DELETE FROM function_dna_execution_metrics WHERE function_id = $1", functionID)
	repo.db.ExecContext(ctx, "DELETE FROM function_dna_profiles WHERE function_id = $1", functionID)
}

func TestIntegration_GetOrCreateProfile(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	p, err := repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")
	if err != nil {
		t.Fatalf("GetOrCreateProfile: %v", err)
	}
	if p.FunctionID != fid {
		t.Errorf("FunctionID = %s, want %s", p.FunctionID, fid)
	}
	if p.TenantID != "test-tenant" {
		t.Errorf("TenantID = %s, want test-tenant", p.TenantID)
	}
	if p.FitnessScore != 50.0 {
		t.Errorf("FitnessScore = %f, want 50.0", p.FitnessScore)
	}
	if !p.EvolutionEnabled {
		t.Error("expected evolution enabled by default")
	}

	p2, err := repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")
	if err != nil {
		t.Fatalf("GetOrCreateProfile (2nd call): %v", err)
	}
	if p2.ID != p.ID {
		t.Errorf("expected same profile on 2nd call, got different IDs")
	}
}

func TestIntegration_GetOrCreateProfile_ConcurrentSafety(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")
			errs <- err
		}()
	}
	for i := 0; i < 5; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent GetOrCreateProfile failed: %v", err)
		}
	}
}

func TestIntegration_UpdateProfile(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	p, _ := repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")
	p.FitnessScore = 85.5
	p.TotalExecutions = 5000
	p.Generation = 3
	hash := "sha256:abc123"
	p.DNAHash = &hash

	if err := repo.UpdateProfile(ctx, p); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	p2, _ := repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")
	if p2.FitnessScore != 85.5 {
		t.Errorf("FitnessScore = %f, want 85.5", p2.FitnessScore)
	}
	if p2.TotalExecutions != 5000 {
		t.Errorf("TotalExecutions = %d, want 5000", p2.TotalExecutions)
	}
	if p2.DNAHash == nil || *p2.DNAHash != "sha256:abc123" {
		t.Errorf("DNAHash = %v, want sha256:abc123", p2.DNAHash)
	}
}

func TestIntegration_SetEvolutionEnabled(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	repo.GetOrCreateProfile(ctx, fid, "registry", "test-tenant")

	if err := repo.SetEvolutionEnabled(ctx, fid, "registry", false); err != nil {
		t.Fatalf("SetEvolutionEnabled(false): %v", err)
	}
	p, _ := repo.GetProfileReadOnly(ctx, fid, "registry")
	if p.EvolutionEnabled {
		t.Error("expected evolution disabled")
	}

	if err := repo.SetEvolutionEnabled(ctx, fid, "registry", true); err != nil {
		t.Fatalf("SetEvolutionEnabled(true): %v", err)
	}
	p, _ = repo.GetProfileReadOnly(ctx, fid, "registry")
	if !p.EvolutionEnabled {
		t.Error("expected evolution re-enabled")
	}
}

func TestIntegration_InsertExecutionMetric(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	statusCode := 200
	region := "us-east-1"
	m := &ExecutionMetric{
		FunctionID:    fid,
		FunctionType:  "registry",
		ExecutionID:   strPtr(uuid.New().String()),
		DurationMs:    150,
		StatusCode:    &statusCode,
		ErrorCategory: "none",
		Region:        &region,
	}

	if err := repo.InsertExecutionMetric(ctx, m); err != nil {
		t.Fatalf("InsertExecutionMetric: %v", err)
	}

	m2 := &ExecutionMetric{
		FunctionID:    fid,
		FunctionType:  "registry",
		ExecutionID:   strPtr(uuid.New().String()),
		DurationMs:    5000,
		StatusCode:    intPtr(500),
		ErrorCategory: "runtime",
		ColdStart:     true,
	}
	repo.InsertExecutionMetric(ctx, m2)
}

func TestIntegration_AggregateMetrics(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	for i := 0; i < 20; i++ {
		code := 200
		cat := "none"
		if i >= 18 {
			code = 500
			cat = "runtime"
		}
		repo.InsertExecutionMetric(ctx, &ExecutionMetric{
			FunctionID:    fid,
			FunctionType:  "registry",
			ExecutionID:   strPtr(uuid.New().String()),
			DurationMs:    100 + i*10,
			StatusCode:    &code,
			ErrorCategory: cat,
			ColdStart:     i < 3,
		})
	}

	metrics, err := repo.AggregateMetrics(ctx, fid, 1*time.Hour)
	if err != nil {
		t.Fatalf("AggregateMetrics: %v", err)
	}
	if metrics.TotalExecutions != 20 {
		t.Errorf("TotalExecutions = %d, want 20", metrics.TotalExecutions)
	}
	if metrics.SuccessRate < 0.85 || metrics.SuccessRate > 0.95 {
		t.Errorf("SuccessRate = %f, want ~0.90", metrics.SuccessRate)
	}
	if metrics.ColdStartRate < 0.1 || metrics.ColdStartRate > 0.2 {
		t.Errorf("ColdStartRate = %f, want ~0.15", metrics.ColdStartRate)
	}
	if len(metrics.ErrorDistribution) != 1 {
		t.Errorf("ErrorDistribution len = %d, want 1", len(metrics.ErrorDistribution))
	}
}

func TestIntegration_MutationLifecycle(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	reason := "P99 latency above threshold"
	origCode := "func handler() { time.Sleep(1*time.Second) }"
	mutCode := "func handler() { /* optimized */ }"
	conf := 0.85
	model := "gpt-4"

	m := &Mutation{
		FunctionID:    fid,
		FunctionType:  "registry",
		TenantID:      "test-tenant",
		Generation:    1,
		MutationType:  "optimize_latency",
		Status:        "proposed",
		TriggerReason: &reason,
		OriginalCode:  &origCode,
		MutatedCode:   &mutCode,
		OriginalHash:  strPtr("sha256:abc"),
		MutatedHash:   strPtr("sha256:def"),
		Diff:          strPtr("--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new"),
		EstimatedImpact: json.RawMessage(`{"latency_improvement_pct": 40.0}`),
		Confidence:    &conf,
		ModelUsed:     &model,
	}

	if err := repo.CreateMutation(ctx, m); err != nil {
		t.Fatalf("CreateMutation: %v", err)
	}
	if m.ID == "" {
		t.Fatal("expected mutation ID to be set")
	}

	got, err := repo.GetMutation(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetMutation: %v", err)
	}
	if got.Status != "proposed" {
		t.Errorf("Status = %s, want proposed", got.Status)
	}
	if got.OriginalCode == nil || *got.OriginalCode != origCode {
		t.Error("OriginalCode mismatch")
	}

	if err := repo.UpdateMutationStatus(ctx, m.ID, "accepted", map[string]interface{}{"accepted_by": "user-1"}); err != nil {
		t.Fatalf("UpdateMutationStatus(accepted): %v", err)
	}
	got, _ = repo.GetMutation(ctx, m.ID)
	if got.Status != "accepted" {
		t.Errorf("Status = %s, want accepted", got.Status)
	}
	if got.AcceptedBy == nil || *got.AcceptedBy != "user-1" {
		t.Error("AcceptedBy mismatch")
	}

	if err := repo.UpdateMutationStatus(ctx, m.ID, "rejected", map[string]interface{}{"reason": "not good enough"}); err != nil {
		t.Fatalf("UpdateMutationStatus(rejected): %v", err)
	}
	got, _ = repo.GetMutation(ctx, m.ID)
	if got.Status != "rejected" {
		t.Errorf("Status = %s, want rejected", got.Status)
	}
}

func TestIntegration_ListMutations(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	for i := 0; i < 5; i++ {
		m := &Mutation{
			FunctionID:      fid,
			FunctionType:    "registry",
			TenantID:        "test-tenant",
			Generation:      i + 1,
			MutationType:    "optimize_latency",
			Status:          "proposed",
			EstimatedImpact: json.RawMessage(`{}`),
		}
		if err := repo.CreateMutation(ctx, m); err != nil {
			t.Fatalf("CreateMutation %d: %v", i, err)
		}
	}

	mutations, total, err := repo.ListMutations(ctx, fid, "", 10, 0)
	if err != nil {
		t.Fatalf("ListMutations: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(mutations) != 5 {
		t.Errorf("len(mutations) = %d, want 5", len(mutations))
	}

	mutations, total, _ = repo.ListMutations(ctx, fid, "proposed", 2, 0)
	if total != 5 {
		t.Errorf("total with status filter = %d, want 5", total)
	}
	if len(mutations) != 2 {
		t.Errorf("len with limit = %d, want 2", len(mutations))
	}
}

func TestIntegration_AnalysisQueue(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	if err := repo.EnqueueAnalysis(ctx, fid, "registry", "test-tenant", 3); err != nil {
		t.Fatalf("EnqueueAnalysis: %v", err)
	}
	if err := repo.EnqueueAnalysis(ctx, fid, "registry", "test-tenant", 1); err != nil {
		t.Fatalf("EnqueueAnalysis (2nd): %v", err)
	}

	id, dequeuedFid, _, err := repo.DequeueAnalysis(ctx)
	if err != nil {
		t.Fatalf("DequeueAnalysis: %v", err)
	}
	if id == "" {
		t.Fatal("expected a queued item")
	}
	if dequeuedFid != fid {
		t.Errorf("DequeueAnalysis functionID = %s, want %s", dequeuedFid, fid)
	}

	if err := repo.CompleteAnalysis(ctx, id); err != nil {
		t.Fatalf("CompleteAnalysis: %v", err)
	}

	id2, _, _, _ := repo.DequeueAnalysis(ctx)
	if id2 == "" {
		t.Fatal("expected 2nd queued item")
	}
	if err := repo.FailAnalysis(ctx, id2, "test error"); err != nil {
		t.Fatalf("FailAnalysis: %v", err)
	}

	id3, _, _, _ := repo.DequeueAnalysis(ctx)
	if id3 != "" {
		t.Error("expected no more items after processing all")
	}
}

func TestIntegration_PartitionManagement(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	created, err := repo.CreateFuturePartitions(ctx, 3)
	if err != nil {
		t.Fatalf("CreateFuturePartitions: %v", err)
	}
	if created < 1 {
		t.Errorf("created = %d, want >= 1", created)
	}

	partitions, err := repo.ListPartitions(ctx)
	if err != nil {
		t.Fatalf("ListPartitions: %v", err)
	}
	if len(partitions) < 4 {
		t.Errorf("partitions = %d, want >= 4", len(partitions))
	}
}

func TestIntegration_GetProfileReadOnly_NonExistent(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	p, err := repo.GetProfileReadOnly(ctx, "non-existent-id", "registry")
	if err != nil {
		t.Fatalf("GetProfileReadOnly: %v", err)
	}
	if p != nil {
		t.Error("expected nil for non-existent profile")
	}
}

func TestIntegration_GetDistinctTenantIDs(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	fid := uuid.New().String()
	defer cleanupTestData(t, repo, fid)

	repo.GetOrCreateProfile(ctx, fid, "registry", "unique-test-tenant-"+fid[:8])

	tenants, err := repo.GetDistinctTenantIDs(ctx)
	if err != nil {
		t.Fatalf("GetDistinctTenantIDs: %v", err)
	}
	found := false
	for _, tid := range tenants {
		if tid == "unique-test-tenant-"+fid[:8] {
			found = true
		}
	}
	if !found {
		t.Error("expected to find test tenant in distinct list")
	}
}

func TestIntegration_InsertInsight(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	tenantID := "insight-test-" + uuid.New().String()[:8]

	ins := &TenantInsights{
		TotalFunctionsAnalyzed:   10,
		TotalMutationsProposed:   5,
		TotalMutationsAccepted:   3,
		AvgFitnessScore:          75.5,
		AvgLatencyImprovementPct: 20.0,
		TotalCostSavingsUSD:      100.0,
		TopBottleneckCategories:  json.RawMessage(`[{"category":"timeout","count":5}]`),
		EvolutionLeaderboard:     json.RawMessage(`[{"function_id":"abc","generation":3,"fitness_score":90}]`),
	}

	now := time.Now()
	if err := repo.InsertInsight(ctx, ins, tenantID, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("InsertInsight: %v", err)
	}

	var count int
	repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM function_dna_insights WHERE tenant_id = $1", tenantID).Scan(&count)
	if count != 1 {
		t.Errorf("insight count = %d, want 1", count)
	}
	repo.db.ExecContext(ctx, "DELETE FROM function_dna_insights WHERE tenant_id = $1", tenantID)
}

func strPtr(s string) *string { return &s }
func intPtr(v int) *int       { return &v }

func init() {
	logrus.SetLevel(logrus.WarnLevel)
}
