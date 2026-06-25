package frg_collab

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type FRGWSClient struct {
	GraphID   string
	UserID    uuid.UUID
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *FRGCollaborationHub
	ExpiresAt time.Time
}

type FRGCollaborationHub struct {
	clients    map[*FRGWSClient]bool
	register   chan *FRGWSClient
	unregister chan *FRGWSClient
	broadcast  chan *BroadcastMessage
	logger     *logrus.Logger
	stop       chan struct{}
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	GraphID string
	Data    []byte
	Exclude *FRGWSClient
}

func NewFRGCollaborationHub(logger *logrus.Logger) *FRGCollaborationHub {
	return &FRGCollaborationHub{
		clients:    make(map[*FRGWSClient]bool),
		register:   make(chan *FRGWSClient),
		unregister: make(chan *FRGWSClient),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

func (h *FRGCollaborationHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.WithFields(logrus.Fields{
				"graph_id": client.GraphID,
				"user_id":  client.UserID,
			}).Debug("FRG collab client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.GraphID == msg.GraphID && client != msg.Exclude {
					select {
					case client.Send <- msg.Data:
					default:
						close(client.Send)
						h.mu.Unlock()
						h.mu.Lock()
						delete(h.clients, client)
						h.mu.Unlock()
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()

		case <-h.stop:
			return
		}
	}
}

func (h *FRGCollaborationHub) Register(client *FRGWSClient) {
	h.register <- client
}

func (h *FRGCollaborationHub) Unregister(client *FRGWSClient) {
	h.unregister <- client
}

func (h *FRGCollaborationHub) BroadcastToGraph(graphID string, msg *FRGWSMessage, exclude *FRGWSClient) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal FRG broadcast message")
		return
	}
	h.broadcast <- &BroadcastMessage{
		GraphID: graphID,
		Data:    data,
		Exclude: exclude,
	}
}

func (h *FRGCollaborationHub) GetClientsForGraph(graphID string) []*FRGWSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var clients []*FRGWSClient
	for client := range h.clients {
		if client.GraphID == graphID {
			clients = append(clients, client)
		}
	}
	return clients
}

type FRGWSMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	GraphID   string     `json:"graph_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

func newJoinMessage(p *Participant) *FRGWSMessage {
	return &FRGWSMessage{
		Type:      "participant_joined",
		Payload:   p,
		Timestamp: time.Now().UnixMilli(),
	}
}

func newLeaveMessage(userID uuid.UUID) *FRGWSMessage {
	return &FRGWSMessage{
		Type:      "participant_left",
		Payload:   map[string]string{"user_id": userID.String()},
		Timestamp: time.Now().UnixMilli(),
	}
}

func newCursorMessage(userID uuid.UUID, cursor *CursorPosition) *FRGWSMessage {
	return &FRGWSMessage{
		Type: "cursor_update",
		Payload: map[string]interface{}{
			"user_id": userID.String(),
			"cursor":  cursor,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

func newViewportMessage(userID uuid.UUID, x, y, zoom float64) *FRGWSMessage {
	return &FRGWSMessage{
		Type: "viewport_update",
		Payload: map[string]interface{}{
			"user_id":    userID.String(),
			"viewport_x": x,
			"viewport_y": y,
			"zoom":       zoom,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

func newNodeSelectionMessage(userID uuid.UUID, nodeID string, selected bool) *FRGWSMessage {
	return &FRGWSMessage{
		Type: "node_selection",
		Payload: map[string]interface{}{
			"user_id":  userID.String(),
			"node_id":  nodeID,
			"selected": selected,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}
