package connectors

import (
	"context"
	"fmt"

	extractorPkg "github.com/functionfly/functionfly/internal/agent/connectors/extractors"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type SignalExtractor interface {
	Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error)
	SupportedSignalTypes() []string
	ConnectorSlug() string
}

type Registry struct {
	extractors map[string]SignalExtractor
}

func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]SignalExtractor),
	}
}

func (r *Registry) Register(extractor SignalExtractor) {
	r.extractors[extractor.ConnectorSlug()] = extractor
}

func (r *Registry) Get(slug string) (SignalExtractor, bool) {
	ext, ok := r.extractors[slug]
	return ext, ok
}

func (r *Registry) SupportedSlugs() []string {
	slugs := make([]string, 0, len(r.extractors))
	for slug := range r.extractors {
		slugs = append(slugs, slug)
	}
	return slugs
}

// ExtractSignals runs the appropriate extractor for a connector
func (r *Registry) ExtractSignals(ctx context.Context, connectorSlug string, tenantID uuid.UUID, rawData []byte) ([]*storage.BrainSignal, error) {
	ext, ok := r.extractors[connectorSlug]
	if !ok {
		return nil, fmt.Errorf("no extractor registered for connector: %s", connectorSlug)
	}

	signals, err := ext.Extract(ctx, rawData)
	if err != nil {
		return nil, fmt.Errorf("extract signals for %s: %w", connectorSlug, err)
	}

	// Set tenant ID on all signals
	for _, s := range signals {
		s.TenantID = tenantID
	}

	return signals, nil
}

// DefaultRegistry creates a registry with all built-in extractors
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(extractorPkg.NewGitHubExtractor())
	r.Register(extractorPkg.NewNotionExtractor())
	r.Register(extractorPkg.NewSlackExtractor())
	r.Register(extractorPkg.NewGmailExtractor())
	r.Register(extractorPkg.NewLinearExtractor())
	return r
}
