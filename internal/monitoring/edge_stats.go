package monitoring

import (
	"os"
	"strings"
	"sync"
	"time"
)

// EdgeNodeStats holds per-node (VPS) edge info for display. Region is the GeoDNS region label.
type EdgeNodeStats struct {
	Host   string `json:"host"`   // IP or hostname (e.g. 209.46.125.113)
	Region string `json:"region"` // e.g. "Americas/APAC", "Europe/Africa"
	Status string `json:"status"` // same as aggregate health until per-node probe is implemented
}

// EdgeStats holds current edge monitoring stats for the status API.
type EdgeStats struct {
	ProbeConfigured    bool            `json:"probe_configured"` // true when EDGE_HEALTH_URL is set
	Health             string          `json:"health"`           // "ok", "degraded", "down"
	UptimeRatio        float64         `json:"uptime_ratio"`     // 0.0–1.0 from recent probes
	TotalRequests      uint64          `json:"total_requests"`   // from counter (read via Prometheus or local)
	LastProbeAt        time.Time       `json:"last_probe_at"`    // last probe timestamp
	LastProbeOK        bool            `json:"last_probe_ok"`    // last probe success
	LastProbeLatencyMs int             `json:"last_probe_latency_ms"`
	LastError          string          `json:"last_error,omitempty"` // last probe error message
	ProbeErrorsTotal   uint64          `json:"probe_errors_total"`   // total probe failures
	Nodes              []EdgeNodeStats `json:"nodes,omitempty"`      // VPS nodes and their regions (from EDGE_NODES or default)
}

const (
	edgeProbeWindowSize = 100 // number of recent probes for uptime ratio
)

var (
	edgeStatsMu      sync.RWMutex
	edgeStats        EdgeStats
	edgeProbeHistory []bool // recent probe results for uptime ratio
	edgeRequestCount uint64 // local count for API (in addition to Prometheus counter)
)

// UpdateEdgeProbeAndMetrics updates in-memory edge stats and Prometheus metrics after a probe.
func UpdateEdgeProbeAndMetrics(ok bool, latencyMs int, errorMessage string) {
	edgeStatsMu.Lock()
	defer edgeStatsMu.Unlock()

	now := time.Now()
	edgeStats.LastProbeAt = now
	edgeStats.LastProbeOK = ok
	edgeStats.LastProbeLatencyMs = latencyMs
	edgeStats.LastError = errorMessage
	if !ok {
		edgeStats.ProbeErrorsTotal++
	}

	// Rolling uptime over last N probes
	if len(edgeProbeHistory) >= edgeProbeWindowSize {
		edgeProbeHistory = edgeProbeHistory[1:]
	}
	edgeProbeHistory = append(edgeProbeHistory, ok)
	successes := 0
	for _, v := range edgeProbeHistory {
		if v {
			successes++
		}
	}
	if len(edgeProbeHistory) > 0 {
		edgeStats.UptimeRatio = float64(successes) / float64(len(edgeProbeHistory))
	}
	edgeStats.TotalRequests = edgeRequestCount

	// Health label: ok if last probe OK and uptime > 0.95; degraded if uptime 0.8–0.95; down otherwise
	if edgeStats.LastProbeOK && edgeStats.UptimeRatio >= 0.95 {
		edgeStats.Health = "ok"
	} else if edgeStats.UptimeRatio >= 0.8 {
		edgeStats.Health = "degraded"
	} else {
		edgeStats.Health = "down"
	}

	// Update Prometheus
	UpdateEdgeProbeResult(ok, latencyMs, errorMessage)
	UpdateEdgeUptimeRatio(edgeStats.UptimeRatio)
}

// RecordEdgeRequestAndMetric increments local request count and Prometheus counter.
func RecordEdgeRequestAndMetric() {
	edgeStatsMu.Lock()
	edgeRequestCount++
	edgeStats.TotalRequests = edgeRequestCount
	edgeStatsMu.Unlock()
	RecordEdgeRequest()
}

// defaultEdgeNodes are the two VPS nodes from deploy/edge and cloudflare-geo-dns (Americas/APAC vs Europe/Africa).
var defaultEdgeNodes = []EdgeNodeStats{
	{Host: "209.46.125.113", Region: "Americas / APAC", Status: ""},
	{Host: "217.160.124.206", Region: "Europe / Africa", Status: ""},
}

// parseEdgeNodes returns nodes from EDGE_NODES env (format "host:region,host:region") or default list.
func parseEdgeNodes(aggregateStatus string) []EdgeNodeStats {
	const envKey = "EDGE_NODES"
	s := os.Getenv(envKey)
	if s == "" {
		out := make([]EdgeNodeStats, len(defaultEdgeNodes))
		for i := range defaultEdgeNodes {
			out[i] = EdgeNodeStats{
				Host:   defaultEdgeNodes[i].Host,
				Region: defaultEdgeNodes[i].Region,
				Status: aggregateStatus,
			}
		}
		return out
	}
	var nodes []EdgeNodeStats
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, ":")
		host, region := part, ""
		if idx > 0 {
			host = strings.TrimSpace(part[:idx])
			region = strings.TrimSpace(part[idx+1:])
		}
		if host != "" {
			nodes = append(nodes, EdgeNodeStats{Host: host, Region: region, Status: aggregateStatus})
		}
	}
	if len(nodes) == 0 {
		return defaultEdgeNodes
	}
	return nodes
}

// GetEdgeStats returns a copy of current edge stats for the API.
func GetEdgeStats() EdgeStats {
	edgeStatsMu.RLock()
	defer edgeStatsMu.RUnlock()
	probeConfigured := os.Getenv("EDGE_HEALTH_URL") != ""
	agg := edgeStats.Health
	if agg == "" {
		agg = "unknown"
	}
	return EdgeStats{
		ProbeConfigured:    probeConfigured,
		Health:             edgeStats.Health,
		UptimeRatio:        edgeStats.UptimeRatio,
		TotalRequests:      edgeRequestCount,
		LastProbeAt:        edgeStats.LastProbeAt,
		LastProbeOK:        edgeStats.LastProbeOK,
		LastProbeLatencyMs: edgeStats.LastProbeLatencyMs,
		LastError:          edgeStats.LastError,
		ProbeErrorsTotal:   edgeStats.ProbeErrorsTotal,
		Nodes:              parseEdgeNodes(agg),
	}
}
