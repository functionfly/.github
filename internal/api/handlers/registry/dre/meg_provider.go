// Package dre implements the DRE 2.0 API handlers.
// This file provides storage-backed retrieval of the original MEG for replay certification.
package dre

import (
	"context"

	"github.com/functionfly/functionfly/internal/dre/cert"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// RegistryMEGProvider retrieves the original MEG from registry storage by execution ID.
// It implements cert.OriginalMEGProvider for use with the replay certification service.
type RegistryMEGProvider struct {
	Repo *registry.RegistryRepository
}

// NewRegistryMEGProvider returns an OriginalMEGProvider that uses the registry repository.
func NewRegistryMEGProvider(repo *registry.RegistryRepository) cert.OriginalMEGProvider {
	return &RegistryMEGProvider{Repo: repo}
}

// GetOriginalMEG loads the MEG record for the given execution and returns it as a MEGResult.
func (p *RegistryMEGProvider) GetOriginalMEG(ctx context.Context, executionID string) (*drecrypto.MEGResult, error) {
	id, err := uuid.Parse(executionID)
	if err != nil {
		return nil, err
	}
	rec, err := p.Repo.GetMEGByExecutionID(id)
	if err != nil || rec == nil {
		return nil, err
	}
	return MEGRecordToMEGResult(rec), nil
}

// MEGRecordToMEGResult converts a stored MEG record to a crypto MEGResult for comparison and drift analysis.
func MEGRecordToMEGResult(rec *registry.MEGRecord) *drecrypto.MEGResult {
	if rec == nil {
		return nil
	}
	return &drecrypto.MEGResult{
		ExecutionRootHash: rec.ExecutionRootHash,
		InputHash:         rec.InputHash,
		EnvironmentHash:   rec.EnvironmentHash,
		DependencyHash:    rec.DependencyHash,
		TraceHash:         rec.TraceHash,
		ResourceHash:      rec.ResourceHash,
		OutputHash:        rec.OutputHash,
		MetadataHash:      rec.MetadataHash,
	}
}
