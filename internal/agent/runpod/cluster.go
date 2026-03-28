package runpod

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Cluster represents a regional GPU cluster
type Cluster struct {
	ID            string                   // Unique cluster identifier
	Name          string                   // Human-readable cluster name
	Region        string                   // Geographic region (e.g., "us-east-1")
	GPUType       string                   // GPU type (e.g., "NVIDIA A100")
	Pool          *InstancePool            // Instance pool for this cluster
	Config        *RegionConfig            // Region-specific configuration
	HealthStatus  ClusterHealth            // Current health status
	TotalRequests int64                    // Total requests handled
	AvgLatencyMs  float64                  // Average request latency in ms
	mu            sync.RWMutex
}

// ClusterHealth represents the health status of a cluster
type ClusterHealth struct {
	Status      string    // "healthy", "degraded", "unhealthy"
	LastCheck   time.Time // Last health check timestamp
	LatencyMs   float64   // Current estimated latency
	ErrorRate   float64   // Current error rate (0-1)
	HealthyInstances int   // Number of healthy instances
	TotalInstances int    // Total instances
}

// ClusterManager manages multiple regional GPU clusters
type ClusterManager struct {
	config     *Config              // Global configuration
	clusters   map[string]*Cluster  // clusters by ID
	clients    map[string]*RunPodClient // clients by region
	mu         sync.RWMutex
	selector   *ClusterSelector     // Load balancing strategy
}

// ClusterSelector handles load balancing across clusters
type ClusterSelector struct {
	strategy LoadBalancingStrategy
	mu       sync.RWMutex
}

// LoadBalancingStrategy defines how to select clusters
type LoadBalancingStrategy string

const (
	// LeastLoaded selects the cluster with the fewest running instances
	LeastLoaded LoadBalancingStrategy = "least_loaded"
	// WeightedRoundRobin selects clusters based on configured weights
	WeightedRoundRobin LoadBalancingStrategy = "weighted_round_robin"
	// LatencyBased selects the cluster with the lowest latency
	LatencyBased LoadBalancingStrategy = "latency_based"
	// GeoAware selects the closest region first, then falls back to others
	GeoAware LoadBalancingStrategy = "geo_aware"
)

// NewClusterManager creates a new cluster manager
func NewClusterManager(config *Config) *ClusterManager {
	cm := &ClusterManager{
		config:   config,
		clusters: make(map[string]*Cluster),
		clients:  make(map[string]*RunPodClient),
	}

	// Set default load balancing strategy
	cm.selector = &ClusterSelector{
		strategy: LeastLoaded,
	}

	// Initialize clusters from config
	if len(config.Regions) > 0 {
		for _, rc := range config.Regions {
			if rc.Enabled {
				cm.addClusterFromRegionConfig(&rc)
			}
		}
	}

	return cm
}

// addClusterFromRegionConfig creates a cluster from a region config
func (cm *ClusterManager) addClusterFromRegionConfig(rc *RegionConfig) error {
	clusterID := fmt.Sprintf("%s-%s", rc.Region, rc.GPUType)

	// Create RunPod client for this region
	client := NewRunPodClient(cm.config.RunPodAPIKey, rc.BaseURL)
	if rc.BaseURL == "" {
		client = NewRunPodClient(cm.config.RunPodAPIKey, cm.config.RunPodAPIBaseURL)
	}
	cm.clients[rc.Region] = client

	// Create regional config
	regConfig := &Config{
		Mode:                cm.config.Mode,
		RunPodAPIKey:        cm.config.RunPodAPIKey,
		RunPodAPIBaseURL:    rc.BaseURL,
		GPUType:             rc.GPUType,
		GPUCount:            rc.GPUCount,
		ContainerImage:      rc.ContainerImage,
		InstanceTimeout:     cm.config.InstanceTimeout,
		IdleTimeout:         cm.config.IdleTimeout,
		ProvisioningTimeout: cm.config.ProvisioningTimeout,
		MaxRetries:          cm.config.MaxRetries,
		HTTPHost:            cm.config.HTTPHost,
		HTTPPort:            cm.config.HTTPPort,
		HealthCheckPath:     cm.config.HealthCheckPath,
		ModelName:           cm.config.ModelName,
		MaxInstances:        rc.MaxInstances,
		MinInstances:        rc.MinInstances,
	}

	// Create instance pool for this region/cluster
	pool := NewRegionalPool(regConfig, client, rc.Region, clusterID)

	cluster := &Cluster{
		ID:           clusterID,
		Name:         fmt.Sprintf("cluster-%s-%s", rc.Region, rc.GPUType),
		Region:       rc.Region,
		GPUType:      rc.GPUType,
		Pool:         pool,
		Config:       rc,
		HealthStatus: ClusterHealth{Status: "unknown"},
	}

	cm.mu.Lock()
	cm.clusters[clusterID] = cluster
	cm.mu.Unlock()

	return nil
}

// GetCluster returns a cluster by ID
func (cm *ClusterManager) GetCluster(clusterID string) (*Cluster, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cluster, ok := cm.clusters[clusterID]
	return cluster, ok
}

// GetClusterByRegion returns the cluster for a specific region and GPU type
func (cm *ClusterManager) GetClusterByRegion(region, gpuType string) (*Cluster, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	clusterID := fmt.Sprintf("%s-%s", region, gpuType)
	cluster, ok := cm.clusters[clusterID]
	return cluster, ok
}

// ListClusters returns all clusters
func (cm *ClusterManager) ListClusters() []*Cluster {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusters := make([]*Cluster, 0, len(cm.clusters))
	for _, c := range cm.clusters {
		clusters = append(clusters, c)
	}
	return clusters
}

// SelectCluster selects a cluster based on the load balancing strategy
func (cm *ClusterManager) SelectCluster(preferredRegion string) (*Cluster, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.clusters) == 0 {
		return nil, fmt.Errorf("no clusters available")
	}

	cm.selector.mu.RLock()
	strategy := cm.selector.strategy
	cm.selector.mu.RUnlock()

	switch strategy {
	case LeastLoaded:
		return cm.selectLeastLoaded(preferredRegion)
	case WeightedRoundRobin:
		return cm.selectWeightedRoundRobin(preferredRegion)
	case LatencyBased:
		return cm.selectLatencyBased(preferredRegion)
	case GeoAware:
		return cm.selectGeoAware(preferredRegion)
	default:
		return cm.selectLeastLoaded(preferredRegion)
	}
}

// selectLeastLoaded selects the cluster with the fewest running instances
func (cm *ClusterManager) selectLeastLoaded(preferredRegion string) (*Cluster, error) {
	var preferred, fallback *Cluster
	var minLoad = int(^uint(0) >> 1) // max int

	for _, cluster := range cm.clusters {
		_, running, _, _ := cluster.Pool.GetStats()

		// Prefer clusters in the preferred region
		if preferredRegion != "" && cluster.Region == preferredRegion {
			if running < minLoad {
				preferred = cluster
				minLoad = running
			}
		}

		if running < minLoad {
			fallback = cluster
			minLoad = running
		}
	}

	if preferred != nil {
		return preferred, nil
	}
	return fallback, nil
}

// selectWeightedRoundRobin selects clusters based on configured weights
func (cm *ClusterManager) selectWeightedRoundRobin(preferredRegion string) (*Cluster, error) {
	var totalWeight int
	var preferredClusters []*Cluster
	var allClusters []*Cluster

	for _, cluster := range cm.clusters {
		weight := cluster.Config.Weight
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
		allClusters = append(allClusters, cluster)

		if preferredRegion == "" || cluster.Region == preferredRegion {
			preferredClusters = append(preferredClusters, cluster)
		}
	}

	clusters := preferredClusters
	if len(clusters) == 0 {
		clusters = allClusters
	}

	if len(clusters) == 0 {
		return nil, fmt.Errorf("no clusters available")
	}

	// Random selection based on weights
	r := rand.Intn(totalWeight)
	current := 0
	for _, cluster := range clusters {
		weight := cluster.Config.Weight
		if weight <= 0 {
			weight = 1
		}
		current += weight
		if r < current {
			return cluster, nil
		}
	}

	return clusters[0], nil
}

// selectLatencyBased selects the cluster with the lowest latency
func (cm *ClusterManager) selectLatencyBased(preferredRegion string) (*Cluster, error) {
	var lowestLatency float64 = -1
	var preferred, fallback *Cluster

	for _, cluster := range cm.clusters {
		latency := cluster.HealthStatus.LatencyMs

		// Skip unhealthy clusters
		if cluster.HealthStatus.Status == "unhealthy" {
			continue
		}

		// If no latency data, use default high value
		if latency <= 0 {
			latency = 1000 // 1 second default
		}

		if preferredRegion != "" && cluster.Region == preferredRegion {
			if preferred == nil || latency < lowestLatency {
				preferred = cluster
				lowestLatency = latency
			}
		}

		if fallback == nil || latency < lowestLatency {
			fallback = cluster
			lowestLatency = latency
		}
	}

	if preferred != nil {
		return preferred, nil
	}
	return fallback, nil
}

// selectGeoAware selects the closest region first, then falls back to others
func (cm *ClusterManager) selectGeoAware(preferredRegion string) (*Cluster, error) {
	// First try preferred region
	if preferredRegion != "" {
		for _, cluster := range cm.clusters {
			if cluster.Region == preferredRegion && cluster.HealthStatus.Status != "unhealthy" {
				return cluster, nil
			}
		}
	}

	// Fall back to least loaded among remaining
	return cm.selectLeastLoaded("")
}

// SetLoadBalancingStrategy sets the load balancing strategy
func (cm *ClusterManager) SetLoadBalancingStrategy(strategy LoadBalancingStrategy) {
	cm.selector.mu.Lock()
	defer cm.selector.mu.Unlock()
	cm.selector.strategy = strategy
}

// UpdateClusterHealth updates the health status of a cluster
func (cm *ClusterManager) UpdateClusterHealth(clusterID string, health ClusterHealth) error {
	cm.mu.RLock()
	cluster, ok := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	cluster.mu.Lock()
	cluster.HealthStatus = health
	cluster.mu.Unlock()

	return nil
}

// GetClusterStats returns aggregated statistics for all clusters
func (cm *ClusterManager) GetClusterStats() ClusterStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := ClusterStats{
		TotalClusters:    len(cm.clusters),
		HealthyClusters:  0,
		TotalInstances:    0,
		RunningInstances:  0,
		IdleInstances:     0,
		FailedInstances:   0,
		TotalRequests:    0,
		RegionStats:       make(map[string]RegionStats),
	}

	for _, cluster := range cm.clusters {
		cluster.mu.RLock()
		total, running, idle, failed := cluster.Pool.GetStats()
		cluster.mu.RUnlock()

		stats.TotalInstances += total
		stats.RunningInstances += running
		stats.IdleInstances += idle
		stats.FailedInstances += failed
		stats.TotalRequests += cluster.TotalRequests

		if cluster.HealthStatus.Status == "healthy" {
			stats.HealthyClusters++
		}

		rs := stats.RegionStats[cluster.Region]
		rs.TotalInstances += total
		rs.RunningInstances += running
		rs.IdleInstances += idle
		rs.FailedInstances += failed
		stats.RegionStats[cluster.Region] = rs
	}

	return stats
}

// ClusterStats holds aggregated cluster statistics
type ClusterStats struct {
	TotalClusters    int              // Total number of clusters
	HealthyClusters  int              // Number of healthy clusters
	TotalInstances   int              // Total number of instances across all clusters
	RunningInstances int              // Number of running instances
	IdleInstances    int              // Number of idle instances
	FailedInstances  int              // Number of failed instances
	TotalRequests    int64            // Total requests handled
	RegionStats      map[string]RegionStats // Per-region statistics
}

// RegionStats holds statistics for a specific region
type RegionStats struct {
	TotalInstances   int // Total instances in region
	RunningInstances int // Running instances in region
	IdleInstances    int // Idle instances in region
	FailedInstances  int // Failed instances in region
}

// StartHealthChecker starts a background health check routine
func (cm *ClusterManager) StartHealthChecker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cm.checkClustersHealth(ctx)
			}
		}
	}()
}

// checkClustersHealth performs health checks on all clusters
func (cm *ClusterManager) checkClustersHealth(ctx context.Context) {
	cm.mu.RLock()
	clusters := make([]*Cluster, 0, len(cm.clusters))
	for _, c := range cm.clusters {
		clusters = append(clusters, c)
	}
	cm.mu.RUnlock()

	for _, cluster := range clusters {
		health := cm.performHealthCheck(ctx, cluster)
		cluster.mu.Lock()
		cluster.HealthStatus = health
		cluster.mu.Unlock()
	}
}

// performHealthCheck checks the health of a single cluster
func (cm *ClusterManager) performHealthCheck(ctx context.Context, cluster *Cluster) ClusterHealth {
	health := ClusterHealth{
		LastCheck: time.Now(),
	}

	total, running, _, failed := cluster.Pool.GetStats()
	health.TotalInstances = total
	health.HealthyInstances = running

	if failed > 0 && running == 0 {
		health.Status = "unhealthy"
		health.ErrorRate = 1.0
	} else if failed > 0 || running < total/2 {
		health.Status = "degraded"
		if total > 0 {
			health.ErrorRate = float64(failed) / float64(total)
		}
	} else {
		health.Status = "healthy"
		health.ErrorRate = 0
	}

	// Estimate latency based on instance state
	// In production, this would be based on actual health check pings
	if running > 0 {
		health.LatencyMs = 50 + float64(failed)*20 // Simplified estimation
	} else {
		health.LatencyMs = 5000 // High latency if no running instances
	}

	return health
}

// ProvisionInstance provisions a new instance in the specified cluster
func (cm *ClusterManager) ProvisionInstance(ctx context.Context, clusterID string) (*GPUInstance, error) {
	cm.mu.RLock()
	cluster, ok := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	return cluster.Pool.Provision(ctx)
}

// ProvisionInRegion provisions a new instance in the specified region
func (cm *ClusterManager) ProvisionInRegion(ctx context.Context, region string) (*GPUInstance, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var cluster *Cluster
	for _, c := range cm.clusters {
		if c.Region == region {
			cluster = c
			break
		}
	}

	if cluster == nil {
		return nil, fmt.Errorf("no cluster found for region: %s", region)
	}

	return cluster.Pool.Provision(ctx)
}

// TerminateInstance terminates an instance in a specific cluster
func (cm *ClusterManager) TerminateInstance(ctx context.Context, clusterID, instanceID string) error {
	cm.mu.RLock()
	cluster, ok := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	return cluster.Pool.Terminate(ctx, instanceID)
}

// AddCluster adds a new cluster dynamically
func (cm *ClusterManager) AddCluster(clusterID, region, gpuType string, config *RegionConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clusters[clusterID]; exists {
		return fmt.Errorf("cluster already exists: %s", clusterID)
	}

	// Create client for this region
	client := NewRunPodClient(cm.config.RunPodAPIKey, cm.config.RunPodAPIBaseURL)
	cm.clients[region] = client

	// Create regional config
	regConfig := &Config{
		Mode:                cm.config.Mode,
		RunPodAPIKey:        cm.config.RunPodAPIKey,
		GPUType:             gpuType,
		GPUCount:            config.GPUCount,
		ContainerImage:      config.ContainerImage,
		InstanceTimeout:     cm.config.InstanceTimeout,
		IdleTimeout:         cm.config.IdleTimeout,
		ProvisioningTimeout: cm.config.ProvisioningTimeout,
		MaxInstances:        config.MaxInstances,
		MinInstances:        config.MinInstances,
		HTTPHost:            cm.config.HTTPHost,
		HTTPPort:            cm.config.HTTPPort,
		HealthCheckPath:     cm.config.HealthCheckPath,
		ModelName:           cm.config.ModelName,
	}

	// Create instance pool
	pool := NewRegionalPool(regConfig, client, region, clusterID)

	cluster := &Cluster{
		ID:           clusterID,
		Name:         fmt.Sprintf("cluster-%s-%s", region, gpuType),
		Region:       region,
		GPUType:      gpuType,
		Pool:         pool,
		Config:       config,
		HealthStatus: ClusterHealth{Status: "unknown"},
	}

	cm.clusters[clusterID] = cluster
	log.Printf("Added new cluster: %s in region %s", clusterID, region)

	return nil
}

// RemoveCluster removes a cluster
func (cm *ClusterManager) RemoveCluster(clusterID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cluster, ok := cm.clusters[clusterID]
	if !ok {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	delete(cm.clusters, clusterID)
	log.Printf("Removed cluster: %s", clusterID)

	_ = cluster // cluster will be garbage collected
	return nil
}

// GetClient returns the RunPod client for a specific region
func (cm *ClusterManager) GetClient(region string) (*RunPodClient, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, ok := cm.clients[region]
	return client, ok
}
