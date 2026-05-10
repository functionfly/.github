package timemachine

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"

	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/google/uuid"
)

func TestBuildMerkleTree_Empty(t *testing.T) {
	root := buildMerkleTree(nil)
	if len(root) == 0 {
		t.Error("expected non-empty root for nil input")
	}
}

func TestBuildMerkleTree_Single(t *testing.T) {
	hash := sha256.Sum256([]byte("test"))
	root := buildMerkleTree([][]byte{hash[:]})
	if hex.EncodeToString(root) != hex.EncodeToString(hash[:]) {
		t.Error("single leaf should equal root")
	}
}

func TestBuildMerkleTree_TwoLeaves(t *testing.T) {
	h1 := sha256.Sum256([]byte("a"))
	h2 := sha256.Sum256([]byte("b"))
	root := buildMerkleTree([][]byte{h1[:], h2[:]})
	if len(root) == 0 {
		t.Error("expected non-empty root")
	}
	combined := append(h1[:], h2[:]...)
	expected := sha256.Sum256(combined)
	if hex.EncodeToString(root) != hex.EncodeToString(expected[:]) {
		t.Error("root should be hash of concatenated children")
	}
}

func TestBuildMerkleTree_FourLeaves(t *testing.T) {
	leaves := make([][]byte, 4)
	for i := range leaves {
		h := sha256.Sum256([]byte{byte(i)})
		leaves[i] = h[:]
	}
	root := buildMerkleTree(leaves)
	if len(root) == 0 {
		t.Error("expected non-empty root")
	}
}

func TestBuildMerkleTree_OddLeaves(t *testing.T) {
	leaves := make([][]byte, 3)
	for i := range leaves {
		h := sha256.Sum256([]byte{byte(i)})
		leaves[i] = h[:]
	}
	root := buildMerkleTree(leaves)
	if len(root) == 0 {
		t.Error("expected non-empty root for odd number of leaves")
	}
}

func TestBuildMerkleTree_Deterministic(t *testing.T) {
	leaves := make([][]byte, 5)
	for i := range leaves {
		h := sha256.Sum256([]byte{byte(i)})
		leaves[i] = h[:]
	}
	root1 := buildMerkleTree(leaves)
	root2 := buildMerkleTree(leaves)
	if hex.EncodeToString(root1) != hex.EncodeToString(root2) {
		t.Error("merkle tree should be deterministic")
	}
}

func TestAuditGenerator_buildSummary(t *testing.T) {
	gen := &AuditGenerator{}

	items := []tmstorage.ReplayItem{
		{Status: "completed", DiffType: sql.NullString{String: "identical", Valid: true}},
		{Status: "completed", DiffType: sql.NullString{String: "minor", Valid: true}},
		{Status: "completed", DiffType: sql.NullString{String: "major", Valid: true}},
		{Status: "completed", DiffType: sql.NullString{String: "breaking", Valid: true}},
		{Status: "failed"},
		{Status: "completed"},
	}

	s := gen.buildSummary(nil, items)
	if s.identical != 2 {
		t.Errorf("expected 2 identical, got %d", s.identical)
	}
	if s.minor != 1 {
		t.Errorf("expected 1 minor, got %d", s.minor)
	}
	if s.major != 1 {
		t.Errorf("expected 1 major, got %d", s.major)
	}
	if s.breaking != 1 {
		t.Errorf("expected 1 breaking, got %d", s.breaking)
	}
	if s.errors != 1 {
		t.Errorf("expected 1 error, got %d", s.errors)
	}
}

func TestAuditGenerator_buildItems(t *testing.T) {
	gen := &AuditGenerator{}
	newDur := int32(120)

	items := []tmstorage.ReplayItem{
		{
			OriginalExecutionID: uuid.New(),
			OriginalVersion:     "1.0.0",
			OriginalOutput:      json.RawMessage(`{"result":"ok"}`),
			OriginalDurationMs:  100,
			NewDurationMs:       sql.NullInt32{Int32: newDur, Valid: true},
			Status:              "completed",
			DiffType:            sql.NullString{String: "minor", Valid: true},
		},
	}

	auditItems := gen.buildItems(items, "2.0.0")
	if len(auditItems) != 1 {
		t.Fatalf("expected 1 audit item, got %d", len(auditItems))
	}

	item := auditItems[0]
	if item.OriginalVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", item.OriginalVersion)
	}
	if item.TargetVersion != "2.0.0" {
		t.Errorf("expected target 2.0.0, got %s", item.TargetVersion)
	}
	if item.DiffType != "minor" {
		t.Errorf("expected minor, got %s", item.DiffType)
	}
	if item.OriginalDuration != 100 {
		t.Errorf("expected 100ms, got %d", item.OriginalDuration)
	}
	if item.NewDuration != 120 {
		t.Errorf("expected 120ms, got %d", item.NewDuration)
	}
	if item.ItemHash == "" {
		t.Error("expected non-empty item hash")
	}
}

func TestAuditGenerator_buildSummaryJSON(t *testing.T) {
	gen := &AuditGenerator{}
	replay := &tmstorage.Replay{
		TargetVersion:      "2.0.0",
		Reason:             "bug fix",
		ReconciliationMode: "dry_run",
	}

	s := diffSummary{identical: 8, minor: 1, major: 1, breaking: 0, errors: 0}
	result := gen.buildSummaryJSON(replay, s)

	if result.TotalExecutions != 10 {
		t.Errorf("expected 10 total, got %d", result.TotalExecutions)
	}
	if result.Identical != 8 {
		t.Errorf("expected 8 identical, got %d", result.Identical)
	}
	if result.Changed != 2 {
		t.Errorf("expected 2 changed, got %d", result.Changed)
	}
	if result.SuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %.1f", result.SuccessRate)
	}
}

func TestReconciliationEngine_buildAction(t *testing.T) {
	engine := &ReconciliationEngine{}

	tests := []struct {
		name       string
		diffType   string
		expectType string
		expectRisk string
	}{
		{"minor", "minor", "update_output", "low"},
		{"major", "major", "update_output", "medium"},
		{"breaking", "breaking", "update_output_with_review", "high"},
		{"error", "error", "flag_error", "high"},
		{"identical", "identical", "no_action", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tmstorage.ReplayItem{
				OriginalExecutionID: uuid.New(),
				OriginalOutput:      json.RawMessage(`{}`),
				NewOutput:           json.RawMessage(`{}`),
			}

			action := engine.buildAction(item, tt.diffType)
			if action.Type != tt.expectType {
				t.Errorf("expected type %s, got %s", tt.expectType, action.Type)
			}
			if action.RiskLevel != tt.expectRisk {
				t.Errorf("expected risk %s, got %s", tt.expectRisk, action.RiskLevel)
			}
		})
	}
}

func TestNullStrPtr(t *testing.T) {
	result := nullStrPtr("test")
	if result == nil || *result != "test" {
		t.Error("expected non-nil pointer to 'test'")
	}

	result = nullStrPtr("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
}
