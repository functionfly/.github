package dna

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// RegistrySourceCodeFetcher implements SourceCodeFetcher using the registry repository.
type RegistrySourceCodeFetcher struct {
	repo *registry.RegistryRepository
}

// NewRegistrySourceCodeFetcher creates a new fetcher.
func NewRegistrySourceCodeFetcher(repo *registry.RegistryRepository) *RegistrySourceCodeFetcher {
	return &RegistrySourceCodeFetcher{repo: repo}
}

// GetFunctionSourceCode returns the source code and runtime for the latest version of a function.
func (f *RegistrySourceCodeFetcher) GetFunctionSourceCode(functionID string) (string, string, error) {
	fid, err := uuid.Parse(functionID)
	if err != nil {
		return "", "", fmt.Errorf("invalid function ID: %w", err)
	}

	fn, err := f.repo.GetFunctionByID(fid)
	if err != nil {
		return "", "", fmt.Errorf("get function: %w", err)
	}

	version, err := f.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		return "", "", fmt.Errorf("get latest version: %w", err)
	}

	if !version.SourceCode.Valid || version.SourceCode.String == "" {
		return "", "", nil
	}

	return version.SourceCode.String, version.Runtime, nil
}
