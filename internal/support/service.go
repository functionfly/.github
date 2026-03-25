package support

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service provides support system business logic
type Service struct {
	repo     Repository
	aiClient AIChatClient
	redis    RedisClient
	logger   *logrus.Logger
}

// AIChatClient interface for AI chat functionality
type AIChatClient interface {
	GenerateSupportResponse(ctx context.Context, req *AIRequest) (*AIResponse, error)
}

// RedisClient interface for Redis pub/sub
type RedisClient interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (<-chan string, error)
}

// AIRequest represents a request to the AI chat service
type AIRequest struct {
	ConversationID uuid.UUID
	UserMessage    string
	Context        *SupportContext
	History        []*SupportMessage
}

// AIResponse represents a response from the AI chat service
type AIResponse struct {
	Message    string
	Confidence float64
	Model     string
	Actions   []string
}

// NewService creates a new support service
func NewService(repo Repository, aiClient AIChatClient, redis RedisClient, logger *logrus.Logger) *Service {
	if logger == nil {
		logger = logrus.New()
	}
	return &Service{
		repo:     repo,
		aiClient: aiClient,
		redis:    redis,
		logger:   logger,
	}
}

// CreateConversation creates a new support conversation
func (s *Service) CreateConversation(ctx context.Context, userID uuid.UUID, req *CreateConversationRequest) (*SupportConversation, error) {
	conversation := &SupportConversation{
		UserID:   userID,
		Type:     req.Type,
		Status:   StatusActive,
		Priority: req.Priority,
		Title:    req.Title,
	}

	if req.FunctionRef != nil {
		data, _ := json.Marshal(req.FunctionRef)
		conversation.FunctionRefJSON = data
		conversation.FunctionRef = req.FunctionRef
	}

	if req.DeploymentID != nil {
		conversation.DeploymentID = req.DeploymentID
	}

	if req.IsEmergency {
		conversation.IsEmergency = true
		conversation.Priority = PriorityCritical
		conversation.Status = StatusPending
	}

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	// Publish event for real-time updates
	s.publishEvent(ctx, "conversation.created", conversation)

	return conversation, nil
}

// CreateConversationRequest is the request to create a conversation
type CreateConversationRequest struct {
	Type             ConversationType
	Priority         Priority
	Title           string
	FunctionRef     *FunctionRef
	DeploymentID    *uuid.UUID
	DeploymentLogs  string
	DeploymentError string
	IsEmergency     bool
	Metadata        map[string]interface{}
}

// GetConversation retrieves a support conversation
func (s *Service) GetConversation(ctx context.Context, id uuid.UUID) (*SupportConversation, error) {
	return s.repo.GetConversation(ctx, id)
}

// ListConversations lists support conversations for a user
func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*SupportConversation, error) {
	return s.repo.ListConversations(ctx, userID, limit, offset)
}

// ListActiveConversations lists all active conversations (for staff)
func (s *Service) ListActiveConversations(ctx context.Context, limit, offset int) ([]*SupportConversation, error) {
	return s.repo.ListActiveConversations(ctx, limit, offset)
}

// SendMessage sends a message in a support conversation
func (s *Service) SendMessage(ctx context.Context, conversationID, authorID uuid.UUID, authorType AuthorType, content string) (*SupportMessage, error) {
	// Verify conversation exists and user has access
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil || conversation == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Create the message
	message := &SupportMessage{
		ConversationID: conversationID,
		AuthorID:       authorID,
		AuthorType:    authorType,
		MessageType:   TypeMessage,
		Content:        content,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Publish for real-time
	s.publishEvent(ctx, "message.created", message)

	// If this is a user message and AI is handling, get AI response
	if authorType == AuthorUser && conversation.AIHandled && conversation.StaffID == nil {
		go s.handleAIMessage(context.Background(), conversation, message)
	}

	return message, nil
}

// SendAIResponse sends an AI response to a conversation
func (s *Service) SendAIResponse(ctx context.Context, conversationID uuid.UUID, content string, confidence float64, model string) (*SupportMessage, error) {
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil || conversation == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	message := &SupportMessage{
		ConversationID: conversationID,
		AuthorID:       uuid.Nil, // AI has no user ID
		AuthorType:    AuthorAI,
		MessageType:   TypeAIResponse,
		Content:        content,
		AIConfidence:   &confidence,
		AIModel:       model,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Increment AI attempts
	s.repo.IncrementAIAttempts(ctx, conversationID)

	// Publish for real-time
	s.publishEvent(ctx, "message.created", message)

	// If AI has tried too many times without resolution, suggest escalation
	if conversation.AIAttempts >= 3 {
		s.publishEvent(ctx, "conversation.suggest_escalation", map[string]interface{}{
			"conversation_id": conversationID,
			"reason":          "AI resolution attempts exceeded",
		})
	}

	return message, nil
}

// EscalateToHuman escalates a conversation to human support
func (s *Service) EscalateToHuman(ctx context.Context, conversationID uuid.UUID) error {
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil || conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	// Find available staff
	staff, err := s.repo.ListOnlineStaff(ctx)
	if err != nil || len(staff) == 0 {
		// No staff available, keep as pending
		s.repo.UpdateConversationStatus(ctx, conversationID, StatusPending)
		s.publishEvent(ctx, "conversation.no_staff_available", map[string]interface{}{
			"conversation_id": conversationID,
		})
		return nil
	}

	// Assign to first available staff member
	staffMember := staff[0]
	if err := s.repo.UpdateConversationStaff(ctx, conversationID, staffMember.StaffID); err != nil {
		return fmt.Errorf("assign staff: %w", err)
	}

	// Increment staff active chats
	s.repo.IncrementActiveChats(ctx, staffMember.StaffID)

	// Add staff as participant
	s.repo.AddParticipant(ctx, &SupportConversationParticipant{
		ConversationID: conversationID,
		UserID:         staffMember.StaffID,
		Role:           "helper",
		JoinedAt:       time.Now(),
	})

	// Update status
	s.repo.UpdateConversationStatus(ctx, conversationID, StatusActive)

	// Send system message
	s.sendSystemMessage(ctx, conversationID, "A support staff member has joined the conversation.")

	// Notify real-time
	s.publishEvent(ctx, "conversation.escalated", map[string]interface{}{
		"conversation_id": conversationID,
		"staff_id":        staffMember.StaffID,
	})

	return nil
}

// ResolveConversation resolves a support conversation
func (s *Service) ResolveConversation(ctx context.Context, conversationID, resolvedBy uuid.UUID, note string) error {
	conversation, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil || conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	// Decrement staff active chats if staff was assigned
	if conversation.StaffID != nil {
		s.repo.DecrementActiveChats(ctx, *conversation.StaffID)
	}

	if err := s.repo.ResolveConversation(ctx, conversationID, resolvedBy, note); err != nil {
		return fmt.Errorf("resolve conversation: %w", err)
	}

	// Send system message
	s.sendSystemMessage(ctx, conversationID, "This conversation has been resolved.")

	// Publish event
	s.publishEvent(ctx, "conversation.resolved", map[string]interface{}{
		"conversation_id": conversationID,
		"resolved_by":     resolvedBy,
	})

	return nil
}

// GetMessages retrieves messages for a conversation
func (s *Service) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*SupportMessage, error) {
	return s.repo.ListMessages(ctx, conversationID, limit, offset)
}

// EmergencyFixRequestInput represents a request for emergency fix
type EmergencyFixRequestInput struct {
	ConversationID uuid.UUID
	UserID        uuid.UUID
	FunctionID    uuid.UUID
	Reason        string
}

// CreateEmergencyFixRequest creates an emergency fix request
func (s *Service) CreateEmergencyFixRequest(ctx context.Context, req *EmergencyFixRequestInput) (*EmergencyFixRequest, error) {
	// Create the emergency request
	emergency := &EmergencyFixRequest{
		ConversationID: req.ConversationID,
		UserID:         req.UserID,
		FunctionID:     req.FunctionID,
		Reason:         req.Reason,
		Status:         "pending",
	}

	if err := s.repo.CreateEmergencyRequest(ctx, emergency); err != nil {
		return nil, fmt.Errorf("create emergency request: %w", err)
	}

	// Also update the conversation to emergency status
	conversation, _ := s.repo.GetConversation(ctx, req.ConversationID)
	if conversation != nil {
		conversation.IsEmergency = true
		conversation.Priority = PriorityCritical
		s.repo.UpdateConversationStatus(ctx, req.ConversationID, StatusPending)
	}

	// Publish to emergency channel
	s.publishEvent(ctx, "emergency.created", emergency)

	return emergency, nil
}

// ListPendingEmergencies lists all pending emergency requests
func (s *Service) ListPendingEmergencies(ctx context.Context) ([]*EmergencyFixRequest, error) {
	return s.repo.ListPendingEmergencies(ctx)
}

// AcceptEmergency allows a staff member to accept an emergency
func (s *Service) AcceptEmergency(ctx context.Context, emergencyID, staffID uuid.UUID) error {
	emergency, err := s.repo.GetEmergencyRequest(ctx, emergencyID)
	if err != nil || emergency == nil {
		return fmt.Errorf("emergency request not found")
	}

	if err := s.repo.UpdateEmergencyStatus(ctx, emergencyID, staffID, "accepted"); err != nil {
		return fmt.Errorf("accept emergency: %w", err)
	}

	// Also update the conversation staff
	if err := s.repo.UpdateConversationStaff(ctx, emergency.ConversationID, staffID); err != nil {
		s.logger.WithError(err).Warn("failed to update conversation staff")
	}

	// Publish event
	s.publishEvent(ctx, "emergency.accepted", map[string]interface{}{
		"emergency_id": emergencyID,
		"staff_id":     staffID,
	})

	return nil
}

// SetStaffOnline sets a staff member's online status
func (s *Service) SetStaffOnline(ctx context.Context, staffID uuid.UUID, online bool) error {
	return s.repo.SetStaffOnline(ctx, staffID, online)
}

// GetStaffAvailability gets a staff member's availability
func (s *Service) GetStaffAvailability(ctx context.Context, staffID uuid.UUID) (*StaffAvailability, error) {
	return s.repo.GetStaffAvailability(ctx, staffID)
}

// ListOnlineStaff lists all online staff
func (s *Service) ListOnlineStaff(ctx context.Context) ([]*StaffAvailability, error) {
	return s.repo.ListOnlineStaff(ctx)
}

// GetSupportContext gathers contextual information for support
func (s *Service) GetSupportContext(ctx context.Context, userID uuid.UUID, functionRef *FunctionRef) (*SupportContext, error) {
	context := &SupportContext{}

	// In a full implementation, this would:
	// 1. Fetch function code from registry
	// 2. Fetch recent deployment logs
	// 3. Fetch execution history
	// 4. Fetch environment variables (masked)
	// 5. Fetch user info

	// For now, return empty context - will be populated by handlers
	return context, nil
}

// handleAIMessage handles getting an AI response for a user message
func (s *Service) handleAIMessage(ctx context.Context, conversation *SupportConversation, userMessage *SupportMessage) {
	if s.aiClient == nil {
		s.logger.Warn("AI client not configured, skipping AI response")
		return
	}

	// Get conversation history
	history, err := s.repo.ListMessages(ctx, conversation.ID, 50, 0)
	if err != nil {
		s.logger.WithError(err).Error("failed to get message history")
		return
	}

	// Get context
	context, err := s.GetSupportContext(ctx, conversation.UserID, conversation.FunctionRef)
	if err != nil {
		s.logger.WithError(err).Error("failed to get support context")
		return
	}

	// Call AI
	req := &AIRequest{
		ConversationID: conversation.ID,
		UserMessage:    userMessage.Content,
		Context:        context,
		History:        history,
	}

	resp, err := s.aiClient.GenerateSupportResponse(ctx, req)
	if err != nil {
		s.logger.WithError(err).Error("AI response generation failed")
		s.sendSystemMessage(ctx, conversation.ID, "I'm having trouble processing your request. A human support agent will be with you shortly.")
		return
	}

	// Send AI response
	_, err = s.SendAIResponse(ctx, conversation.ID, resp.Message, resp.Confidence, resp.Model)
	if err != nil {
		s.logger.WithError(err).Error("failed to send AI response")
	}
}

// sendSystemMessage sends a system message to a conversation
func (s *Service) sendSystemMessage(ctx context.Context, conversationID uuid.UUID, content string) error {
	message := &SupportMessage{
		ConversationID: conversationID,
		AuthorID:       uuid.Nil,
		AuthorType:    AuthorSystem,
		MessageType:   TypeSystem,
		Content:        content,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return err
	}

	s.publishEvent(ctx, "message.created", message)
	return nil
}

// publishEvent publishes an event to Redis for real-time updates
func (s *Service) publishEvent(ctx context.Context, channel string, data interface{}) {
	if s.redis == nil {
		return
	}

	message, err := json.Marshal(data)
	if err != nil {
		s.logger.WithError(err).Error("failed to marshal event")
		return
	}

	if err := s.redis.Publish(ctx, channel, string(message)); err != nil {
		s.logger.WithError(err).Error("failed to publish event")
	}
}

// GetOrCreateConversation gets or creates a support conversation for a user
func (s *Service) GetOrCreateConversation(ctx context.Context, userID uuid.UUID, req *CreateConversationRequest) (*SupportConversation, error) {
	// For MVP, always create a new conversation
	// Future: could look for existing active conversation
	return s.CreateConversation(ctx, userID, req)
}
