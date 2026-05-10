package timemachine

import (
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type AuditGenerator struct {
	tmRepo     *tmstorage.Repository
	signingKey ed25519.PrivateKey
	keyID      string
}

func NewAuditGenerator(tmRepo *tmstorage.Repository, signingKey ed25519.PrivateKey, keyID string) *AuditGenerator {
	return &AuditGenerator{
		tmRepo:     tmRepo,
		signingKey: signingKey,
		keyID:      keyID,
	}
}

type AuditCertificateJSON struct {
	Version        string                 `json:"version"`
	CertificateID  string                 `json:"certificate_id"`
	ReplayID       string                 `json:"replay_id"`
	TenantID       string                 `json:"tenant_id"`
	FunctionID     string                 `json:"function_id"`
	GeneratedAt    time.Time              `json:"generated_at"`
	ReplayWindow   ReplayWindowJSON       `json:"replay_window"`
	Summary        AuditSummaryJSON       `json:"summary"`
	Items          []AuditItemJSON        `json:"items"`
	MerkleRoot     string                 `json:"merkle_root"`
	Signature      string                 `json:"signature,omitempty"`
	KeyID          string                 `json:"key_id,omitempty"`
	ComplianceInfo []string               `json:"compliance_frameworks"`
	RetentionPolicy string               `json:"retention_policy"`
	IntegrityProof map[string]interface{} `json:"integrity_proof"`
}

type ReplayWindowJSON struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type AuditSummaryJSON struct {
	TotalExecutions int     `json:"total_executions"`
	Identical       int     `json:"identical"`
	Changed         int     `json:"changed"`
	Failed          int     `json:"failed"`
	Breaking        int     `json:"breaking"`
	SuccessRate     float64 `json:"success_rate"`
	TargetVersion   string  `json:"target_version"`
	Reason          string  `json:"reason"`
	ReconciliationMode string `json:"reconciliation_mode"`
}

type AuditItemJSON struct {
	ExecutionID      string  `json:"execution_id"`
	OriginalVersion  string  `json:"original_version"`
	TargetVersion    string  `json:"target_version"`
	DiffType         string  `json:"diff_type"`
	OriginalDuration int     `json:"original_duration_ms"`
	NewDuration      int     `json:"new_duration_ms"`
	DurationDelta    float64 `json:"duration_delta_percent"`
	ItemHash         string  `json:"item_hash"`
}

func (g *AuditGenerator) Generate(replayID uuid.UUID) (*tmstorage.AuditCertificate, error) {
	replay, items, err := g.tmRepo.GetReplayWithItems(replayID)
	if err != nil {
		return nil, fmt.Errorf("load replay: %w", err)
	}
	if replay == nil {
		return nil, fmt.Errorf("replay not found")
	}
	if replay.Status != "completed" {
		return nil, fmt.Errorf("replay must be completed before generating audit certificate")
	}

	certID := fmt.Sprintf("ATC-%s-%s", replayID.String()[:8], time.Now().UTC().Format("20060102150405"))

	summary := g.buildSummary(replay, items)
	auditItems := g.buildItems(items, replay.TargetVersion)

	itemHashes := make([][]byte, len(auditItems))
	for i, item := range auditItems {
		hash := sha256.Sum256([]byte(item.ItemHash))
		itemHashes[i] = hash[:]
	}
	merkleRoot := buildMerkleTree(itemHashes)

	summaryJSON := g.buildSummaryJSON(replay, summary)
	windowJSON := ReplayWindowJSON{
		Start: replay.WindowStart,
		End:   replay.WindowEnd,
	}

	certJSON := AuditCertificateJSON{
		Version:       "1.0",
		CertificateID: certID,
		ReplayID:      replayID.String(),
		TenantID:      replay.TenantID.String(),
		FunctionID:    replay.FunctionID.String(),
		GeneratedAt:   time.Now().UTC(),
		ReplayWindow:  windowJSON,
		Summary:       summaryJSON,
		Items:         auditItems,
		MerkleRoot:    hex.EncodeToString(merkleRoot),
		ComplianceInfo: []string{"SOC2", "ISO27001", "HIPAA"},
		RetentionPolicy: "7_years",
		IntegrityProof: map[string]interface{}{
			"algorithm":  "sha256-merkle",
			"item_count": len(items),
			"root_hash":  hex.EncodeToString(merkleRoot),
		},
	}

	var signature string
	if g.signingKey != nil {
		certBytes, _ := json.Marshal(certJSON)
		hash := sha256.Sum256(certBytes)
		sig := ed25519.Sign(g.signingKey, hash[:])
		signature = hex.EncodeToString(sig)
		certJSON.Signature = signature
		certJSON.KeyID = g.keyID
	}

	certJSONBytes, err := json.Marshal(certJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal cert json: %w", err)
	}

	certHash := sha256.Sum256(certJSONBytes)

	prevCert, _ := g.tmRepo.GetLatestAuditCertificateForFunction(replay.FunctionID)

	cert := &tmstorage.AuditCertificate{
		ID:                   uuid.New(),
		ReplayID:             replayID,
		CertificateID:        certID,
		CertJSON:             certJSONBytes,
		CertHash:             hex.EncodeToString(certHash[:]),
		ComplianceFrameworks: []string{"SOC2", "ISO27001", "HIPAA"},
		RetentionPolicy:      "7_years",
	}

	if prevCert != nil {
		cert.PreviousCertHash = sql.NullString{String: prevCert.CertHash, Valid: true}
	}

	if signature != "" {
		cert.Signature = sql.NullString{String: signature, Valid: true}
	}

	cert.MerkleRoot = sql.NullString{String: hex.EncodeToString(merkleRoot), Valid: true}

	if err := g.tmRepo.CreateAuditCertificate(cert); err != nil {
		return nil, fmt.Errorf("store certificate: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"replay_id":      replayID,
		"certificate_id": certID,
		"merkle_root":    hex.EncodeToString(merkleRoot)[:16],
		"items":          len(items),
	}).Info("Audit certificate generated")

	return cert, nil
}

type diffSummary struct {
	identical int
	minor     int
	major     int
	breaking  int
	errors    int
}

func (g *AuditGenerator) buildSummary(replay *tmstorage.Replay, items []tmstorage.ReplayItem) diffSummary {
	s := diffSummary{}
	for _, item := range items {
		if item.Status == "failed" {
			s.errors++
			continue
		}
		if !item.DiffType.Valid {
			s.identical++
			continue
		}
		switch item.DiffType.String {
		case "identical":
			s.identical++
		case "minor":
			s.minor++
		case "major":
			s.major++
		case "breaking":
			s.breaking++
		default:
			s.identical++
		}
	}
	return s
}

func (g *AuditGenerator) buildSummaryJSON(replay *tmstorage.Replay, s diffSummary) AuditSummaryJSON {
	total := s.identical + s.minor + s.major + s.breaking + s.errors
	changed := s.minor + s.major + s.breaking
	var successRate float64
	if total > 0 {
		successRate = float64(total-s.errors) / float64(total) * 100
	}
	return AuditSummaryJSON{
		TotalExecutions:    total,
		Identical:          s.identical,
		Changed:            changed,
		Failed:             s.errors,
		Breaking:           s.breaking,
		SuccessRate:        successRate,
		TargetVersion:      replay.TargetVersion,
		Reason:             replay.Reason,
		ReconciliationMode: replay.ReconciliationMode,
	}
}

func (g *AuditGenerator) buildItems(items []tmstorage.ReplayItem, targetVersion string) []AuditItemJSON {
	auditItems := make([]AuditItemJSON, 0, len(items))
	for _, item := range items {
		diffType := "identical"
		if item.DiffType.Valid {
			diffType = item.DiffType.String
		}

		newDuration := 0
		if item.NewDurationMs.Valid {
			newDuration = int(item.NewDurationMs.Int32)
		}

		var delta float64
		if item.OriginalDurationMs > 0 && newDuration > 0 {
			delta = float64(newDuration-item.OriginalDurationMs) / float64(item.OriginalDurationMs) * 100
		}

		itemData := fmt.Sprintf("%s:%s:%s:%s",
			item.OriginalExecutionID,
			item.OriginalVersion,
			diffType,
			string(item.OriginalOutput),
		)
		itemHash := sha256.Sum256([]byte(itemData))

		auditItems = append(auditItems, AuditItemJSON{
			ExecutionID:      item.OriginalExecutionID.String(),
			OriginalVersion:  item.OriginalVersion,
			TargetVersion:    targetVersion,
			DiffType:         diffType,
			OriginalDuration: item.OriginalDurationMs,
			NewDuration:      newDuration,
			DurationDelta:    delta,
			ItemHash:         hex.EncodeToString(itemHash[:]),
		})
	}
	return auditItems
}

func buildMerkleTree(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		empty := sha256.Sum256([]byte{})
		return empty[:]
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	var nextLevel [][]byte
	for i := 0; i < len(hashes); i += 2 {
		if i+1 < len(hashes) {
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash[:])
		} else {
			combined := append(hashes[i], hashes[i]...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash[:])
		}
	}

	return buildMerkleTree(nextLevel)
}

func nullStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
