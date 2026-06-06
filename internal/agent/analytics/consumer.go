package analytics

import (
	"context"
	"encoding/json"
	"os"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	nc       *nats.Conn
	repo     *storage.AnalyticsEventRepository
	logger   *logrus.Logger
	stopCh   chan struct{}
}

func NewConsumer(repo *storage.AnalyticsEventRepository, logger *logrus.Logger) (*Consumer, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL,
		nats.Name("Brain Analytics Consumer"),
		nats.MaxReconnects(10),
	)
	if err != nil {
		logger.WithError(err).Warn("NATS not available, analytics consumer disabled")
		return &Consumer{
			repo:   repo,
			logger: logger,
			stopCh: make(chan struct{}),
		}, nil
	}

	return &Consumer{
		nc:     nc,
		repo:   repo,
		logger: logger,
		stopCh: make(chan struct{}),
	}, nil
}

// Start subscribes to all brain analytics NATS subjects
func (c *Consumer) Start(ctx context.Context) {
	if c.nc == nil {
		return
	}

	subjects := []string{
		SubjectSignalCreated,
		SubjectBrainQueries,
		SubjectBrainFeedback,
		SubjectConnectorSync,
		SubjectConnectorLinked,
	}

	for _, subject := range subjects {
		subj := subject
		c.nc.Subscribe(subj, func(msg *nats.Msg) {
			var event storage.AnalyticsEvent
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				c.logger.WithError(err).WithField("subject", subj).Debug("Failed to unmarshal analytics event")
				return
			}
			if err := c.repo.SaveEvent(ctx, &event); err != nil {
				c.logger.WithError(err).WithField("subject", subj).Debug("Failed to save analytics event")
			}
		})
	}

	c.logger.Info("Brain analytics consumer started")
}

func (c *Consumer) Stop() {
	close(c.stopCh)
	if c.nc != nil {
		c.nc.Drain()
	}
	c.logger.Info("Brain analytics consumer stopped")
}
