package routing

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GlobalLoadBalancer provides global load balancing across regions
type GlobalLoadBalancer struct {
	geoRouter     *GeoRouter
	healthChecker *HealthChecker
	regionalStats map[string]*RegionalStats
	mu            sync.RWMutex
	config        *GLBConfig
}

// GLBConfig holds Global Load Balancer configuration
type GLBConfig struct {
	// HealthCheckInterval is how often to check backend health
	HealthCheckInterval time.Duration
	// FailoverThreshold is consecutive failures before failover
	FailoverThreshold int
	// RecoveryThreshold is successes needed to restore a backend
	RecoveryThreshold int
	// LatencyWeight is weight for latency in scoring (0.0-1.0)
	LatencyWeight float64
	// LoadWeight is weight for current load in scoring (0.0-1.0)
	LoadWeight float64
	// ErrorRateWeight is weight for error rate in scoring (0.0-1.0)
	ErrorRateWeight float64
	// RegionWeight defines regional preferences
	RegionWeight map[string]float64
}

// RegionalStats holds statistics for a region
type RegionalStats struct {
	Region          string
	TotalRequests   int64
	FailedRequests  int64
	AverageLatency  time.Duration
	LastHealthCheck time.Time
	Healthy         bool
	CurrentLoad     float64
	ErrorRate       float64
}

// NewGlobalLoadBalancer creates a new global load balancer
func NewGlobalLoadBalancer(geoRouter *GeoRouter, config *GLBConfig) *GlobalLoadBalancer {
	if config == nil {
		config = DefaultGLBConfig()
	}

	glb := &GlobalLoadBalancer{
		geoRouter:     geoRouter,
		healthChecker: NewHealthChecker(config.HealthCheckInterval),
		regionalStats: make(map[string]*RegionalStats),
		config:        config,
	}

	// Initialize regional stats
	for region := range config.RegionWeight {
		glb.regionalStats[region] = &RegionalStats{
			Region:  region,
			Healthy: true,
		}
	}

	return glb
}

// DefaultGLBConfig returns default GLB configuration
func DefaultGLBConfig() *GLBConfig {
	return &GLBConfig{
		HealthCheckInterval: 10 * time.Second,
		FailoverThreshold:   3,
		RecoveryThreshold:   5,
		LatencyWeight:       0.4,
		LoadWeight:          0.3,
		ErrorRateWeight:     0.3,
		RegionWeight: map[string]float64{
			"iad": 1.0, // Primary (US East)
			"lax": 0.5, // US West
			"fra": 0.5, // Europe
		},
	}
}

// Start begins the global load balancer
func (glb *GlobalLoadBalancer) Start(ctx context.Context) {
	logrus.Info("Starting Global Load Balancer")

	// Start health checker
	glb.healthChecker.Start(ctx, func() []*BackendLoad {
		return glb.geoRouter.GetAllBackends()
	})
}

// SelectBackend selects the best backend using global load balancing
func (glb *GlobalLoadBalancer) SelectBackend(ctx context.Context, clientIP string) (*BackendLoad, RegionalStats, error) {
	// Get client region
	clientRegion := glb.getClientRegion(clientIP)

	// Get regional backends
	regionalBackends := glb.geoRouter.GetBackendsByRegion(GeographicRegion(clientRegion))
	if len(regionalBackends) == 0 {
		// Fallback to any available backend
		regionalBackends = glb.geoRouter.GetAllBackends()
		if len(regionalBackends) == 0 {
			return nil, RegionalStats{}, ErrNoBackendsAvailable
		}
	}

	// Filter healthy backends
	healthyBackends := make([]*BackendLoad, 0)
	for _, backend := range regionalBackends {
		if backend.Healthy {
			healthyBackends = append(healthyBackends, backend)
		}
	}

	if len(healthyBackends) == 0 {
		// No healthy backends in preferred region, try all backends
		healthyBackends = regionalBackends
		if len(healthyBackends) == 0 {
			return nil, RegionalStats{}, ErrNoBackendsAvailable
		}
	}

	// Score and select backend
	selected, stats := glb.scoreAndSelect(ctx, healthyBackends, clientRegion)

	return selected, stats, nil
}

// getClientRegion determines client region from IP
func (glb *GlobalLoadBalancer) getClientRegion(clientIP string) string {
	// Use existing GeoRouter's GeoIP client
	return glb.geoRouter.LookupRegion(clientIP)
}

// scoreAndSelect scores backends and selects the best one
func (glb *GlobalLoadBalancer) scoreAndSelect(ctx context.Context, backends []*BackendLoad, preferredRegion string) (*BackendLoad, RegionalStats) {
	var best *BackendLoad
	bestScore := float64(^uint(0) >> 1) // Max int

	stats := RegionalStats{
		Region:  preferredRegion,
		Healthy: true,
	}

	for _, backend := range backends {
		if !backend.Healthy {
			continue
		}

		// Calculate composite score (lower is better)
		score := glb.calculateScore(backend, preferredRegion)

		if score < bestScore {
			bestScore = score
			best = backend
		}
	}

	if best == nil {
		stats.Healthy = false
		return nil, stats
	}

	// Update regional stats
	glb.mu.Lock()
	if regionalStats, ok := glb.regionalStats[string(best.Region)]; ok {
		regionalStats.TotalRequests++
		stats = *regionalStats
	}
	glb.mu.Unlock()

	return best, stats
}

// calculateScore calculates a composite score for a backend
func (glb *GlobalLoadBalancer) calculateScore(backend *BackendLoad, preferredRegion string) float64 {
	// Latency score (normalized to 0-100)
	latencyScore := backend.ResponseTime.Seconds() * 10

	// Load score (0-100)
	loadScore := backend.CurrentLoad * 100

	// Error rate score (0-100)
	errorScore := backend.ErrorRate * 100

	// Region preference
	regionWeight := glb.config.RegionWeight[string(backend.Region)]
	if regionWeight == 0 {
		regionWeight = 0.5
	}

	// Composite score
	score := (latencyScore * glb.config.LatencyWeight) +
		(loadScore * glb.config.LoadWeight) +
		(errorScore * glb.config.ErrorRateWeight)

	// Apply region preference (lower is better for preferred region)
	if regionWeight < 1.0 {
		score = score / regionWeight
	}

	return score
}

// GetRegionalStats returns statistics for all regions
func (glb *GlobalLoadBalancer) GetRegionalStats() map[string]*RegionalStats {
	glb.mu.RLock()
	defer glb.mu.RUnlock()

	stats := make(map[string]*RegionalStats)
	for k, v := range glb.regionalStats {
		stats[k] = v
	}
	return stats
}

// HealthChecker performs active health checks on backends
type HealthChecker struct {
	backends map[string]*HealthState
	mu       sync.RWMutex
	interval time.Duration
	client   *http.Client
	checkURL string
}

// HealthState tracks health state for a backend
type HealthState struct {
	BackendID       string
	ConsecutiveOK   int
	ConsecutiveFail int
	LastCheck       time.Time
	Latency         time.Duration
	ErrorRate       float64
	IsHealthy       bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		backends: make(map[string]*HealthState),
		interval: interval,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		checkURL: "/healthz",
	}
}

// Start begins health checking all registered backends
func (hc *HealthChecker) Start(ctx context.Context, getBackends func() []*BackendLoad) {
	ticker := time.NewTicker(hc.interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				backends := getBackends()
				for _, backend := range backends {
					hc.check(ctx, backend)
				}
			}
		}
	}()
}

// check performs a health check on a single backend
func (hc *HealthChecker) check(ctx context.Context, backend *BackendLoad) {
	url := backend.URL + hc.checkURL

	start := time.Now()
	resp, err := hc.client.Get(url)
	latency := time.Since(start)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	state, exists := hc.backends[backend.BackendID]
	if !exists {
		state = &HealthState{BackendID: backend.BackendID}
		hc.backends[backend.BackendID] = state
	}

	state.LastCheck = time.Now()
	state.Latency = latency

	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		state.ConsecutiveFail++
		state.ConsecutiveOK = 0

		if state.ConsecutiveFail >= 3 {
			state.IsHealthy = false
		}
	} else {
		state.ConsecutiveOK++
		state.ConsecutiveFail = 0

		if state.ConsecutiveOK >= 5 {
			state.IsHealthy = true
		}
	}

	// Update backend health in GeoRouter
	backend.Healthy = state.IsHealthy
	backend.LastHealthCheck = state.LastCheck
	backend.ResponseTime = latency

	if resp != nil {
		resp.Body.Close()
	}
}

// GetHealthState returns the health state for a backend
func (hc *HealthChecker) GetHealthState(backendID string) *HealthState {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.backends[backendID]
}

// GetAllHealthStates returns all health states
func (hc *HealthChecker) GetAllHealthStates() map[string]*HealthState {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	states := make(map[string]*HealthState)
	for k, v := range hc.backends {
		states[k] = v
	}
	return states
}

// ErrNoBackendsAvailable is returned when no backends are available
var ErrNoBackendsAvailable = &NoBackendsError{}

type NoBackendsError struct{}

func (e *NoBackendsError) Error() string {
	return "no backends available"
}

// GetBackendsByRegion returns backends for a specific region
func (geo *GeoRouter) GetBackendsByRegion(region GeographicRegion) []*BackendLoad {
	geo.mu.RLock()
	defer geo.mu.RUnlock()

	return geo.regions[region]
}

// GetAllBackends returns all registered backends
func (geo *GeoRouter) GetAllBackends() []*BackendLoad {
	geo.mu.RLock()
	defer geo.mu.RUnlock()

	backends := make([]*BackendLoad, 0, len(geo.backends))
	for _, backend := range geo.backends {
		backends = append(backends, backend)
	}
	return backends
}

// LookupRegion looks up the region for an IP address using GeoIP
func (geo *GeoRouter) LookupRegion(ip string) string {
	geo.mu.RLock()
	defer geo.mu.RUnlock()

	if geo.geoIPClient == nil {
		return string(RegionNorthAmerica) // Default to North America
	}

	country, err := geo.geoIPClient.LookupCountry(ip)
	if err != nil {
		return string(RegionNorthAmerica)
	}

	// Map country code to geographic region
	return mapCountryToRegion(country)
}

// mapCountryToRegion maps a country code to a geographic region
func mapCountryToRegion(country string) string {
	switch country {
	case "US", "CA", "MX":
		return string(RegionNorthAmerica)
	case "GB", "DE", "FR", "IT", "ES", "NL", "BE", "SE", "NO", "DK", "FI", "AT", "CH", "PL", "PT", "IE":
		return string(RegionEurope)
	case "JP", "CN", "KR", "IN", "SG", "HK", "TW", "MY", "TH", "VN", "ID":
		return string(RegionAsia)
	case "BR", "AR", "CO", "CL", "PE":
		return string(RegionSouthAmerica)
	case "AU", "NZ":
		return string(RegionOceania)
	case "ZA", "EG", "NG", "KE", "MA":
		return string(RegionAfrica)
	default:
		return string(RegionNorthAmerica)
	}
}
