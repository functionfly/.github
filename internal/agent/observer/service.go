package observer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service hooks into graph execution and emits conversion-tagged traces.
type Service struct {
	db           *gorm.DB
	traceRepo    *graph.TraceRepository
	eventTaggers map[string]EventTagger
}

// EventTagger determines the EventType and RevenueCents for a node execution.
type EventTagger func(nodeName string, output map[string]any) (eventType string, revenueCents int64)

// NewService creates a new observer service.
func NewService(db *gorm.DB, traceRepo *graph.TraceRepository) *Service {
	s := &Service{
		db:           db,
		traceRepo:    traceRepo,
		eventTaggers: make(map[string]EventTagger),
	}
	s.registerDefaultTaggers()
	return s
}

// RegisterTagger registers a custom event tagger for a node name.
func (s *Service) RegisterTagger(nodePattern string, tagger EventTagger) {
	s.eventTaggers[nodePattern] = tagger
}

func (s *Service) registerDefaultTaggers() {
	s.eventTaggers["checkout_start"] = func(nodeName string, output map[string]any) (string, int64) {
		return "checkout_started", 0
	}
	s.eventTaggers["payment"] = func(nodeName string, output map[string]any) (string, int64) {
		if status, ok := output["status"].(string); ok && status == "success" {
			if amount, ok := output["amount_cents"].(float64); ok {
				return "payment_success", int64(amount)
			}
			return "payment_success", 0
		}
		if status, ok := output["status"].(string); ok && status == "failed" {
			return "payment_failed", 0
		}
		return "payment_processed", 0
	}
	s.eventTaggers["fraud_detection"] = func(nodeName string, output map[string]any) (string, int64) {
		if decision, ok := output["decision"].(string); ok && decision == "declined" {
			return "fraud_declined", 0
		}
		return "fraud_cleared", 0
	}
	s.eventTaggers["cart"] = func(nodeName string, output map[string]any) (string, int64) {
		if total, ok := output["total_cents"].(float64); ok {
			return "checkout_started", int64(total)
		}
		return "checkout_started", 0
	}
	s.eventTaggers["default"] = func(nodeName string, output map[string]any) (string, int64) {
		return "node_executed", 0
	}
}

// TraceNode emits a trace for a single node execution.
func (s *Service) TraceNode(ctx context.Context, params TraceParams) error {
	eventType, revenueCents := s.determineEvent(params.NodeName, params.Output)

	inputJSON, _ := json.Marshal(params.Input)
	outputJSON, _ := json.Marshal(params.Output)

	trace := &graph.ExecutionTrace{
		ID:           uuid.New(),
		ExecutionID:  params.ExecutionID,
		GraphID:      params.GraphID,
		TenantID:     params.TenantID,
		NodeID:       params.NodeID,
		NodeName:     params.NodeName,
		VerticalTag:  params.VerticalTag,
		Input:        string(inputJSON),
		Output:       string(outputJSON),
		LatencyMs:    params.LatencyMs,
		Status:       string(params.Status),
		EventType:    eventType,
		RevenueCents: revenueCents,
		CreatedAt:    time.Now().UTC(),
	}

	return s.traceRepo.Create(ctx, trace)
}

// BatchTrace emits multiple traces in a single transaction.
func (s *Service) BatchTrace(ctx context.Context, traces []TraceParams) error {
	if len(traces) == 0 {
		return nil
	}
	execTraces := make([]graph.ExecutionTrace, 0, len(traces))
	for _, p := range traces {
		eventType, revenueCents := s.determineEvent(p.NodeName, p.Output)
		inputJSON, _ := json.Marshal(p.Input)
		outputJSON, _ := json.Marshal(p.Output)
		execTraces = append(execTraces, graph.ExecutionTrace{
			ID:           uuid.New(),
			ExecutionID:  p.ExecutionID,
			GraphID:      p.GraphID,
			TenantID:     p.TenantID,
			NodeID:       p.NodeID,
			NodeName:     p.NodeName,
			VerticalTag:  p.VerticalTag,
			Input:        string(inputJSON),
			Output:       string(outputJSON),
			LatencyMs:    p.LatencyMs,
			Status:       string(p.Status),
			EventType:    eventType,
			RevenueCents: revenueCents,
			CreatedAt:    time.Now().UTC(),
		})
	}
	return s.traceRepo.BatchCreate(ctx, execTraces)
}

func (s *Service) determineEvent(nodeName string, output map[string]any) (eventType string, revenueCents int64) {
	if tagger, ok := s.eventTaggers[nodeName]; ok {
		return tagger(nodeName, output)
	}
	for pattern, tagger := range s.eventTaggers {
		if pattern != "default" && len(pattern) < len(nodeName) && nodeName[:len(pattern)] == pattern {
			return tagger(nodeName, output)
		}
	}
	if tagger, ok := s.eventTaggers["default"]; ok {
		return tagger(nodeName, output)
	}
	return "node_executed", 0
}

// TraceParams holds all parameters needed to emit a node trace.
type TraceParams struct {
	ExecutionID  uuid.UUID
	GraphID      uuid.UUID
	TenantID     uuid.UUID
	NodeID       uuid.UUID
	NodeName     string
	Input        map[string]any
	Output       map[string]any
	LatencyMs    int64
	Status       NodeStatus
	VerticalTag  string
}

// NodeStatus represents the outcome of a node execution.
type NodeStatus string

const (
	NodeStatusSuccess NodeStatus = "success"
	NodeStatusFailure NodeStatus = "failure"
	NodeStatusDropOff NodeStatus = "drop_off"
)
