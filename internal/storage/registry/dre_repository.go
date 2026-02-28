package registry

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StoreMEGRecord persists a Merkle Execution Graph record.
// This is called asynchronously after each execution of a deterministic function.
func (r *RegistryRepository) StoreMEGRecord(rec *MEGRecord) error {
	if err := r.db.Create(rec).Error; err != nil {
		return fmt.Errorf("dre: store meg record: %w", err)
	}
	return nil
}

// GetMEGByExecutionID retrieves the MEG record for a specific execution.
func (r *RegistryRepository) GetMEGByExecutionID(executionID uuid.UUID) (*MEGRecord, error) {
	var rec MEGRecord
	err := r.db.Where("execution_id = ?", executionID).First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("dre: get meg by execution id: %w", err)
	}
	return &rec, nil
}

// UpdateMEGReplayResult updates the replay verification fields of a MEG record.
func (r *RegistryRepository) UpdateMEGReplayResult(megID uuid.UUID, replayRootHash, replayNodeID string, verifiedAt time.Time) error {
	if err := r.db.Model(&MEGRecord{}).Where("id = ?", megID).Updates(map[string]interface{}{
		"replay_root_hash":   replayRootHash,
		"replay_verified_at": verifiedAt,
		"replay_node_id":     replayNodeID,
	}).Error; err != nil {
		return fmt.Errorf("dre: update meg replay result: %w", err)
	}
	return nil
}

// StoreCertificate persists an FXCERT execution certificate.
func (r *RegistryRepository) StoreCertificate(cert *ExecutionCertificate) error {
	if err := r.db.Create(cert).Error; err != nil {
		return fmt.Errorf("dre: store certificate: %w", err)
	}
	return nil
}

// GetCertificateByID retrieves a certificate by its certificate_id (e.g., "fxc_01H...").
func (r *RegistryRepository) GetCertificateByID(certID string) (*ExecutionCertificate, error) {
	var cert ExecutionCertificate
	err := r.db.Where("certificate_id = ?", certID).First(&cert).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("dre: get certificate by id: %w", err)
	}
	return &cert, nil
}

// GetCertificatesByFunctionID lists certificates for a function (paginated).
func (r *RegistryRepository) GetCertificatesByFunctionID(functionID uuid.UUID, limit, offset int) ([]*ExecutionCertificate, error) {
	var certs []*ExecutionCertificate
	err := r.db.Where("function_id = ?", functionID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&certs).Error
	if err != nil {
		return nil, fmt.Errorf("dre: get certificates by function id: %w", err)
	}
	return certs, nil
}

// StoreDriftReport persists a drift report and updates the function's execution passport.
func (r *RegistryRepository) StoreDriftReport(report *DriftReportRecord) error {
	if err := r.db.Create(report).Error; err != nil {
		return fmt.Errorf("dre: store drift report: %w", err)
	}

	// Update passport: increment drift incidents
	now := time.Now()
	update := PassportUpdate{
		IncrementDrift: true,
		TrustPenalty:   report.TrustPenalty,
	}
	if err := r.UpdatePassport(report.FunctionID, update); err != nil {
		// Log but don't fail — drift report is already stored
		fmt.Printf("dre: update passport after drift: %v\n", err)
	}

	_ = now
	return nil
}

// GetOrCreatePassport retrieves or initializes the execution passport for a function.
func (r *RegistryRepository) GetOrCreatePassport(functionID uuid.UUID) (*ExecutionPassport, error) {
	var passport ExecutionPassport

	// Use upsert to avoid race conditions
	err := r.db.Where("function_id = ?", functionID).First(&passport).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new passport
			passport = ExecutionPassport{
				ID:                        uuid.New(),
				FunctionID:                functionID,
				DeterministicReliability:  0,
				ReplayDriftIncidents:      0,
				VerifiedExecutionsTotal:   0,
				TotalExecutions:           0,
				DeterminismScore:          0,
				ReplayIntegrityScore:      0,
				PerformanceStabilityScore: 0,
				DriftScore:                1.0,
				UpdatedAt:                 time.Now(),
			}
			if createErr := r.db.Create(&passport).Error; createErr != nil {
				return nil, fmt.Errorf("dre: create passport: %w", createErr)
			}
			return &passport, nil
		}
		return nil, fmt.Errorf("dre: get passport: %w", err)
	}

	return &passport, nil
}

// UpdatePassport updates the execution passport after each verified execution or drift event.
func (r *RegistryRepository) UpdatePassport(functionID uuid.UUID, update PassportUpdate) error {
	// Get current passport
	passport, err := r.GetOrCreatePassport(functionID)
	if err != nil {
		return fmt.Errorf("dre: update passport get: %w", err)
	}

	// Apply updates
	if update.IncrementTotal {
		passport.TotalExecutions++
	}
	if update.IncrementVerified {
		passport.VerifiedExecutionsTotal++
		if update.LastVerifiedAt != nil {
			passport.LastVerifiedAt = update.LastVerifiedAt
		} else {
			now := time.Now()
			passport.LastVerifiedAt = &now
		}
	}
	if update.IncrementDrift {
		passport.ReplayDriftIncidents++
	}

	// Update capsule version history
	if update.CapsuleDescriptorHash != "" {
		var versions []string
		if len(passport.CapsuleVersionsUsed) > 0 {
			_ = json.Unmarshal(passport.CapsuleVersionsUsed, &versions)
		}
		// Add if not already present
		found := false
		for _, v := range versions {
			if v == update.CapsuleDescriptorHash {
				found = true
				break
			}
		}
		if !found {
			versions = append(versions, update.CapsuleDescriptorHash)
			if b, err := json.Marshal(versions); err == nil {
				passport.CapsuleVersionsUsed = b
			}
		}
	}

	// Recompute DRE sub-scores
	passport.DeterminismScore = computeDeterminismScore(passport.VerifiedExecutionsTotal, passport.TotalExecutions)
	passport.ReplayIntegrityScore = computeReplayIntegrityScore(passport.ReplayDriftIncidents, passport.VerifiedExecutionsTotal)
	passport.DriftScore = computeDriftScore(passport.ReplayDriftIncidents)
	if passport.TotalExecutions > 0 {
		passport.DeterministicReliability = float64(passport.VerifiedExecutionsTotal) / float64(passport.TotalExecutions)
	}

	passport.UpdatedAt = time.Now()

	// Persist using upsert
	if err := r.db.Model(&ExecutionPassport{}).
		Where("function_id = ?", functionID).
		Updates(map[string]interface{}{
			"deterministic_reliability":    passport.DeterministicReliability,
			"replay_drift_incidents":       passport.ReplayDriftIncidents,
			"verified_executions_total":    passport.VerifiedExecutionsTotal,
			"total_executions":             passport.TotalExecutions,
			"determinism_score":            passport.DeterminismScore,
			"replay_integrity_score":       passport.ReplayIntegrityScore,
			"performance_stability_score":  passport.PerformanceStabilityScore,
			"drift_score":                  passport.DriftScore,
			"capsule_versions_used":        passport.CapsuleVersionsUsed,
			"last_verified_at":             passport.LastVerifiedAt,
			"updated_at":                   passport.UpdatedAt,
		}).Error; err != nil {
		return fmt.Errorf("dre: persist passport update: %w", err)
	}

	return nil
}

// GetDREScoresForTrust retrieves the 4 DRE sub-scores for TrustScore v2 calculation.
func (r *RegistryRepository) GetDREScoresForTrust(functionID uuid.UUID) (*DREScores, error) {
	var passport ExecutionPassport
	err := r.db.Where("function_id = ?", functionID).First(&passport).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No passport yet — return default scores (neutral)
			return &DREScores{
				DeterminismScore:          0,
				ReplayIntegrityScore:      0,
				PerformanceStabilityScore: 0,
				DriftScore:                1.0,
			}, nil
		}
		return nil, fmt.Errorf("dre: get dre scores: %w", err)
	}

	return &DREScores{
		DeterminismScore:          passport.DeterminismScore,
		ReplayIntegrityScore:      passport.ReplayIntegrityScore,
		PerformanceStabilityScore: passport.PerformanceStabilityScore,
		DriftScore:                passport.DriftScore,
	}, nil
}

// GetPassportByFunctionID retrieves the execution passport for a function.
// Returns nil if no passport exists yet.
func (r *RegistryRepository) GetPassportByFunctionID(functionID uuid.UUID) (*ExecutionPassport, error) {
	var passport ExecutionPassport
	err := r.db.Where("function_id = ?", functionID).First(&passport).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("dre: get passport by function id: %w", err)
	}
	return &passport, nil
}

// GetDriftReportsByFunctionID retrieves drift reports for a function (paginated).
func (r *RegistryRepository) GetDriftReportsByFunctionID(functionID uuid.UUID, limit, offset int) ([]*DriftReportRecord, error) {
	var reports []*DriftReportRecord
	err := r.db.Where("function_id = ?", functionID).
		Order("detected_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("dre: get drift reports: %w", err)
	}
	return reports, nil
}

// UpdateTrustScoreV2 updates the DRE v2 trust score fields in registry_function_ratings.
func (r *RegistryRepository) UpdateTrustScoreV2(functionID uuid.UUID, scores *DREScores, trustScoreV2 float64) error {
	now := time.Now()
	if err := r.db.Model(&RegistryFunctionRating{}).
		Where("function_id = ?", functionID).
		Updates(map[string]interface{}{
			"determinism_score":             scores.DeterminismScore,
			"replay_integrity_score":        scores.ReplayIntegrityScore,
			"performance_stability_score":   scores.PerformanceStabilityScore,
			"drift_score":                   scores.DriftScore,
			"trust_score_v2":                trustScoreV2,
			"trust_v2_updated_at":           now,
		}).Error; err != nil {
		return fmt.Errorf("dre: update trust score v2: %w", err)
	}
	return nil
}

// UpsertPassport creates or updates an execution passport using ON CONFLICT.
func (r *RegistryRepository) UpsertPassport(passport *ExecutionPassport) error {
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "function_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"deterministic_reliability",
			"replay_drift_incidents",
			"verified_executions_total",
			"total_executions",
			"determinism_score",
			"replay_integrity_score",
			"performance_stability_score",
			"drift_score",
			"capsule_versions_used",
			"last_verified_at",
			"updated_at",
		}),
	}).Create(passport).Error; err != nil {
		return fmt.Errorf("dre: upsert passport: %w", err)
	}
	return nil
}

// ============================================
// DRE Score Computation Helpers
// ============================================

// computeDeterminismScore computes: verified_executions / total_executions
func computeDeterminismScore(verified, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(verified) / float64(total)
}

// computeReplayIntegrityScore computes: 1 - (drift_incidents / verified_executions)
func computeReplayIntegrityScore(driftIncidents int, verifiedExecutions int64) float64 {
	if verifiedExecutions == 0 {
		return 0
	}
	score := 1.0 - float64(driftIncidents)/float64(verifiedExecutions)
	if score < 0 {
		return 0
	}
	return score
}

// computeDriftScore computes: exp(-drift_incidents * 0.1)
// Returns 1.0 for zero drift incidents, decays exponentially with each incident.
func computeDriftScore(driftIncidents int) float64 {
	return math.Exp(-float64(driftIncidents) * 0.1)
}
