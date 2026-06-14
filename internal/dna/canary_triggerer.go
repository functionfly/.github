package dna

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/sirupsen/logrus"
)

// CanaryTriggerer triggers a canary deployment for a function version.
type CanaryTriggerer interface {
	TriggerCanary(ctx context.Context, functionID string, version string, trafficPercent int) error
}

// RegistryCanaryTriggerer implements CanaryTriggerer using the canary config repository.
type RegistryCanaryTriggerer struct {
	canaryRepo   *registry.CanaryConfigRepository
	functionRepo *registry.RegistryRepository
	logger       *logrus.Logger
}

// NewRegistryCanaryTriggerer creates a new RegistryCanaryTriggerer.
func NewRegistryCanaryTriggerer(
	canaryRepo *registry.CanaryConfigRepository,
	functionRepo *registry.RegistryRepository,
	logger *logrus.Logger,
) *RegistryCanaryTriggerer {
	return &RegistryCanaryTriggerer{
		canaryRepo:   canaryRepo,
		functionRepo: functionRepo,
		logger:       logger,
	}
}

// TriggerCanary creates a canary deployment for the given function version.
func (t *RegistryCanaryTriggerer) TriggerCanary(ctx context.Context, functionID string, version string, trafficPercent int) error {
	fnUUID, err := uuid.Parse(functionID)
	if err != nil {
		return fmt.Errorf("invalid function ID: %w", err)
	}

	// Verify function exists
	_, err = t.functionRepo.GetFunctionByID(ctx, fnUUID)
	if err != nil {
		return fmt.Errorf("function not found: %w", err)
	}

	// Cancel any existing active canary before creating new one
	existing, _ := t.canaryRepo.GetByFunctionID(fnUUID)
	if existing != nil {
		if err := t.canaryRepo.UpdateStatus(existing.ID, "cancelled"); err != nil {
			t.logger.WithError(err).Warn("dna: failed to cancel existing canary before creating new one")
		}
	}

	if trafficPercent <= 0 || trafficPercent > 100 {
		trafficPercent = 10
	}

	canary := &registry.CanaryConfig{
		FunctionID:       fnUUID,
		Version:          version,
		TrafficPercent:   trafficPercent,
		AutoPromote:      true,
		PromoteThreshold: 0.05, // 5% error rate threshold (lenient for DNA mutations)
		PromoteWindow:    600,  // 10 minutes
		Status:           "active",
	}

	if err := t.canaryRepo.Create(canary); err != nil {
		return fmt.Errorf("create canary: %w", err)
	}

	t.logger.WithFields(logrus.Fields{
		"function_id":     functionID,
		"version":         version,
		"traffic_percent": trafficPercent,
		"canary_id":       canary.ID,
	}).Info("dna: canary deployment triggered for mutation acceptance")

	return nil
}
