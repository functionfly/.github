package recommendations

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RecordInteraction records a user interaction for future recommendations
func (s *Service) RecordInteraction(ctx context.Context, userID *uuid.UUID, functionID uuid.UUID, interactionType InteractionType, sessionID *string, referrerFunctionID *uuid.UUID) error {
	interaction := &UserFunctionInteraction{
		ID:                 uuid.New(),
		UserID:             userID,
		FunctionID:         functionID,
		InteractionType:    interactionType,
		SessionID:          sessionID,
		ReferrerFunctionID: referrerFunctionID,
		Timestamp:          time.Now(),
	}

	if err := s.repo.RecordUserInteraction(ctx, interaction); err != nil {
		return err
	}

	// If it's an execute interaction, also track co-occurrences
	if interactionType == InteractionTypeExecute && sessionID != nil {
		s.trackCooccurrences(ctx, *sessionID, functionID)
	}

	return nil
}

// RecordExecution records a function execution for recommendations
func (s *Service) RecordExecution(ctx context.Context, userID *uuid.UUID, functionID uuid.UUID, sessionID string) error {
	// Record session usage
	usage := &SessionFunctionUsage{
		ID:             uuid.New(),
		SessionID:      sessionID,
		FunctionID:     functionID,
		UserID:         userID,
		ExecutionCount: 1,
		FirstUsedAt:    time.Now(),
		LastUsedAt:     time.Now(),
	}

	if err := s.repo.RecordSessionUsage(ctx, usage); err != nil {
		logrus.WithError(err).Warn("failed to record session usage")
	}

	// Track co-occurrences
	s.trackCooccurrences(ctx, sessionID, functionID)

	// Record interaction
	return s.RecordInteraction(ctx, userID, functionID, InteractionTypeExecute, &sessionID, nil)
}

// RecordFeedback records recommendation feedback
func (s *Service) RecordFeedback(ctx context.Context, userID *uuid.UUID, functionID, recommendedFunctionID uuid.UUID, feedbackType string, recommendationType *string) error {
	feedback := &RecommendationFeedback{
		ID:                    uuid.New(),
		UserID:                userID,
		FunctionID:            functionID,
		RecommendedFunctionID: recommendedFunctionID,
		FeedbackType:          feedbackType,
		RecommendationType:    recommendationType,
		Timestamp:             time.Now(),
	}

	return s.repo.RecordFeedback(ctx, feedback)
}
