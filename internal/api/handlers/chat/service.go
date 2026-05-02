package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo      *Repository
	aiClient  *AIServiceClient
	connectorRegistry *ConnectorRegistry
	logger    *logrus.Logger
}

func NewService(repo *Repository, aiClient *AIServiceClient, connectorRegistry *ConnectorRegistry, logger *logrus.Logger) *Service {
	if logger == nil {
		logger = logrus.New()
	}
	return &Service{
		repo:      repo,
		aiClient:  aiClient,
		connectorRegistry: connectorRegistry,
		logger:    logger,
	}
}

type ChatRequest struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	TenantID   uuid.UUID
	Message    string
	History    []ChatMessage
	Model      string
	Connectors []ChatConnector
}

type AIResponse struct {
	Message string
	Tokens  int
	Model   string
}

func (s *Service) GenerateResponse(ctx context.Context, req *ChatRequest) (*AIResponse, error) {
	aiReq := &AIServiceRequest{
		SessionID:  req.SessionID.String(),
		Message:    req.Message,
		History:    req.History,
		Model:      req.Model,
		TenantID:   req.TenantID.String(),
		UserID:     req.UserID.String(),
	}

	connectorContext := s.gatherConnectorContext(ctx, req.Connectors)
	if connectorContext != "" {
		aiReq.Context = map[string]string{
			"connector_data": connectorContext,
		}
	}

	start := time.Now()
	resp, err := s.aiClient.ChatMessage(ctx, aiReq)
	latency := time.Since(start)

	if err != nil {
		s.logger.WithError(err).Warn("AI service call failed, using fallback")
		return &AIResponse{
			Message: "I'm currently unavailable. Please try again in a moment.",
			Tokens:  0,
			Model:   req.Model,
		}, nil
	}

	tokens := 0
	if resp.Usage.TotalTokens > 0 {
		tokens = resp.Usage.TotalTokens
	} else if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 {
		tokens = resp.Usage.PromptTokens + resp.Usage.CompletionTokens
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": req.SessionID,
		"model":      resp.Model,
		"tokens":     tokens,
		"latency":    latency,
	}).Debug("AI response generated")

	return &AIResponse{
		Message: resp.Message,
		Tokens:  tokens,
		Model:   resp.Model,
	}, nil
}

func (s *Service) gatherConnectorContext(ctx context.Context, connectors []ChatConnector) string {
	if len(connectors) == 0 {
		return ""
	}

	var context string
	for _, conn := range connectors {
		connector := s.connectorRegistry.Get(conn.Type)
		if connector == nil {
			continue
		}

		configMap := map[string]interface{}(conn.Config)
		data, err := connector.FetchData(ctx, configMap)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"connector": conn.Type,
				"error":     err,
			}).Warn("Failed to fetch connector data")
			continue
		}

		if jsonData, ok := data.(string); ok {
			context += conn.Name + ": " + jsonData + "\n"
		}
	}

	return context
}

func (s *Service) CreateFunction(ctx context.Context, tenantID, userID uuid.UUID, sessionID *uuid.UUID, prompt, connectorType string) (*ChatFunction, error) {
	req := &CreateFunctionRequest{
		Prompt:        prompt,
		ConnectorType: connectorType,
	}

	code, err := s.aiClient.GenerateFunctionCode(ctx, req)
	if err != nil {
		return nil, err
	}

	fn := &ChatFunction{
		TenantID:  tenantID,
		UserID:   userID,
		SessionID: sessionID,
		Name:     extractFunctionName(code),
		Code:     code,
	}

	if err := s.repo.CreateFunction(ctx, fn); err != nil {
		return nil, err
	}

	return fn, nil
}

func extractFunctionName(code string) string {
	return "generated_function"
}

type CreateFunctionRequest struct {
	Prompt        string
	ConnectorType string
}