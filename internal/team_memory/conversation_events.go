package team_memory

import (
	"context"
	"sync"

	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ConversationEventHandler handles conversation lifecycle events
type ConversationEventHandler struct {
	autoUpdater *AutoUpdater
	enabled     bool
}

// NewConversationEventHandler creates a new event handler for conversation events
func NewConversationEventHandler(autoUpdater *AutoUpdater) *ConversationEventHandler {
	return &ConversationEventHandler{
		autoUpdater: autoUpdater,
		enabled:     true,
	}
}

// SetEnabled enables or disables event processing
func (h *ConversationEventHandler) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// OnConversationResolved is called when a conversation is marked as resolved
// This triggers the memory extraction process asynchronously
func (h *ConversationEventHandler) OnConversationResolved(ctx context.Context, conv *conversations.Conversation) error {
	if !h.enabled || h.autoUpdater == nil {
		return nil
	}

	// Only process team conversations that have an organization ID
	if conv.OrganizationID == nil {
		logrus.Debug("Skipping memory extraction: conversation has no organization")
		return nil
	}

	// Use organization_id as team_id for processing
	teamID := *conv.OrganizationID

	logrus.WithFields(logrus.Fields{
		"conversation_id": conv.ID,
		"team_id":         teamID,
		"type":            conv.Type,
	}).Info("Triggering memory extraction for resolved conversation")

	// Process the conversation for memory extraction
	// Use background context to ensure completion even if request context is cancelled
	// The auto-updater handles actual extraction asynchronously via goroutines
	bgCtx := context.WithoutCancel(ctx)
	err := h.autoUpdater.ProcessConversation(bgCtx, conv.ID, teamID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"conversation_id": conv.ID,
			"team_id":         teamID,
		}).Error("Failed to process conversation for memory extraction")
		return err
	}

	return nil
}

// OnConversationUpdated is called when a conversation is updated
// Can be used for other lifecycle events in the future
func (h *ConversationEventHandler) OnConversationUpdated(ctx context.Context, conv *conversations.Conversation) error {
	// Currently no-op, but can be extended for other events
	return nil
}

// ConversationEventPublisher defines the interface for publishing conversation events
type ConversationEventPublisher interface {
	PublishResolved(ctx context.Context, conv *conversations.Conversation) error
}

// SimpleEventPublisher is a simple implementation of the event publisher
type SimpleEventPublisher struct {
	mu       sync.RWMutex
	handlers []ConversationEventHandlerInterface
}

// ConversationEventHandlerInterface is the interface for event handlers
type ConversationEventHandlerInterface interface {
	OnConversationResolved(ctx context.Context, conv *conversations.Conversation) error
}

// NewSimpleEventPublisher creates a new simple event publisher
func NewSimpleEventPublisher() *SimpleEventPublisher {
	return &SimpleEventPublisher{
		handlers: make([]ConversationEventHandlerInterface, 0),
	}
}

// RegisterHandler registers an event handler (thread-safe)
func (p *SimpleEventPublisher) RegisterHandler(handler ConversationEventHandlerInterface) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// PublishResolved publishes a conversation resolved event to all handlers (thread-safe)
func (p *SimpleEventPublisher) PublishResolved(ctx context.Context, conv *conversations.Conversation) error {
	p.mu.RLock()
	handlers := make([]ConversationEventHandlerInterface, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.OnConversationResolved(ctx, conv); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"conversation_id": conv.ID,
				"handler_type":    "team_memory",
			}).Warn("Event handler failed for conversation resolved")
		}
	}
	return nil
}

// DefaultEventPublisher is the global event publisher instance
var DefaultEventPublisher = NewSimpleEventPublisher()

// RegisterTeamMemoryHandler registers the team memory handler with the default publisher
func RegisterTeamMemoryHandler(autoUpdater *AutoUpdater) {
	handler := NewConversationEventHandler(autoUpdater)
	DefaultEventPublisher.RegisterHandler(handler)
}

// PublishConversationResolved publishes a conversation resolved event
func PublishConversationResolved(ctx context.Context, conv *conversations.Conversation) {
	if err := DefaultEventPublisher.PublishResolved(ctx, conv); err != nil {
		logrus.WithError(err).Warn("Failed to publish conversation resolved event")
	}
}

// GetTeamIDFromConversation extracts the team ID from a conversation
// Uses OrganizationID as the team ID for now
func GetTeamIDFromConversation(conv *conversations.Conversation) (uuid.UUID, bool) {
	if conv.OrganizationID == nil {
		return uuid.Nil, false
	}
	return *conv.OrganizationID, true
}
