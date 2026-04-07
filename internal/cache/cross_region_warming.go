package cache

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CrossRegionWarmingConfig holds configuration for cross-region cache warming
type CrossRegionWarmingConfig struct {
	Enabled              bool          // Whether warming is enabled
	WarmOnDeploy         bool          // Warm cache when function is deployed
	WarmOnHighPopularity bool          // Warm when function reaches popularity threshold
	WarmInterval         time.Duration // How often to run warming cycle
	Timeout              time.Duration // HTTP timeout for warming requests
	RetryAttempts        int           // Number of retry attempts per region
	Concurrency          int           // Number of concurrent warming workers
	Regions              []RegionConfig // Target regions to warm
	MinPopularityScore   int           // Minimum popularity to trigger warming
}

// RegionConfig represents a target region for cache warming
type RegionConfig struct {
	Name        string // Region identifier (e.g., "eu-west-1", "ap-southeast-1")
	Endpoint    string // Full URL endpoint (e.g., "https://eu.functionfly.com")
	Weight      int    // Priority weight (higher = warmed first)
	HealthCheck string // Health check URL
	Active      bool   // Whether this region is active
}

// DefaultCrossRegionWarmingConfig returns sensible defaults
type DefaultCrossRegionWarmingConfig struct{}

func (DefaultCrossRegionWarmingConfig) Get() *CrossRegionWarmingConfig {
	return &CrossRegionWarmingConfig{
		Enabled:              true,
		WarmOnDeploy:         true,
		WarmOnHighPopularity: true,
		WarmInterval:         10 * time.Minute,
		Timeout:              30 * time.Second,
		RetryAttempts:        3,
		Concurrency:          5,
		MinPopularityScore:   100,
		Regions: []RegionConfig{
			{
				Name:        "us-east-1",
				Endpoint:    "https://api.functionfly.com",
				Weight:      100,
				HealthCheck: "https://api.functionfly.com/healthz",
				Active:      true,
			},
			{
				Name:        "eu-west-1",
				Endpoint:    "https://eu.functionfly.com",
				Weight:      80,
				HealthCheck: "https://eu.functionfly.com/healthz",
				Active:      true,
			},
			{
				Name:        "ap-southeast-1",
				Endpoint:    "https://apac.functionfly.com",
				Weight:      60,
				HealthCheck: "https://apac.functionfly.com/healthz",
				Active:      true,
			},
		},
	}
}

// CrossRegionWarmer manages cross-region cache warming
type CrossRegionWarmer struct {
	config      *CrossRegionWarmingConfig
	httpClient  *http.Client
	registryCache *RegistryRedisCache
	edgeCache   *EdgeCacheService

	// Region health tracking
	regionHealth   map[string]bool
	regionHealthMu sync.RWMutex

	// Warming queue
	warmingQueue   chan WarmingRequest
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// WarmingRequest represents a function to warm across regions
type WarmingRequest struct {
	FunctionID   uuid.UUID
	FunctionName string
	Author       string
	Version      string
	Regions      []string // Specific regions to warm (empty = all)
	Priority     int      // Warming priority (higher = warmed first)
	RetryCount   int      // Current retry count
	Trigger      string   // What triggered the warming (e.g., "deploy", "popularity", "scheduled")
	Timestamp    time.Time
}

// NewCrossRegionWarmer creates a new cross-region cache warmer
func NewCrossRegionWarmer(config *CrossRegionWarmingConfig, registryCache *RegistryRedisCache, edgeCache *EdgeCacheService) *CrossRegionWarmer {
	if config == nil {
		config = DefaultCrossRegionWarmingConfig{}.Get()
	}

	warmer := &CrossRegionWarmer{
		config:         config,
		httpClient:     &http.Client{Timeout: config.Timeout},
		registryCache:  registryCache,
		edgeCache:      edgeCache,
		regionHealth:   make(map[string]bool),
		warmingQueue:   make(chan WarmingRequest, 1000),
		stopCh:         make(chan struct{}),
	}

	// Initialize region health
	for _, region := range config.Regions {
		warmer.regionHealth[region.Name] = region.Active
	}

	if config.Enabled {
		// Start warming workers
		for i := 0; i < config.Concurrency; i++ {
			warmer.wg.Add(1)
			go warmer.warmingWorker(i)
		}

		// Start health check loop
		warmer.wg.Add(1)
		go warmer.healthCheckLoop()

		// Start scheduled warming loop
		warmer.wg.Add(1)
		go warmer.scheduledWarmingLoop()
	}

	return warmer
}

// Stop gracefully stops the warmer
func (w *CrossRegionWarmer) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

// QueueWarmingRequest adds a function to the warming queue
func (w *CrossRegionWarmer) QueueWarmingRequest(req WarmingRequest) error {
	if !w.config.Enabled {
		return nil
	}

	// Set defaults
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	select {
	case w.warmingQueue <- req:
		logrus.Debugf("Queued warming request for function %s in %d regions",
			req.FunctionID, len(req.Regions))
		return nil
	default:
		return fmt.Errorf("warming queue is full")
	}
}

// WarmFunctionOnDeploy warms a function across regions after deployment
func (w *CrossRegionWarmer) WarmFunctionOnDeploy(ctx context.Context, functionID uuid.UUID, functionName, author, version string) error {
	if !w.config.Enabled || !w.config.WarmOnDeploy {
		return nil
	}

	req := WarmingRequest{
		FunctionID:   functionID,
		FunctionName: functionName,
		Author:       author,
		Version:      version,
		Priority:     100, // High priority for new deployments
		Trigger:      "deploy",
		Timestamp:    time.Now(),
	}

	return w.QueueWarmingRequest(req)
}

// WarmFunctionOnPopularity warms a function when it becomes popular
func (w *CrossRegionWarmer) WarmFunctionOnPopularity(ctx context.Context, functionID uuid.UUID, functionName, author string, popularityScore int) error {
	if !w.config.Enabled || !w.config.WarmOnHighPopularity {
		return nil
	}

	// Only warm if popularity exceeds threshold
	if popularityScore < w.config.MinPopularityScore {
		return nil
	}

	req := WarmingRequest{
		FunctionID:   functionID,
		FunctionName: functionName,
		Author:       author,
		Priority:     popularityScore, // Priority based on popularity
		Trigger:      "popularity",
		Timestamp:    time.Now(),
	}

	return w.QueueWarmingRequest(req)
}

// warmingWorker processes warming requests from the queue
func (w *CrossRegionWarmer) warmingWorker(id int) {
	defer w.wg.Done()

	logrus.Infof("Starting warming worker %d", id)

	for {
		select {
		case <-w.stopCh:
			logrus.Infof("Warming worker %d stopping", id)
			return

		case req := <-w.warmingQueue:
			if err := w.processWarmingRequest(req); err != nil {
				logrus.Warnf("Warming worker %d failed to warm function %s: %v",
					id, req.FunctionID, err)

				// Retry if attempts remain
				if req.RetryCount < w.config.RetryAttempts {
					req.RetryCount++
					time.Sleep(time.Duration(req.RetryCount) * time.Second)

					select {
					case w.warmingQueue <- req:
						// Requeued
					default:
						logrus.Warnf("Failed to requeue warming request for %s", req.FunctionID)
					}
				}
			}
		}
	}
}

// processWarmingRequest warms a single function across regions
func (w *CrossRegionWarmer) processWarmingRequest(req WarmingRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.Timeout*2)
	defer cancel()

	// Determine which regions to warm
	regions := w.getTargetRegions(req.Regions)
	if len(regions) == 0 {
		return fmt.Errorf("no active regions available for warming")
	}

	logrus.Infof("Warming function %s/%s@%s across %d regions (trigger: %s, priority: %d)",
		req.Author, req.FunctionName, req.Version, len(regions), req.Trigger, req.Priority)

	// Warm each region concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(regions))

	for _, region := range regions {
		wg.Add(1)
		go func(r RegionConfig) {
			defer wg.Done()

			if err := w.warmRegion(ctx, req, r); err != nil {
				errChan <- fmt.Errorf("region %s: %w", r.Name, err)
			}
		}(region)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("warming failed in %d/%d regions: %v",
			len(errors), len(regions), errors)
	}

	// Record success
	w.recordWarmingSuccess(req, regions)

	return nil
}

// warmRegion warms a single region by making a warming request
func (w *CrossRegionWarmer) warmRegion(ctx context.Context, req WarmingRequest, region RegionConfig) error {
	// Build warming URL
	warmURL := fmt.Sprintf("%s/v1/registry/%s/%s/warm",
		region.Endpoint, req.Author, req.FunctionName)
	if req.Version != "" {
		warmURL = fmt.Sprintf("%s?version=%s", warmURL, req.Version)
	}

	// Make warming request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", warmURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add warming headers
	httpReq.Header.Set("X-Cache-Warming", "true")
	httpReq.Header.Set("X-Source-Region", "primary")
	httpReq.Header.Set("X-Warming-Priority", fmt.Sprintf("%d", req.Priority))
	httpReq.Header.Set("X-Warming-Trigger", req.Trigger)

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("warming request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and discard body
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("warming returned status %d", resp.StatusCode)
	}

	logrus.Debugf("Successfully warmed %s/%s@%s in region %s",
		req.Author, req.FunctionName, req.Version, region.Name)

	return nil
}

// getTargetRegions returns active regions based on request specification
func (w *CrossRegionWarmer) getTargetRegions(requested []string) []RegionConfig {
	w.regionHealthMu.RLock()
	defer w.regionHealthMu.RUnlock()

	var regions []RegionConfig
	for _, region := range w.config.Regions {
		// Skip inactive regions
		if !region.Active {
			continue
		}

		// Skip unhealthy regions
		if healthy, ok := w.regionHealth[region.Name]; !ok || !healthy {
			continue
		}

		// If specific regions requested, filter
		if len(requested) > 0 {
			found := false
			for _, r := range requested {
				if r == region.Name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		regions = append(regions, region)
	}

	return regions
}

// healthCheckLoop periodically checks region health
func (w *CrossRegionWarmer) healthCheckLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial health check
	w.checkAllRegionsHealth()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.checkAllRegionsHealth()
		}
	}
}

// checkAllRegionsHealth performs health checks on all configured regions
func (w *CrossRegionWarmer) checkAllRegionsHealth() {
	for _, region := range w.config.Regions {
		if region.HealthCheck == "" {
			continue
		}

		healthy := w.checkRegionHealth(region)

		w.regionHealthMu.Lock()
		oldHealth := w.regionHealth[region.Name]
		w.regionHealth[region.Name] = healthy
		w.regionHealthMu.Unlock()

		// Log status changes
		if oldHealth != healthy {
			if healthy {
				logrus.Infof("Region %s is now healthy", region.Name)
			} else {
				logrus.Warnf("Region %s is now unhealthy", region.Name)
			}
		}
	}
}

// checkRegionHealth checks health of a single region
func (w *CrossRegionWarmer) checkRegionHealth(region RegionConfig) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", region.HealthCheck, nil)
	if err != nil {
		return false
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// scheduledWarmingLoop periodically warms popular functions
func (w *CrossRegionWarmer) scheduledWarmingLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.WarmInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.runScheduledWarming()
		}
	}
}

// runScheduledWarming warms the most popular functions across regions
func (w *CrossRegionWarmer) runScheduledWarming() {
	// Get edge cache candidates (popular functions)
	candidates, err := w.edgeCache.GetEdgeCacheCandidates(context.Background())
	if err != nil {
		logrus.Warnf("Failed to get edge cache candidates for scheduled warming: %v", err)
		return
	}

	// Queue warming for top candidates
	warmed := 0
	for _, candidate := range candidates {
		// Only warm if popularity is high enough
		if candidate.PopularityScore < w.config.MinPopularityScore {
			continue
		}

		req := WarmingRequest{
			FunctionID:   candidate.FunctionID,
			FunctionName: candidate.FunctionName,
			Author:       candidate.Author,
			Version:      candidate.Version,
			Priority:     candidate.PopularityScore,
			Trigger:      "scheduled",
			Timestamp:    time.Now(),
		}

		if err := w.QueueWarmingRequest(req); err != nil {
			logrus.Warnf("Failed to queue scheduled warming for %s: %v",
				candidate.FunctionID, err)
			continue
		}

		warmed++
		if warmed >= 50 { // Limit to 50 functions per cycle
			break
		}
	}

	if warmed > 0 {
		logrus.Infof("Queued %d functions for scheduled warming", warmed)
	}
}

// recordWarmingSuccess records successful warming in cache
func (w *CrossRegionWarmer) recordWarmingSuccess(req WarmingRequest, regions []RegionConfig) {
	ctx := context.Background()

	// Record in Redis for monitoring
	key := fmt.Sprintf("cache:warming:%s:last_success", req.FunctionID)
	w.registryCache.SetJSONWithTTL(ctx, key, time.Now().Format(time.RFC3339), 24*time.Hour)

	// Record warmed regions
	regionNames := make([]string, len(regions))
	for i, r := range regions {
		regionNames[i] = r.Name
	}

	metaKey := fmt.Sprintf("cache:warming:%s:regions", req.FunctionID)
	w.registryCache.SetJSONWithTTL(ctx, metaKey, regionNames, 24*time.Hour)

	logrus.Infof("Successfully warmed function %s in regions: %v",
		req.FunctionID, regionNames)
}

// GetWarmingStatus returns the warming status for a function
func (w *CrossRegionWarmer) GetWarmingStatus(ctx context.Context, functionID uuid.UUID) (*WarmingStatus, error) {
	status := &WarmingStatus{
		FunctionID: functionID,
	}

	// Get last success time
	lastSuccessKey := fmt.Sprintf("cache:warming:%s:last_success", functionID)
	data, err := w.registryCache.Get(ctx, lastSuccessKey)
	if err == nil {
		lastSuccess := string(data)
		t, _ := time.Parse(time.RFC3339, lastSuccess)
		status.LastWarmed = t
		status.IsWarmed = time.Since(t) < w.config.WarmInterval*2
	}

	// Get warmed regions
	regionsKey := fmt.Sprintf("cache:warming:%s:regions", functionID)
	var regions []string
	if err := w.registryCache.GetJSON(ctx, regionsKey, &regions); err == nil {
		status.WarmedRegions = regions
	}

	return status, nil
}

// WarmingStatus represents the warming status of a function
type WarmingStatus struct {
	FunctionID    uuid.UUID `json:"function_id"`
	IsWarmed      bool      `json:"is_warmed"`
	LastWarmed    time.Time `json:"last_warmed"`
	WarmedRegions []string  `json:"warmed_regions"`
}

// IsRegionWarmed checks if a specific region has been warmed
func (s *WarmingStatus) IsRegionWarmed(region string) bool {
	for _, r := range s.WarmedRegions {
		if r == region {
			return true
		}
	}
	return false
}

// GetRegionHealth returns the current health status of all regions
func (w *CrossRegionWarmer) GetRegionHealth() map[string]bool {
	w.regionHealthMu.RLock()
	defer w.regionHealthMu.RUnlock()

	health := make(map[string]bool)
	for k, v := range w.regionHealth {
		health[k] = v
	}
	return health
}

// WarmAllActiveFunctions triggers warming for all currently active edge-cached functions
func (w *CrossRegionWarmer) WarmAllActiveFunctions(ctx context.Context) error {
	candidates, err := w.edgeCache.GetEdgeCacheCandidates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get edge cache candidates: %w", err)
	}

	queued := 0
	for _, candidate := range candidates {
		req := WarmingRequest{
			FunctionID:   candidate.FunctionID,
			FunctionName: candidate.FunctionName,
			Author:       candidate.Author,
			Version:      candidate.Version,
			Priority:     candidate.PopularityScore,
			Trigger:      "manual",
			Timestamp:    time.Now(),
		}

		if err := w.QueueWarmingRequest(req); err == nil {
			queued++
		}
	}

	logrus.Infof("Manually queued %d functions for warming", queued)
	return nil
}

// GetStats returns warming statistics
func (w *CrossRegionWarmer) GetStats() map[string]interface{} {
	w.regionHealthMu.RLock()
	defer w.regionHealthMu.RUnlock()

	healthyRegions := 0
	for _, h := range w.regionHealth {
		if h {
			healthyRegions++
		}
	}

	return map[string]interface{}{
		"enabled":         w.config.Enabled,
		"queue_length":    len(w.warmingQueue),
		"total_regions":   len(w.config.Regions),
		"healthy_regions": healthyRegions,
		"config":          w.config,
	}
}
