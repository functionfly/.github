package execution

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BootstrapFXCERT creates an initial MEG record and FXCERT for a function at publish time
// (e.g. for functionfly-authored functions) so that History and Certificates tabs show
// a certificate without requiring a real execution. Uses empty input {} and output {}.
// If nodeKey is non-nil, the cert is signed (Node Key ID in UI); if platformKey is non-nil, Platform Key ID is set.
// Call from a goroutine so publish response is not blocked.
func BootstrapFXCERT(
	repo *registry.RegistryRepository,
	fn *registry.RegistryFunction,
	fnVersion *registry.RegistryFunctionVersion,
	nodeID, region string,
	nodeKey ed25519.PrivateKey,
	platformKey ed25519.PrivateKey,
) {
	if repo == nil || fn == nil || fnVersion == nil {
		return
	}

	emptyInput := json.RawMessage(`{}`)
	emptyOutput := json.RawMessage(`{}`)

	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	execMeta := ExecutionMetadata{
		ExecutionID:     uuid.New().String(),
		FunctionID:      fn.ID.String(),
		OwnerID:         "",
		CallerID:        "",
		NodeID:          nodeID,
		Region:          region,
		Nonce:           nonce,
		ProtocolVersion: "dre/1.0",
	}
	if fn.OwnerUserID != nil {
		execMeta.OwnerID = fn.OwnerUserID.String()
	}

	capsuleDesc := capsule.Default(execMeta.ExecutionID, "", "")

	megResult, err := BuildMEGFromExecution(fnVersion, emptyInput, emptyOutput, nil, capsuleDesc, execMeta)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fn.ID,
			"version":     fnVersion.Version,
		}).Warn("DRE: Bootstrap FXCERT failed to build MEG")
		return
	}

	capsuleHash, err := capsuleDesc.Hash()
	if err != nil {
		capsuleHash = ""
	}

	execUUID, err := uuid.Parse(execMeta.ExecutionID)
	if err != nil || execUUID == uuid.Nil {
		execUUID = uuid.New()
	}

	megRecord := &registry.MEGRecord{
		ID:                    uuid.New(),
		ExecutionID:           execUUID,
		FunctionID:            fn.ID,
		Version:               fnVersion.Version,
		ExecutionRootHash:     megResult.ExecutionRootHash,
		InputHash:             megResult.InputHash,
		EnvironmentHash:       megResult.EnvironmentHash,
		DependencyHash:        megResult.DependencyHash,
		TraceHash:             megResult.TraceHash,
		ResourceHash:          megResult.ResourceHash,
		OutputHash:            megResult.OutputHash,
		MetadataHash:          megResult.MetadataHash,
		CapsuleDescriptorHash: capsuleHash,
		DeterminismTier:       capsuleDesc.DeterminismTier,
		ProtocolVersion:       "dre/1.0",
	}

	if err := repo.StoreMEGRecord(megRecord); err != nil {
		logrus.WithError(err).WithField("function_id", fn.ID).Warn("DRE: Bootstrap FXCERT failed to store MEG")
		return
	}

	certExec := drecert.ExecutionSection{
		ExecutionID:      execMeta.ExecutionID,
		FunctionID:       fmt.Sprintf("fx://%s/%s/%s", fn.Author, fn.Name, fnVersion.Version),
		OwnerID:          execMeta.OwnerID,
		CallerID:         execMeta.CallerID,
		NodeID:           nodeID,
		Region:           region,
		TimestampVirtual: capsuleDesc.TimeSeed,
		TimestampRealUTC: time.Now().UTC().Format(time.RFC3339),
		ProtocolVersion:  "dre/1.0",
	}

	certCapsule := drecert.CapsuleSection{
		CapsuleDescriptorHash: capsuleHash,
		DeterminismTier:       capsuleDesc.DeterminismTier,
		ProtocolVersion:       capsuleDesc.ProtocolVersion,
	}

	certTrust := drecert.TrustSection{
		TrustScore:       0,
		DeterminismScore: 0,
	}

	cert, err := drecert.Generate(megResult, certExec, certCapsule, certTrust, drecert.CertLevelStandard, nodeKey, platformKey)
	if err != nil {
		logrus.WithError(err).Warn("DRE: Bootstrap FXCERT failed to generate cert")
		return
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		logrus.WithError(err).Warn("DRE: Bootstrap FXCERT failed to marshal cert")
		return
	}

	execCert := &registry.ExecutionCertificate{
		ID:                uuid.New(),
		CertificateID:     cert.CertificateID,
		ExecutionID:       megRecord.ID, // Use MEG record ID as proxy for execution ID
		MEGRecordID:       megRecord.ID,
		FunctionID:        fn.ID,
		CertLevel:         string(drecert.CertLevelStandard),
		CertJSON:          certJSON,
		ExecutionRootHash: megResult.ExecutionRootHash,
		CertificateHash:   cert.Integrity.CertificateHash,
	}

	if err := repo.StoreCertificate(execCert); err != nil {
		logrus.WithError(err).WithField("certificate_id", cert.CertificateID).Warn("DRE: Bootstrap FXCERT failed to store cert")
		return
	}

	now := time.Now()
	_ = repo.UpdatePassport(fn.ID, registry.PassportUpdate{
		IncrementTotal:        true,
		IncrementVerified:     false,
		CapsuleDescriptorHash: capsuleHash,
		LastVerifiedAt:        &now,
		ResourceHash:          megResult.ResourceHash,
	})

	logrus.WithFields(logrus.Fields{
		"function":       fmt.Sprintf("%s/%s", fn.Author, fn.Name),
		"version":        fnVersion.Version,
		"certificate_id": cert.CertificateID,
	}).Info("DRE: Bootstrap FXCERT created after publish")
}
