package iot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	coapmux "github.com/plgd-dev/go-coap/v3/mux"
	coapnet "github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp/server"
	"github.com/sirupsen/logrus"
)

type COAPServer struct {
	port     int
	bridge   *Bridge
	logger   *logrus.Logger
	server   *server.Server
	listener *coapnet.UDPConn
	mu       sync.Mutex
	started  bool
}

type COAPRequest struct {
	DeviceID  string          `json:"device_id"`
	AuthToken string          `json:"auth_token,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}

type COAPResponse struct {
	Status   string          `json:"status"`
	DeviceID string          `json:"device_id,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	Error    string          `json:"error,omitempty"`
}

func NewCOAPServer(port int, bridge *Bridge, logger *logrus.Logger) *COAPServer {
	if logger == nil {
		logger = logrus.New()
	}
	return &COAPServer{
		port:   port,
		bridge: bridge,
		logger: logger,
	}
}

func (s *COAPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("COAP server already started")
	}

	router := coapmux.NewRouter()
	router.HandleFunc("/telemetry", s.handleTelemetry)
	router.HandleFunc("/state", s.handleState)
	router.HandleFunc("/command", s.handleCommand)
	router.HandleFunc("/observe", s.handleObserve)

	s.server = server.New(
		options.WithMux(router),
		options.WithContext(ctx),
		options.WithErrors(func(err error) {
			s.logger.WithError(err).Debug("COAP server error")
		}),
	)

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := coapnet.NewListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP %s: %w", addr, err)
	}
	s.listener = listener

	go func() {
		if err := s.server.Serve(listener); err != nil {
			s.logger.WithError(err).Error("COAP server serve error")
		}
	}()

	s.started = true
	s.logger.WithField("port", s.port).Info("COAP server started")
	return nil
}

func (s *COAPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	if s.server != nil {
		s.server.Stop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.started = false
	s.logger.Info("COAP server stopped")
	return nil
}

func (s *COAPServer) Port() int { return s.port }

func (s *COAPServer) handleTelemetry(w coapmux.ResponseWriter, r *coapmux.Message) {
	s.dispatch(w, r, "telemetry")
}

func (s *COAPServer) handleState(w coapmux.ResponseWriter, r *coapmux.Message) {
	s.dispatch(w, r, "state")
}

func (s *COAPServer) handleCommand(w coapmux.ResponseWriter, r *coapmux.Message) {
	s.dispatch(w, r, "command")
}

func (s *COAPServer) handleObserve(w coapmux.ResponseWriter, r *coapmux.Message) {
	s.dispatch(w, r, "observe")
}

func (s *COAPServer) dispatch(w coapmux.ResponseWriter, r *coapmux.Message, eventType string) {
	if s.bridge == nil {
		_ = w.SetResponse(codes.InternalServerError, message.AppJSON, bytes.NewReader([]byte(`{"status":"error","error":"bridge not configured"}`)))
		return
	}

	body, err := readCOAPBody(r)
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.AppJSON, bytes.NewReader([]byte(`{"status":"error","error":"invalid body"}`)))
		return
	}

	var req COAPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		_ = w.SetResponse(codes.BadRequest, message.AppJSON, bytes.NewReader([]byte(`{"status":"error","error":"invalid json"}`)))
		return
	}

	if req.DeviceID == "" {
		_ = w.SetResponse(codes.BadRequest, message.AppJSON, bytes.NewReader([]byte(`{"status":"error","error":"missing device_id"}`)))
		return
	}

	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.AppJSON, bytes.NewReader([]byte(`{"status":"error","error":"invalid device_id"}`)))
		return
	}

	if req.Timestamp == 0 {
		req.Timestamp = time.Now().Unix()
	}

	if err := s.bridge.HandleCOAPRequest(context.Background(), deviceID, eventType, req.Payload, req.AuthToken); err != nil {
		errBody, _ := json.Marshal(COAPResponse{Status: "error", Error: err.Error()})
		_ = w.SetResponse(codes.InternalServerError, message.AppJSON, bytes.NewReader(errBody))
		return
	}

	respBody, _ := json.Marshal(COAPResponse{Status: "ok", DeviceID: req.DeviceID})
	_ = w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader(respBody))
}

func readCOAPBody(r *coapmux.Message) ([]byte, error) {
	if r == nil || r.Message == nil {
		return nil, fmt.Errorf("nil message")
	}
	return r.Message.ReadBody()
}
