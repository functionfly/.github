package analyzer

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/google/uuid"
)

// Service analyzes execution traces to detect optimization opportunities.
type Service struct {
	traceRepo *graph.TraceRepository
}

// NewService creates a new analyzer service.
func NewService(traceRepo *graph.TraceRepository) *Service {
	return &Service{traceRepo: traceRepo}
}

// AnalyzeByTenant runs conversion analysis across all graphs for a tenant.
func (s *Service) AnalyzeByTenant(ctx context.Context, params AnalyzeByTenantParams) ([]ConversionAnalysis, error) {
	// List traces for all graphs owned by the tenant in the time window
	since := time.Now().Add(-params.TimeWindow)
	traces, err := s.traceRepo.ListByTenant(ctx, params.TenantID, since, 50000)
	if err != nil {
		return nil, err
	}

	if len(traces) == 0 {
		return nil, nil
	}

	// Group traces by graph
	tracesByGraph := make(map[uuid.UUID][]graph.ExecutionTrace)
	for _, t := range traces {
		tracesByGraph[t.GraphID] = append(tracesByGraph[t.GraphID], t)
	}

	var analyses []ConversionAnalysis
	for graphID, graphTraces := range tracesByGraph {
		analysis := s.analyzeTraces(params.TenantID, graphID, graphTraces)
		analyses = append(analyses, *analysis)
	}
	return analyses, nil
}

// AnalyzeByTenantParams holds parameters for tenant-wide analysis.
type AnalyzeByTenantParams struct {
	TenantID    uuid.UUID
	TimeWindow  time.Duration
}

// analyzeTraces performs the actual analysis on a slice of traces.
func (s *Service) analyzeTraces(tenantID, graphID uuid.UUID, traces []graph.ExecutionTrace) *ConversionAnalysis {
	funnelMetrics := s.computeFunnelMetrics(traces)
	dropOffPatterns := s.detectDropOffs(traces)
	failurePatterns := s.detectFailures(traces)
	improvementTargets := s.rankImprovements(funnelMetrics, dropOffPatterns, failurePatterns)

	return &ConversionAnalysis{
		TenantID:           tenantID,
		GraphID:            graphID,
		FunnelMetrics:     funnelMetrics,
		DropOffPatterns:   dropOffPatterns,
		FailurePatterns:   failurePatterns,
		ImprovementTargets: improvementTargets,
	}
}
func (s *Service) Analyze(ctx context.Context, params AnalyzeParams) (*ConversionAnalysis, error) {
	since := time.Now().Add(-params.TimeWindow)

	traces, err := s.traceRepo.ListByGraph(ctx, params.GraphID, since, 10000)
	if err != nil {
		return nil, err
	}

	if len(traces) == 0 {
		return &ConversionAnalysis{
			TenantID:           params.TenantID,
			GraphID:            params.GraphID,
			FunnelMetrics:     map[string]*FunnelMetric{},
			DropOffPatterns:   []DropOffPattern{},
			FailurePatterns:   []FailurePattern{},
			ImprovementTargets: []ImprovementTarget{},
		}, nil
	}

	funnelMetrics := s.computeFunnelMetrics(traces)
	dropOffPatterns := s.detectDropOffs(traces)
	failurePatterns := s.detectFailures(traces)
	improvementTargets := s.rankImprovements(funnelMetrics, dropOffPatterns, failurePatterns)

	return &ConversionAnalysis{
		TenantID:           params.TenantID,
		GraphID:            params.GraphID,
		FunnelMetrics:     funnelMetrics,
		DropOffPatterns:   dropOffPatterns,
		FailurePatterns:   failurePatterns,
		ImprovementTargets: improvementTargets,
	}, nil
}

// AnalyzeParams holds parameters for running an analysis.
type AnalyzeParams struct {
	TenantID    uuid.UUID
	GraphID     uuid.UUID
	TimeWindow  time.Duration
}

// ConversionAnalysis is the output of the analyzer.
type ConversionAnalysis struct {
	TenantID           uuid.UUID                `json:"tenant_id"`
	GraphID            uuid.UUID                `json:"graph_id"`
	FunnelMetrics     map[string]*FunnelMetric  `json:"funnel_metrics"`
	DropOffPatterns   []DropOffPattern          `json:"drop_off_patterns"`
	FailurePatterns   []FailurePattern          `json:"failure_patterns"`
	ImprovementTargets []ImprovementTarget      `json:"improvement_targets"`
}

// FunnelMetric holds per-node funnel statistics.
type FunnelMetric struct {
	NodeID             uuid.UUID `json:"node_id"`
	NodeName           string    `json:"node_name"`
	TotalHits          int       `json:"total_hits"`
	Successes          int       `json:"successes"`
	Failures           int       `json:"failures"`
	DropOffs           int       `json:"drop_offs"`
	SuccessRatePct    float64   `json:"success_rate_pct"`
	AvgLatencyMs       int64     `json:"avg_latency_ms"`
	TotalRevenueCents int64     `json:"total_revenue_cents"`
}

// DropOffPattern identifies a node where users disproportionately abandon the flow.
type DropOffPattern struct {
	NodeID               uuid.UUID `json:"node_id"`
	NodeName             string    `json:"node_name"`
	DropOffCount         int       `json:"drop_off_count"`
	TotalHits            int       `json:"total_hits"`
	DropOffRatePct       float64   `json:"drop_off_rate_pct"`
	EstimatedRevenueLoss int64     `json:"estimated_revenue_loss_cents"`
	Severity             string    `json:"severity"`
}

// FailurePattern identifies a node or sequence that frequently leads to failure.
type FailurePattern struct {
	NodeID          uuid.UUID `json:"node_id"`
	NodeName        string    `json:"node_name"`
	FailureCount    int       `json:"failure_count"`
	TotalHits       int       `json:"total_hits"`
	FailureRatePct  float64   `json:"failure_rate_pct"`
	CommonErrorCode string    `json:"common_error_code,omitempty"`
	Severity        string    `json:"severity"`
}

// ImprovementTarget is a ranked optimization opportunity.
type ImprovementTarget struct {
	NodeID               uuid.UUID `json:"node_id"`
	NodeName             string    `json:"node_name"`
	ChangeType           string    `json:"change_type"`
	Description          string    `json:"description"`
	ExpectedLiftPct       float64   `json:"expected_lift_pct"`
	ExpectedRevenueCents int64     `json:"expected_revenue_cents"`
	RiskScore            float64   `json:"risk_score"`
	Priority             int       `json:"priority"`
}

type nodeStatsAccum struct {
	NodeID       uuid.UUID
	NodeName     string
	TotalHits    int
	Successes    int
	Failures     int
	DropOffs     int
	LatencySum   int64
	RevenueSum   int64
}

func (s *Service) computeFunnelMetrics(traces []graph.ExecutionTrace) map[string]*FunnelMetric {
	metrics := make(map[string]*FunnelMetric)
	nodeStats := make(map[string]*nodeStatsAccum)

	for _, t := range traces {
		key := t.NodeID.String()
		if _, ok := nodeStats[key]; !ok {
			nodeStats[key] = &nodeStatsAccum{NodeID: t.NodeID, NodeName: t.NodeName}
		}
		acc := nodeStats[key]
		acc.TotalHits++
		acc.LatencySum += t.LatencyMs
		acc.RevenueSum += t.RevenueCents

		switch t.Status {
		case "success":
			acc.Successes++
		case "failure":
			acc.Failures++
		case "drop_off":
			acc.DropOffs++
		}
	}

	for key, acc := range nodeStats {
		rate := 0.0
		if acc.TotalHits > 0 {
			rate = float64(acc.Successes) / float64(acc.TotalHits) * 100
		}
		avgLat := int64(0)
		if acc.TotalHits > 0 {
			avgLat = acc.LatencySum / int64(acc.TotalHits)
		}
		metrics[key] = &FunnelMetric{
			NodeID:             acc.NodeID,
			NodeName:           acc.NodeName,
			TotalHits:          acc.TotalHits,
			Successes:          acc.Successes,
			Failures:           acc.Failures,
			DropOffs:           acc.DropOffs,
			SuccessRatePct:     rate,
			AvgLatencyMs:       avgLat,
			TotalRevenueCents: acc.RevenueSum,
		}
	}
	return metrics
}

func (s *Service) detectDropOffs(traces []graph.ExecutionTrace) []DropOffPattern {
	execTraces := make(map[uuid.UUID][]graph.ExecutionTrace)
	for _, t := range traces {
		execTraces[t.ExecutionID] = append(execTraces[t.ExecutionID], t)
	}

	nodeDropoffs := make(map[string]int)
	nodeTotals := make(map[string]int)
	for _, tList := range execTraces {
		seen := make(map[string]bool)
		for _, t := range tList {
			nodeTotals[t.NodeID.String()]++
			if t.Status == "drop_off" && !seen[t.NodeID.String()] {
				nodeDropoffs[t.NodeID.String()]++
				seen[t.NodeID.String()] = true
			}
		}
	}

	var patterns []DropOffPattern
	for nodeIDStr, dropOffs := range nodeDropoffs {
		nodeID, _ := uuid.Parse(nodeIDStr)
		total := nodeTotals[nodeIDStr]
		if total == 0 {
			continue
		}
		rate := float64(dropOffs) / float64(total) * 100
		severity := "low"
		if rate > 20 {
			severity = "critical"
		} else if rate > 10 {
			severity = "high"
		} else if rate > 5 {
			severity = "medium"
		}
		patterns = append(patterns, DropOffPattern{
			NodeID:               nodeID,
			NodeName:             findNodeName(traces, nodeID),
			DropOffCount:         dropOffs,
			TotalHits:            total,
			DropOffRatePct:       rate,
			EstimatedRevenueLoss: int64(float64(rate)/100.0*10000),
			Severity:             severity,
		})
	}
	return patterns
}

func (s *Service) detectFailures(traces []graph.ExecutionTrace) []FailurePattern {
	nodeFailures := make(map[string]int)
	nodeTotals := make(map[string]int)
	for _, t := range traces {
		key := t.NodeID.String()
		nodeTotals[key]++
		if t.Status == "failure" {
			nodeFailures[key]++
		}
	}
	var patterns []FailurePattern
	for nodeIDStr, failures := range nodeFailures {
		nodeID, _ := uuid.Parse(nodeIDStr)
		total := nodeTotals[nodeIDStr]
		if total == 0 {
			continue
		}
		rate := float64(failures) / float64(total) * 100
		severity := "low"
		if rate > 20 {
			severity = "critical"
		} else if rate > 10 {
			severity = "high"
		} else if rate > 5 {
			severity = "medium"
		}
		patterns = append(patterns, FailurePattern{
			NodeID:         nodeID,
			NodeName:       findNodeName(traces, nodeID),
			FailureCount:   failures,
			TotalHits:      total,
			FailureRatePct: rate,
			Severity:       severity,
		})
	}
	return patterns
}

func (s *Service) rankImprovements(funnelMetrics map[string]*FunnelMetric, dropOffs []DropOffPattern, failures []FailurePattern) []ImprovementTarget {
	var targets []ImprovementTarget

	for _, d := range dropOffs {
		if d.Severity == "critical" || d.Severity == "high" {
			targets = append(targets, ImprovementTarget{
				NodeID:               d.NodeID,
				NodeName:             d.NodeName,
				ChangeType:           "add_retry",
				Description:          "Add a retry specialist node after " + d.NodeName + " to recover drop-offs",
				ExpectedLiftPct:       d.DropOffRatePct / 2,
				ExpectedRevenueCents: d.EstimatedRevenueLoss / 2,
				RiskScore:            0.15,
				Priority:             1,
			})
		}
	}

	for _, f := range failures {
		if f.Severity == "critical" || f.Severity == "high" {
			targets = append(targets, ImprovementTarget{
				NodeID:               f.NodeID,
				NodeName:             f.NodeName,
				ChangeType:           "add_specialist",
				Description:          "Add a specialist agent to handle failures at " + f.NodeName,
				ExpectedLiftPct:       f.FailureRatePct * 0.8,
				ExpectedRevenueCents: int64(f.FailureRatePct * 100),
				RiskScore:            0.3,
				Priority:             2,
			})
		}
	}

	for _, m := range funnelMetrics {
		if m.AvgLatencyMs > 5000 && m.SuccessRatePct > 70 {
			targets = append(targets, ImprovementTarget{
				NodeID:               m.NodeID,
				NodeName:             m.NodeName,
				ChangeType:           "optimize",
				Description:          "Optimize " + m.NodeName + " — high latency is hurting conversion",
				ExpectedLiftPct:       3.0,
				ExpectedRevenueCents: 5000,
				RiskScore:            0.1,
				Priority:             3,
			})
		}
	}

	for i := range targets {
		for j := i + 1; j < len(targets); j++ {
			if targets[j].ExpectedRevenueCents > targets[i].ExpectedRevenueCents {
				targets[i], targets[j] = targets[j], targets[i]
			}
		}
	}
	for i := range targets {
		targets[i].Priority = i + 1
	}

	return targets
}

func findNodeName(traces []graph.ExecutionTrace, nodeID uuid.UUID) string {
	for _, t := range traces {
		if t.NodeID == nodeID {
			return t.NodeName
		}
	}
	return ""
}
