package frg_collab

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type CursorPosition struct {
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	ViewportX    float64 `json:"viewport_x"`
	ViewportY    float64 `json:"viewport_y"`
	ViewportZoom float64 `json:"viewport_zoom"`
	SelectedNode string  `json:"selected_node,omitempty"`
}

type Participant struct {
	UserID       uuid.UUID      `json:"user_id"`
	DisplayName  string         `json:"display_name"`
	Color        string         `json:"color"`
	Cursor       *CursorPosition `json:"cursor,omitempty"`
	IsActive     bool            `json:"is_active"`
	LastActivity time.Time       `json:"last_activity_at"`
}

type GraphSession struct {
	ID           uuid.UUID      `json:"id"`
	GraphID      uuid.UUID      `json:"graph_id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	OwnerUserID  uuid.UUID      `json:"owner_user_id"`
	Participants  []Participant  `json:"participants"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CursorUpdate struct {
	Type      string          `json:"type"`
	UserID    uuid.UUID       `json:"user_id"`
	Timestamp int64           `json:"timestamp"`
	Cursor    *CursorPosition `json:"cursor"`
}

type ViewportUpdate struct {
	Type      string    `json:"type"`
	UserID    uuid.UUID `json:"user_id"`
	Timestamp int64     `json:"timestamp"`
	ViewportX float64   `json:"viewport_x"`
	ViewportY float64   `json:"viewport_y"`
	Zoom      float64   `json:"zoom"`
}

type NodeSelection struct {
	Type      string      `json:"type"`
	UserID    uuid.UUID   `json:"user_id"`
	Timestamp int64       `json:"timestamp"`
	NodeID    string      `json:"node_id,omitempty"`
	Selected  bool        `json:"selected"`
}

type FRGCollaborationService struct {
	redis    *redis.Client
	logger   *logrus.Logger
	sessions map[string]*GraphSession
	mu       sync.RWMutex
	hub      *FRGCollaborationHub
}

func NewFRGCollaborationService(redisClient *redis.Client, logger *logrus.Logger) *FRGCollaborationService {
	svc := &FRGCollaborationService{
		redis:    redisClient,
		logger:   logger,
		sessions: make(map[string]*GraphSession),
		hub:      NewFRGCollaborationHub(logger),
	}
	go svc.hub.Run()
	return svc
}

func (s *FRGCollaborationService) WebSocketHub() *FRGCollaborationHub {
	return s.hub
}

func (s *FRGCollaborationService) GetOrCreateSession(ctx context.Context, graphID, tenantID, ownerUserID uuid.UUID) (*GraphSession, error) {
	sessionKey := fmt.Sprintf("frg:%s", graphID.String())

	s.mu.RLock()
	if session, ok := s.sessions[sessionKey]; ok {
		s.mu.RUnlock()
		return session, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[sessionKey]; ok {
		return session, nil
	}

	session := &GraphSession{
		ID:          uuid.New(),
		GraphID:     graphID,
		TenantID:    tenantID,
		OwnerUserID: ownerUserID,
		Participants: []Participant{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.sessions[sessionKey] = session

	if s.redis != nil {
		if err := s.cacheSession(ctx, session); err != nil {
			s.logger.WithError(err).Warn("Failed to cache FRG session in Redis")
		}
	}

	return session, nil
}

func (s *FRGCollaborationService) JoinSession(ctx context.Context, graphID uuid.UUID, userID uuid.UUID, displayName string) (*GraphSession, *Participant, error) {
	sessionKey := fmt.Sprintf("frg:%s", graphID.String())

	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return nil, nil, fmt.Errorf("session not found for graph: %s", graphID.String())
	}

	participant := Participant{
		UserID:       userID,
		DisplayName:  displayName,
		Color:        assignColor(len(session.Participants)),
		Cursor:       nil,
		IsActive:     true,
		LastActivity: time.Now(),
	}

	session.Participants = append(session.Participants, participant)
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.redis != nil {
		_ = s.cacheSession(ctx, session)
	}

	s.hub.BroadcastToGraph(graphID.String(), newJoinMessage(&participant), nil)

	return session, &participant, nil
}

func (s *FRGCollaborationService) LeaveSession(ctx context.Context, graphID uuid.UUID, userID uuid.UUID) error {
	sessionKey := fmt.Sprintf("frg:%s", graphID.String())

	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("session not found")
	}

	for i, p := range session.Participants {
		if p.UserID == userID {
			session.Participants = append(session.Participants[:i], session.Participants[i+1:]...)
			break
		}
	}
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.redis != nil {
		_ = s.cacheSession(ctx, session)
	}

	s.hub.BroadcastToGraph(graphID.String(), newLeaveMessage(userID), nil)

	return nil
}

func (s *FRGCollaborationService) UpdateCursor(ctx context.Context, graphID string, userID uuid.UUID, cursor *CursorPosition) error {
	sessionKey := fmt.Sprintf("frg:%s", graphID)

	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("session not found")
	}

	for i := range session.Participants {
		if session.Participants[i].UserID == userID {
			session.Participants[i].Cursor = cursor
			session.Participants[i].LastActivity = time.Now()
			break
		}
	}
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.hub.BroadcastToGraph(graphID, newCursorMessage(userID, cursor), nil)

	return nil
}

func (s *FRGCollaborationService) UpdateViewport(ctx context.Context, graphID string, userID uuid.UUID, x, y, zoom float64) error {
	sessionKey := fmt.Sprintf("frg:%s", graphID)

	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("session not found")
	}

	for i := range session.Participants {
		if session.Participants[i].UserID == userID {
			session.Participants[i].LastActivity = time.Now()
			break
		}
	}
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.hub.BroadcastToGraph(graphID, newViewportMessage(userID, x, y, zoom), nil)

	return nil
}

func (s *FRGCollaborationService) UpdateNodeSelection(ctx context.Context, graphID string, userID uuid.UUID, nodeID string, selected bool) error {
	sessionKey := fmt.Sprintf("frg:%s", graphID)

	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("session not found")
	}

	for i := range session.Participants {
		if session.Participants[i].UserID == userID {
			session.Participants[i].LastActivity = time.Now()
			break
		}
	}
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.hub.BroadcastToGraph(graphID, newNodeSelectionMessage(userID, nodeID, selected), nil)

	return nil
}

func (s *FRGCollaborationService) cacheSession(ctx context.Context, session *GraphSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("frg_collab:session:%s", session.GraphID.String())
	return s.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

func (s *FRGCollaborationService) GetSessionParticipants(ctx context.Context, graphID uuid.UUID) ([]Participant, error) {
	sessionKey := fmt.Sprintf("frg:%s", graphID.String())

	s.mu.RLock()
	session, ok := s.sessions[sessionKey]
	if ok {
		participants := make([]Participant, len(session.Participants))
		copy(participants, session.Participants)
		s.mu.RUnlock()
		return participants, nil
	}
	s.mu.RUnlock()

	if s.redis != nil {
		key := fmt.Sprintf("frg_collab:session:%s", graphID.String())
		data, err := s.redis.Get(ctx, key).Bytes()
		if err == nil {
			var session GraphSession
			if err := json.Unmarshal(data, &session); err == nil {
				return session.Participants, nil
			}
		}
	}

	return []Participant{}, nil
}

var participantColors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4",
	"#FFEAA7", "#DDA0DD", "#98D8C8", "#F7DC6F",
	"#FF8A5B", "#A8E6CF", "#88D8B0", "#FFE4B5",
}

func assignColor(index int) string {
	return participantColors[index%len(participantColors)]
}
