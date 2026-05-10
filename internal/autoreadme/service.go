package autoreadme

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var defaultRepoRoot = "."

func SetRepoRoot(root string) {
	defaultRepoRoot = root
}

// Service generates and persists auto-generated READMEs for registry functions.
type Service struct {
	repo      *registry.RegistryRepository
	generator *Generator
	repoRoot  string
}

// NewService creates a new auto-README service.
func NewService(repo *registry.RegistryRepository, baseURL string) *Service {
	return &Service{
		repo:      repo,
		generator: NewGenerator(baseURL),
		repoRoot:  defaultRepoRoot,
	}
}

// NewServiceWithRoot creates a new auto-README service with a custom repo root.
func NewServiceWithRoot(repo *registry.RegistryRepository, baseURL, repoRoot string) *Service {
	return &Service{
		repo:      repo,
		generator: NewGenerator(baseURL),
		repoRoot:  repoRoot,
	}
}

// GenerateForVersion creates and persists a README for a specific function version.
// If the version already has a non-empty readme, it is skipped unless force=true.
func (s *Service) GenerateForVersion(ctx context.Context, functionID uuid.UUID, version string, force bool) (string, error) {
	fn, err := s.repo.GetFunctionByID(functionID)
	if err != nil {
		return "", fmt.Errorf("get function: %w", err)
	}

	fnVersion, err := s.repo.GetFunctionVersion(functionID, version)
	if err != nil {
		return "", fmt.Errorf("get version: %w", err)
	}

	// Skip if already has readme and not forced
	if !force && fnVersion.Readme.Valid && fnVersion.Readme.String != "" {
		return fnVersion.Readme.String, nil
	}

	meta, err := s.buildMeta(fn, fnVersion)
	if err != nil {
		return "", fmt.Errorf("build meta: %w", err)
	}

	readme := s.generator.Generate(meta)

	// Persist
	fnVersion.Readme = sql.NullString{String: readme, Valid: true}
	if err := s.repo.UpdateFunctionVersionField(fnVersion.ID, "readme", readme); err != nil {
		return "", fmt.Errorf("persist readme: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"version":     version,
		"force":       force,
		"length":      len(readme),
	}).Info("Auto-generated README for function")

	return readme, nil
}

// GenerateForFunction generates readmes for all versions of a function.
func (s *Service) GenerateForFunction(ctx context.Context, functionID uuid.UUID, force bool) (int, error) {
	versions, err := s.repo.ListFunctionVersions(functionID)
	if err != nil {
		return 0, fmt.Errorf("list versions: %w", err)
	}

	count := 0
	for _, v := range versions {
		_, err := s.GenerateForVersion(ctx, functionID, v.Version, force)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"function_id": functionID,
				"version":     v.Version,
			}).Warn("Failed to auto-generate README for version")
			continue
		}
		count++
	}
	return count, nil
}

// BackfillAll generates readmes for all functions that are missing them.
func (s *Service) BackfillAll(ctx context.Context, batchSize int, force bool) (int, error) {
	functions, _, err := s.repo.ListFunctionsForAdmin("", "", "", batchSize, 0)
	if err != nil {
		return 0, fmt.Errorf("list functions: %w", err)
	}

	total := 0
	for _, fn := range functions {
		count, err := s.GenerateForFunction(ctx, fn.ID, force)
		if err != nil {
			logrus.WithError(err).WithField("function_id", fn.ID).Warn("Backfill failed for function")
			continue
		}
		total += count
	}

	logrus.WithField("total_versions", total).Info("README backfill complete")
	return total, nil
}

// GenerateProjectReadme generates a README for the entire project.
func (s *Service) GenerateProjectReadme() string {
	if s.repoRoot == "" {
		s.repoRoot = defaultRepoRoot
	}
	return GenerateProjectReadmeFromDir(s.repoRoot)
}

// GenerateProjectReadmeToFile generates and writes the project README to disk.
func (s *Service) GenerateProjectReadmeToFile(path string) error {
	readme := s.GenerateProjectReadme()
	return os.WriteFile(path, []byte(readme), 0644)
}

// GetProjectContext returns the detected project context.
func (s *Service) GetProjectContext() ProjectContext {
	if s.repoRoot == "" {
		s.repoRoot = defaultRepoRoot
	}
	return GenerateProjectContext(s.repoRoot)
}

func (s *Service) buildMeta(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion) (FunctionMeta, error) {
	meta := FunctionMeta{
		Author:        fn.Author,
		Name:          fn.Name,
		Version:       fnVersion.Version,
		Runtime:       fnVersion.Runtime,
		TimeoutMs:     fnVersion.TimeoutMs,
		MemoryMB:      fnVersion.MemoryMB,
		Deterministic: fnVersion.Deterministic,
		SideEffects:   fnVersion.SideEffects,
		Idempotent:    fnVersion.Idempotent,
		CacheTTL:      fnVersion.CacheTTL,
		PublishedAt:   fnVersion.PublishedAt,
	}

	if fn.Title.Valid {
		meta.Title = fn.Title.String
	}
	if fn.Description.Valid {
		meta.Description = fn.Description.String
	}
	if fn.Category.Valid {
		meta.Category = fn.Category.String
	}

	tags, _ := registry.ParseTags(fn.Tags)
	meta.Tags = tags

	// Parse manifest for input/output
	var manifest map[string]interface{}
	if err := json.Unmarshal(fnVersion.Manifest, &manifest); err == nil {
		if input, ok := manifest["input"].(map[string]interface{}); ok {
			meta.InputSchema = input
			if ex, ok := input["example"]; ok {
				meta.InputExample = ex
			}
		}
		if output, ok := manifest["output"].(map[string]interface{}); ok {
			meta.OutputSchema = output
			if ex, ok := output["example"]; ok {
				meta.OutputExample = ex
			}
		}
		if title, ok := manifest["title"].(string); ok && title != "" && meta.Title == "" {
			meta.Title = title
		}
		if desc, ok := manifest["description"].(string); ok && desc != "" && meta.Description == "" {
			meta.Description = desc
		}
	}

	// Capabilities
	if len(fnVersion.Capabilities) > 0 {
		var caps []string
		json.Unmarshal(fnVersion.Capabilities, &caps)
		meta.Capabilities = caps
	}

	// Rating / trust score
	rating, _ := s.repo.GetRatingByFunctionID(fn.ID)
	if rating != nil {
		score := rating.TrustScore
		if score > 0 && score <= 1 {
			score = score * 100
		}
		meta.TrustScore = score
	}

	return meta, nil
}
