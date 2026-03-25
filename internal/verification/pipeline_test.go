package verification

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestVerificationLevelString(t *testing.T) {
	tests := []struct {
		level    VerificationLevel
		expected string
	}{
		{Level0Unverified, "unverified"},
		{Level1Basic, "basic"},
		{Level2Standard, "standard"},
		{Level3Full, "full"},
		{VerificationLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("VerificationLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseVerificationLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected VerificationLevel
		wantErr  bool
	}{
		{"unverified", Level0Unverified, false},
		{"0", Level0Unverified, false},
		{"basic", Level1Basic, false},
		{"1", Level1Basic, false},
		{"standard", Level2Standard, false},
		{"2", Level2Standard, false},
		{"full", Level3Full, false},
		{"3", Level3Full, false},
		{"invalid", Level0Unverified, true},
		{"", Level0Unverified, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVerificationLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVerificationLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseVerificationLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPipelineRun(t *testing.T) {
	config := PipelineConfig{
		EnableManualReview: true,
		DREConfig: DREConfig{
			MinPassRate:   0.95,
			MinExecutions: 10,
		},
		FXCERTConfig: FXCERTConfig{
			CertificateValidityDays: 30,
			MaxLatencyMs:          5000,
			MinSuccessRate:        0.99,
		},
	}

	pipeline := NewPipeline(config)

	functionID := uuid.New()
	versionID := uuid.New()

	// Test Level 1 Basic
	result, err := pipeline.Run(context.Background(), functionID, versionID, Level1Basic)
	if err != nil {
		t.Fatalf("Pipeline.Run() error = %v", err)
	}

	if result.FunctionID != functionID {
		t.Errorf("FunctionID = %v, want %v", result.FunctionID, functionID)
	}

	if result.FunctionVersionID != versionID {
		t.Errorf("FunctionVersionID = %v, want %v", result.FunctionVersionID, versionID)
	}

	if result.Level != Level1Basic {
		t.Errorf("Level = %v, want %v", result.Level, Level1Basic)
	}

	if result.CompletedAt == nil {
		t.Error("CompletedAt should not be nil after Run()")
	}
}

func TestPipelineStagesForLevel(t *testing.T) {
	config := PipelineConfig{
		EnableManualReview: true,
	}

	pipeline := NewPipeline(config)

	tests := []struct {
		level          VerificationLevel
		expectedStages []VerificationStage
	}{
		{Level0Unverified, []VerificationStage{}},
		{Level1Basic, []VerificationStage{StageMalwareScan}},
		{Level2Standard, []VerificationStage{StageMalwareScan, StageDRE, StageFXCERT}},
		{Level3Full, []VerificationStage{StageMalwareScan, StageDRE, StageFXCERT, StageManualReview}},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			stages := pipeline.getStagesForLevel(tt.level)
			if len(stages) != len(tt.expectedStages) {
				t.Errorf("getStagesForLevel() returned %d stages, want %d", len(stages), len(tt.expectedStages))
				return
			}
			for i, stage := range stages {
				if stage != tt.expectedStages[i] {
					t.Errorf("getStagesForLevel()[%d] = %v, want %v", i, stage, tt.expectedStages[i])
				}
			}
		})
	}
}

func TestPipelineWithoutManualReview(t *testing.T) {
	config := PipelineConfig{
		EnableManualReview: false,
	}

	pipeline := NewPipeline(config)

	functionID := uuid.New()
	versionID := uuid.New()

	result, err := pipeline.Run(context.Background(), functionID, versionID, Level3Full)
	if err != nil {
		t.Fatalf("Pipeline.Run() error = %v", err)
	}

	// Level 3 without manual review enabled should not have manual review stage
	if _, hasManualReview := result.Stages[StageManualReview]; hasManualReview {
		t.Error("Manual review stage should not be present when EnableManualReview is false")
	}
}

func TestGetVerificationLevels(t *testing.T) {
	levels := GetVerificationLevels()

	if len(levels) != 4 {
		t.Errorf("GetVerificationLevels() returned %d levels, want 4", len(levels))
	}

	expectedLevels := []VerificationLevel{Level0Unverified, Level1Basic, Level2Standard, Level3Full}
	for i, expected := range expectedLevels {
		if levels[i].Level != expected {
			t.Errorf("levels[%d].Level = %v, want %v", i, levels[i].Level, expected)
		}
	}

	// Check Level1Basic has malware scan required
	if !levels[1].RequiresMalwareScan {
		t.Error("Level1Basic should require malware scan")
	}

	// Check Level3Full has all requirements
	if !levels[3].RequiresMalwareScan || !levels[3].RequiresDRE || !levels[3].RequiresFXCERT || !levels[3].RequiresManualReview {
		t.Error("Level3Full should require all checks")
	}
}
