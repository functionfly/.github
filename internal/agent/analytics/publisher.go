package analytics

import (
	"context"
	"encoding/json"
	"os"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

const (
	SubjectSignalCreated  = "events.brain.signals.created"
	SubjectBrainQueries   = "events.brain.queries"
	SubjectBrainFeedback  = "events.brain.feedback"
	SubjectConnectorSync  = "events.connectors.sync"
	SubjectConnectorLinked = "events.connectors.linked"
)

type Publisher struct {
	nc        *nats.Conn
	anonymizer *Anonymizer
	logger    *logrus.Logger
}

func NewPublisher(logger *logrus.Logger) (*Publisher, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL,
		nats.Name("Brain Analytics Publisher"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(5),
	)
	if err != nil {
		// NATS not available — analytics is best-effort
		logger.WithError(err).Warn("NATS not available, analytics publishing disabled")
		return &Publisher{
			anonymizer: NewAnonymizer(),
			logger:     logger,
		}, nil
	}

	return &Publisher{
		nc:         nc,
		anonymizer: NewAnonymizer(),
		logger:     logger,
	}, nil
}

func (p *Publisher) PublishSignalEvent(ctx context.Context, signal *storage.BrainSignal, tenantTier string) {
	if p.nc == nil {
		return
	}

	event := p.anonymizer.AnonymizeSignal(signal, tenantTier)
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	if err := p.nc.Publish(SubjectSignalCreated, data); err != nil {
		p.logger.WithError(err).Debug("Failed to publish signal event")
	}
}

func (p *Publisher) PublishQueryEvent(ctx context.Context, tenantTier string, connectorSlug string, signalsCount int) {
	if p.nc == nil {
		return
	}

	event := p.anonymizer.AnonymizeQuery(tenantTier, connectorSlug, signalsCount)
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	if err := p.nc.Publish(SubjectBrainQueries, data); err != nil {
		p.logger.WithError(err).Debug("Failed to publish query event")
	}
}

func (p *Publisher) PublishFeedbackEvent(ctx context.Context, tenantTier string, signalType string, helpful bool) {
	if p.nc == nil {
		return
	}

	event := p.anonymizer.AnonymizeFeedback(tenantTier, signalType, helpful)
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	if err := p.nc.Publish(SubjectBrainFeedback, data); err != nil {
		p.logger.WithError(err).Debug("Failed to publish feedback event")
	}
}

func (p *Publisher) PublishConnectorEvent(ctx context.Context, eventType string, tenantTier string, connectorSlug string) {
	if p.nc == nil {
		return
	}

	event := p.anonymizer.AnonymizeConnectorEvent(eventType, tenantTier, connectorSlug)
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	subject := SubjectConnectorSync
	if eventType == "connectors.linked" {
		subject = SubjectConnectorLinked
	}

	if err := p.nc.Publish(subject, data); err != nil {
		p.logger.WithError(err).Debug("Failed to publish connector event")
	}
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Drain()
	}
}
