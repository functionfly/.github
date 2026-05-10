package collaboration

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Participant struct {
	UserID         uuid.UUID `json:"user_id"`
	DisplayName    string    `json:"display_name"`
	Color          string    `json:"color"`
	CursorPosition int       `json:"cursor_position"`
	SelectionStart int       `json:"selection_start"`
	SelectionEnd   int       `json:"selection_end"`
	IsActive       bool      `json:"is_active"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

type Session struct {
	ID          uuid.UUID      `json:"id"`
	SessionKey  string         `json:"session_key"`
	FunctionID  uuid.UUID      `json:"function_id"`
	OwnerUserID uuid.UUID      `json:"owner_user_id"`
	InputState  json.RawMessage `json:"input_state"`
	Participants []Participant `json:"participants"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

type Operation struct {
	Type      string          `json:"type"`
	UserID    uuid.UUID       `json:"user_id"`
	Timestamp int64           `json:"timestamp"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value,omitempty"`
	OldValue  json.RawMessage `json:"old_value,omitempty"`
}

type CollaborationService struct {
	redis   *redis.Client
	logger  *logrus.Logger
	sessions map[string]*Session
	mu       sync.RWMutex
	hub      *CollaborationHub
}

func NewCollaborationService(redisClient *redis.Client, logger *logrus.Logger) *CollaborationService {
	svc := &CollaborationService{
		redis:   redisClient,
		logger:  logger,
		sessions: make(map[string]*Session),
		hub:     NewCollaborationHub(logger),
	}
	go svc.hub.Run()
	return svc
}

func (s *CollaborationService) WebSocketHub() *CollaborationHub {
	return s.hub
}

func (s *CollaborationService) CreateSession(ctx context.Context, functionID, ownerUserID uuid.UUID, initialInput json.RawMessage) (*Session, error) {
	sessionKey, err := generateSessionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	session := &Session{
		ID:          uuid.New(),
		SessionKey:  sessionKey,
		FunctionID:  functionID,
		OwnerUserID: ownerUserID,
		InputState:  initialInput,
		Participants: []Participant{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.sessions[sessionKey] = session
	s.mu.Unlock()

	if s.redis != nil {
		if err := s.cacheSession(ctx, session); err != nil {
			s.logger.WithError(err).Warn("Failed to cache session in Redis")
		}
	}

	return session, nil
}

func (s *CollaborationService) GetSession(ctx context.Context, sessionKey string) (*Session, error) {
	s.mu.RLock()
	if session, ok := s.sessions[sessionKey]; ok {
		s.mu.RUnlock()
		return session, nil
	}
	s.mu.RUnlock()

	if s.redis != nil {
		cached, err := s.getCachedSession(ctx, sessionKey)
		if err == nil && cached != nil {
			s.mu.Lock()
			s.sessions[sessionKey] = cached
			s.mu.Unlock()
			return cached, nil
		}
	}

	return nil, fmt.Errorf("session not found: %s", sessionKey)
}

func (s *CollaborationService) JoinSession(ctx context.Context, sessionKey string, userID uuid.UUID, displayName string) (*Session, *Participant, error) {
	session, err := s.GetSession(ctx, sessionKey)
	if err != nil {
		return nil, nil, err
	}

	participant := &Participant{
		UserID:         userID,
		DisplayName:    displayName,
		Color:          assignColor(len(session.Participants)),
		CursorPosition: 0,
		SelectionStart: 0,
		SelectionEnd:   0,
		IsActive:       true,
		LastActivityAt: time.Now(),
	}

	s.mu.Lock()
	session.Participants = append(session.Participants, *participant)
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.redis != nil {
		s.cacheSession(ctx, session)
	}

	s.hub.BroadcastToSession(sessionKey, newJoinMessage(participant), nil)

	return session, participant, nil
}

func (s *CollaborationService) LeaveSession(ctx context.Context, sessionKey string, userID uuid.UUID) error {
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
		s.cacheSession(ctx, session)
	}

	s.hub.BroadcastToSession(sessionKey, newLeaveMessage(userID), nil)

	return nil
}

func (s *CollaborationService) ApplyOperation(ctx context.Context, sessionKey string, op *Operation) (*Session, error) {
	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("session not found")
	}

	op.Timestamp = time.Now().UnixMilli()

	merged, err := applyOperationToState(session.InputState, op)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	session.InputState = merged
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.redis != nil {
		s.cacheSession(ctx, session)
	}

	s.hub.BroadcastToSession(sessionKey, newOperationMessage(op), nil)

	return session, nil
}

func (s *CollaborationService) UpdateCursor(ctx context.Context, sessionKey string, userID uuid.UUID, position, start, end int) error {
	s.mu.Lock()
	session, ok := s.sessions[sessionKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("session not found")
	}

	for i := range session.Participants {
		if session.Participants[i].UserID == userID {
			session.Participants[i].CursorPosition = position
			session.Participants[i].SelectionStart = start
			session.Participants[i].SelectionEnd = end
			session.Participants[i].LastActivityAt = time.Now()
			break
		}
	}
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.hub.BroadcastToSession(sessionKey, newCursorMessage(userID, position, start, end), nil)

	return nil
}

func (s *CollaborationService) cacheSession(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, "collab:session:"+session.SessionKey, data, 24*time.Hour).Err()
}

func (s *CollaborationService) getCachedSession(ctx context.Context, sessionKey string) (*Session, error) {
	data, err := s.redis.Get(ctx, "collab:session:"+sessionKey).Bytes()
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func applyOperationToState(state json.RawMessage, op *Operation) (json.RawMessage, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(state, &doc); err != nil {
		doc = make(map[string]interface{})
	}

	switch op.Type {
	case "set":
		setValueAtPath(doc, op.Path, op.Value)
	case "delete":
		deleteValueAtPath(doc, op.Path)
	case "insert_array":
		insertAtArray(doc, op.Path, op.Value)
	}

	return json.Marshal(doc)
}

func setValueAtPath(doc map[string]interface{}, path string, value json.RawMessage) {
	keys := parsePath(path)
	if len(keys) == 0 {
		return
	}

	current := doc
	for i := 0; i < len(keys)-1; i++ {
		if next, ok := current[keys[i]].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[keys[i]] = newMap
			current = newMap
		}
	}

	var val interface{}
	json.Unmarshal(value, &val)
	current[keys[len(keys)-1]] = val
}

func deleteValueAtPath(doc map[string]interface{}, path string) {
	keys := parsePath(path)
	if len(keys) == 0 {
		return
	}

	current := doc
	for i := 0; i < len(keys)-1; i++ {
		if next, ok := current[keys[i]].(map[string]interface{}); ok {
			current = next
		} else {
			return
		}
	}
	delete(current, keys[len(keys)-1])
}

func insertAtArray(doc map[string]interface{}, path string, value json.RawMessage) {
	keys := parsePath(path)
	if len(keys) == 0 {
		return
	}

	current := doc
	for i := 0; i < len(keys)-1; i++ {
		if next, ok := current[keys[i]].(map[string]interface{}); ok {
			current = next
		} else {
			return
		}
	}

	if arr, ok := current[keys[len(keys)-1]].([]interface{}); ok {
		var val interface{}
		json.Unmarshal(value, &val)
		current[keys[len(keys)-1]] = append(arr, val)
	}
}

func parsePath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}
	path = trimLeadingSlash(path)
	return splitPath(path)
}

func trimLeadingSlash(path string) string {
	if len(path) > 0 && path[0] == '/' {
		return path[1:]
	}
	return path
}

func splitPath(path string) []string {
	var result []string
	var current []byte
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if len(current) > 0 {
				result = append(result, string(current))
				current = nil
			}
		} else if path[i] == '.' {
			if len(current) > 0 {
				result = append(result, string(current))
				current = nil
			}
		} else {
			current = append(current, path[i])
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

func generateSessionKey() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cs_" + base64.URLEncoding.EncodeToString(b), nil
}

var participantColors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4",
	"#FFEAA7", "#DDA0DD", "#98D8C8", "#F7DC6F",
}

func assignColor(index int) string {
	return participantColors[index%len(participantColors)]
}

type WSClient struct {
	SessionKey string
	UserID     uuid.UUID
	Conn       *websocket.Conn
	Send       chan []byte
	Hub        *CollaborationHub
}

type CollaborationHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan *BroadcastMessage
	logger     *logrus.Logger
	stop       chan struct{}
}

type BroadcastMessage struct {
	SessionKey string
	Data       []byte
	Exclude    *WSClient
}

func NewCollaborationHub(logger *logrus.Logger) *CollaborationHub {
	return &CollaborationHub{
		clients:    make(map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

func (h *CollaborationHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.WithFields(logrus.Fields{
				"session_key": client.SessionKey,
				"user_id":     client.UserID,
			}).Debug("Collaboration client registered")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				if client.SessionKey == msg.SessionKey && client != msg.Exclude {
					select {
					case client.Send <- msg.Data:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}

		case <-h.stop:
			return
		}
	}
}

func (h *CollaborationHub) Register(client *WSClient) {
	h.register <- client
}

func (h *CollaborationHub) Unregister(client *WSClient) {
	h.unregister <- client
}

func (h *CollaborationHub) BroadcastToSession(sessionKey string, msg *WSMessage, exclude *WSClient) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal broadcast message")
		return
	}
	h.broadcast <- &BroadcastMessage{
		SessionKey: sessionKey,
		Data:       data,
		Exclude:    exclude,
	}
}

type WSMessage struct {
	Type       string      `json:"type"`
	Payload    interface{} `json:"payload,omitempty"`
	SessionKey string      `json:"session_key,omitempty"`
	UserID     string      `json:"user_id,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

func newJoinMessage(p *Participant) *WSMessage {
	return &WSMessage{
		Type:      "participant_joined",
		Payload:   p,
		Timestamp: time.Now().UnixMilli(),
	}
}

func newLeaveMessage(userID uuid.UUID) *WSMessage {
	return &WSMessage{
		Type:      "participant_left",
		Payload:   map[string]string{"user_id": userID.String()},
		Timestamp: time.Now().UnixMilli(),
	}
}

func newOperationMessage(op *Operation) *WSMessage {
	return &WSMessage{
		Type:      "operation",
		Payload:   op,
		UserID:    op.UserID.String(),
		Timestamp: op.Timestamp,
	}
}

func newCursorMessage(userID uuid.UUID, pos, start, end int) *WSMessage {
	return &WSMessage{
		Type: "cursor_update",
		Payload: map[string]interface{}{
			"user_id":         userID.String(),
			"cursor_position": pos,
			"selection_start": start,
			"selection_end":   end,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}