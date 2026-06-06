package analytics

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type Anonymizer struct{}

func NewAnonymizer() *Anonymizer {
	return &Anonymizer{}
}

func (a *Anonymizer) AnonymizeSignal(signal *storage.BrainSignal, tenantTier string) *storage.AnalyticsEvent {
	return &storage.AnalyticsEvent{
		ID:            uuid.New(),
		EventType:     "brain.signal.created",
		TenantTier:    tenantTier,
		ConnectorSlug: signal.ConnectorSlug,
		SignalType:    signal.SignalType,
		Importance:    signal.Importance,
		FactLength:    len(signal.Fact),
		CreatedAt:     time.Now().UTC(),
	}
}

func (a *Anonymizer) AnonymizeQuery(tenantTier string, connectorSlug string, signalsCount int) *storage.AnalyticsEvent {
	return &storage.AnalyticsEvent{
		ID:            uuid.New(),
		EventType:     "brain.query",
		TenantTier:    tenantTier,
		ConnectorSlug: connectorSlug,
		SignalsCount:  signalsCount,
		CreatedAt:     time.Now().UTC(),
	}
}

func (a *Anonymizer) AnonymizeFeedback(tenantTier string, signalType string, helpful bool) *storage.AnalyticsEvent {
	helpfulStr := "false"
	if helpful {
		helpfulStr = "true"
	}
	return &storage.AnalyticsEvent{
		ID:         uuid.New(),
		EventType:  "brain.feedback",
		TenantTier: tenantTier,
		SignalType: signalType,
		Metadata:   []byte(`{"helpful":` + helpfulStr + `}`),
		CreatedAt:  time.Now().UTC(),
	}
}

func (a *Anonymizer) AnonymizeConnectorEvent(eventType string, tenantTier string, connectorSlug string) *storage.AnalyticsEvent {
	return &storage.AnalyticsEvent{
		ID:            uuid.New(),
		EventType:     eventType,
		TenantTier:    tenantTier,
		ConnectorSlug: connectorSlug,
		CreatedAt:     time.Now().UTC(),
	}
}
