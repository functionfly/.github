package routing

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

// GeographicRegion represents a geographic region
type GeographicRegion string

const (
	RegionNorthAmerica GeographicRegion = "north-america"
	RegionEurope       GeographicRegion = "europe"
	RegionAsia         GeographicRegion = "asia"
	RegionSouthAmerica GeographicRegion = "south-america"
	RegionAfrica       GeographicRegion = "africa"
	RegionOceania      GeographicRegion = "oceania"
)

// BackendLoad represents the current load of a backend
type BackendLoad struct {
	BackendID       string
	Region          GeographicRegion
	URL             string
	CurrentLoad     float64 // 0.0 to 1.0
	ResponseTime    time.Duration
	ErrorRate       float64
	LastHealthCheck time.Time
	Healthy         bool
	Capacity        int // Maximum concurrent executions
}

// BackendRepository defines the backend operations needed by GeoRouter
type BackendRepository interface {
	GetBackendByID(id uuid.UUID) (*storage.Backend, error)
}

// GeoRouter manages geographic distribution and load balancing of execution backends
type GeoRouter struct {
	backends       map[string]*BackendLoad
	regions        map[GeographicRegion][]*BackendLoad
	mu             sync.RWMutex
	geoIPClient    *GeoIPClient
	repo           BackendRepository
	config         *GeoRouterConfig
	healthClient   *http.Client
}

// HealthCheckConfig holds configuration for backend health checks
type HealthCheckConfig struct {
	Timeout      time.Duration
	Interval     time.Duration
	MaxRetries   int
	HealthPath   string // Path to check for health (e.g., "/health", "/status")
	ExpectedCode int    // Expected HTTP status code (usually 200)
}

// GeoRouterConfig holds configuration for the geographic router
type GeoRouterConfig struct {
	// GeoIP configuration
	GeoIPDatabasePath string        // Path to store GeoIP database
	GeoIPUpdateInterval time.Duration // How often to check for database updates

	// Health check configuration
	HealthCheckConfig *HealthCheckConfig

	// Load balancing configuration
	DefaultBackendCapacity int           // Default capacity for new backends
	LoadBalancingWeights   LoadBalancingWeights // Weights for load balancing algorithm
}

// LoadBalancingWeights defines weights for the load balancing algorithm
type LoadBalancingWeights struct {
	LoadWeight       float64 // Weight for current load (0.0-1.0)
	ResponseTimeWeight float64 // Weight for response time
	ErrorRateWeight  float64 // Weight for error rate
}

// DefaultGeoRouterConfig returns default configuration for the geographic router
func DefaultGeoRouterConfig() *GeoRouterConfig {
	return &GeoRouterConfig{
		GeoIPDatabasePath:   filepath.Join("data", "geoip", "GeoLite2-Country.mmdb"),
		GeoIPUpdateInterval: 24 * time.Hour, // Update daily
		HealthCheckConfig:   DefaultHealthCheckConfig(),
		DefaultBackendCapacity: 100,
		LoadBalancingWeights: LoadBalancingWeights{
			LoadWeight:         0.5,
			ResponseTimeWeight: 0.3,
			ErrorRateWeight:    0.2,
		},
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *GeoRouterConfig {
	config := DefaultGeoRouterConfig()

	// GeoIP configuration
	if path := os.Getenv("GEOROUTER_GEOIP_DB_PATH"); path != "" {
		config.GeoIPDatabasePath = path
	}

	if interval := os.Getenv("GEOROUTER_GEOIP_UPDATE_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.GeoIPUpdateInterval = d
		}
	}

	// Health check configuration
	if timeout := os.Getenv("GEOROUTER_HEALTH_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.HealthCheckConfig.Timeout = d
		}
	}

	if interval := os.Getenv("GEOROUTER_HEALTH_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.HealthCheckConfig.Interval = d
		}
	}

	if path := os.Getenv("GEOROUTER_HEALTH_PATH"); path != "" {
		config.HealthCheckConfig.HealthPath = path
	}

	if code := os.Getenv("GEOROUTER_HEALTH_EXPECTED_CODE"); code != "" {
		if c, err := strconv.Atoi(code); err == nil {
			config.HealthCheckConfig.ExpectedCode = c
		}
	}

	// Load balancing weights
	if loadWeight := os.Getenv("GEOROUTER_LOAD_WEIGHT"); loadWeight != "" {
		if w, err := strconv.ParseFloat(loadWeight, 64); err == nil {
			config.LoadBalancingWeights.LoadWeight = w
		}
	}

	if responseWeight := os.Getenv("GEOROUTER_RESPONSE_TIME_WEIGHT"); responseWeight != "" {
		if w, err := strconv.ParseFloat(responseWeight, 64); err == nil {
			config.LoadBalancingWeights.ResponseTimeWeight = w
		}
	}

	if errorWeight := os.Getenv("GEOROUTER_ERROR_RATE_WEIGHT"); errorWeight != "" {
		if w, err := strconv.ParseFloat(errorWeight, 64); err == nil {
			config.LoadBalancingWeights.ErrorRateWeight = w
		}
	}

	return config
}

// DefaultHealthCheckConfig returns default health check configuration
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Timeout:      5 * time.Second,
		Interval:     30 * time.Second,
		MaxRetries:   3,
		HealthPath:   "/health",
		ExpectedCode: 200,
	}
}

// GeoIPClient handles IP geolocation using MaxMind GeoLite2 database
type GeoIPClient struct {
	reader       *geoip2.Reader
	mu           sync.RWMutex
	databasePath string
}

// NewGeoIPClient creates a new GeoIP client with default database path
func NewGeoIPClient() (*GeoIPClient, error) {
	return NewGeoIPClientWithConfig(DefaultGeoRouterConfig().GeoIPDatabasePath)
}

// NewGeoIPClientWithConfig creates a new GeoIP client with specified database path
func NewGeoIPClientWithConfig(databasePath string) (*GeoIPClient, error) {
	client := &GeoIPClient{databasePath: databasePath}

	// Try to load existing database first
	if err := client.loadDatabase(); err != nil {
		logrus.Warnf("Failed to load existing GeoIP database: %v", err)
	}

	// Download/update database in background
	go func() {
		if err := client.downloadDatabase(); err != nil {
			logrus.Errorf("Failed to download GeoIP database: %v", err)
		}
	}()

	return client, nil
}

// loadDatabase attempts to load the GeoLite2 database from disk
func (c *GeoIPClient) loadDatabase() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dbPath := c.getDatabasePath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", dbPath)
	}

	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open GeoIP database: %w", err)
	}

	if c.reader != nil {
		c.reader.Close()
	}
	c.reader = reader
	logrus.Infof("Loaded GeoIP database from %s", dbPath)
	return nil
}

// downloadDatabase downloads the latest GeoLite2 Country database from MaxMind
func (c *GeoIPClient) downloadDatabase() error {
	// MaxMind GeoLite2 Country database URL (free, no API key required)
	url := "https://cdn.jsdelivr.net/npm/geolite2-country@1.0.2/GeoLite2-Country.mmdb"

	dbPath := c.getDatabasePath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Download the database
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %s", resp.Status)
	}

	// Create temporary file first
	tempPath := dbPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to write database file: %w", err)
	}
	file.Close()

	// Atomic move to final location
	if err := os.Rename(tempPath, dbPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to move database file: %w", err)
	}

	logrus.Infof("Downloaded GeoIP database to %s", dbPath)

	// Reload the database
	return c.loadDatabase()
}

// getDatabasePath returns the path where the GeoIP database should be stored
func (c *GeoIPClient) getDatabasePath() string {
	return c.databasePath
}

// LookupCountry looks up the country for an IP address
func (c *GeoIPClient) LookupCountry(ipStr string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.reader == nil {
		return "", fmt.Errorf("GeoIP database not loaded")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	record, err := c.reader.Country(ip)
	if err != nil {
		return "", fmt.Errorf("failed to lookup IP: %w", err)
	}

	if record.Country.IsoCode == "" {
		return "", fmt.Errorf("no country found for IP: %s", ipStr)
	}

	return record.Country.IsoCode, nil
}

// Close closes the GeoIP database reader
func (c *GeoIPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// NewGeoRouter creates a new geographic router with the given configuration
func NewGeoRouter(repo BackendRepository, config *GeoRouterConfig) *GeoRouter {
	if config == nil {
		config = DefaultGeoRouterConfig()
	}

	geoIPClient, err := NewGeoIPClientWithConfig(config.GeoIPDatabasePath)
	if err != nil {
		logrus.Errorf("Failed to create GeoIP client: %v", err)
		// Continue with nil client - will fallback to basic IP detection
		geoIPClient = &GeoIPClient{}
	}

	// Configure HTTP client for health checks
	healthClient := &http.Client{
		Timeout: config.HealthCheckConfig.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	return &GeoRouter{
		backends:     make(map[string]*BackendLoad),
		regions:      make(map[GeographicRegion][]*BackendLoad),
		geoIPClient:  geoIPClient,
		repo:         repo,
		config:       config,
		healthClient: healthClient,
	}
}

// NewGeoRouterWithDefaults creates a new geographic router with default configuration
func NewGeoRouterWithDefaults(repo BackendRepository) *GeoRouter {
	return NewGeoRouter(repo, DefaultGeoRouterConfig())
}

// RegisterBackend registers a backend with geographic information
func (r *GeoRouter) RegisterBackend(backend *storage.Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate backend URL
	if backend.URL == "" {
		logrus.WithField("backend_id", backend.ID).Error("Cannot register backend with empty URL")
		return
	}

	region := r.parseRegionFromLocation(backend.Region)

	load := &BackendLoad{
		BackendID:       backend.ID.String(),
		Region:          region,
		URL:             backend.URL,
		CurrentLoad:     0.0,
		ResponseTime:    time.Second, // Default 1s
		ErrorRate:       0.0,
		LastHealthCheck: time.Now(),
		Healthy:         backend.Enabled,
		Capacity:        r.config.DefaultBackendCapacity,
	}

	r.backends[backend.ID.String()] = load

	// Add to region mapping
	if r.regions[region] == nil {
		r.regions[region] = make([]*BackendLoad, 0)
	}
	r.regions[region] = append(r.regions[region], load)

	logrus.WithFields(logrus.Fields{
		"backend_id": backend.ID.String(),
		"region":     region,
		"url":        backend.URL,
	}).Info("Registered backend with geo router")
}

// UpdateBackendLoad updates the load information for a backend
func (r *GeoRouter) UpdateBackendLoad(backendID string, load float64, responseTime time.Duration, errorRate float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if backend, exists := r.backends[backendID]; exists {
		backend.CurrentLoad = load
		backend.ResponseTime = responseTime
		backend.ErrorRate = errorRate
		backend.LastHealthCheck = time.Now()
	}
}

// SetBackendHealth updates the health status of a backend
func (r *GeoRouter) SetBackendHealth(backendID string, healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if backend, exists := r.backends[backendID]; exists {
		backend.Healthy = healthy
		backend.LastHealthCheck = time.Now()
	}
}

// SelectBackend selects the best backend for a request based on geographic location and load
func (r *GeoRouter) SelectBackend(ctx context.Context, clientIP string, functionID string) (*storage.Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Determine client region from IP
	clientRegion := r.getClientRegion(clientIP)

	// Get healthy backends in the preferred region
	regionalBackends := r.getHealthyBackendsInRegion(clientRegion)
	if len(regionalBackends) == 0 {
		// Fall back to any healthy backend
		regionalBackends = r.getAllHealthyBackends()
		if len(regionalBackends) == 0 {
			return nil, fmt.Errorf("no healthy backends available")
		}
	}

	// Select backend using weighted load balancing
	selectedLoad := r.selectBackendByLoad(regionalBackends)

	// Fetch the complete backend data from database
	backendID, err := uuid.Parse(selectedLoad.BackendID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse backend ID %s: %w", selectedLoad.BackendID, err)
	}

	backend, err := r.repo.GetBackendByID(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch backend %s: %w", selectedLoad.BackendID, err)
	}
	if backend == nil {
		return nil, fmt.Errorf("backend %s not found", selectedLoad.BackendID)
	}

	return backend, nil
}

// getClientRegion determines the geographic region of a client IP
// countryToRegion maps ISO country codes to geographic regions
func countryToRegion(countryCode string) GeographicRegion {
	switch strings.ToUpper(countryCode) {
	// North America
	case "US", "CA", "MX", "PR", "VI", "GU", "AS", "MP":
		return RegionNorthAmerica

	// Europe
	case "GB", "FR", "DE", "IT", "ES", "NL", "BE", "CH", "AT", "SE", "NO", "DK",
		 "FI", "IE", "PT", "GR", "PL", "CZ", "HU", "RO", "SK", "SI", "HR", "BA",
		 "ME", "AL", "MK", "RS", "BG", "EE", "LV", "LT", "LU", "MT", "CY", "IS",
		 "LI", "AD", "MC", "SM", "VA", "JE", "GG", "IM":
		return RegionEurope

	// Asia
	case "CN", "JP", "KR", "IN", "SG", "MY", "TH", "ID", "VN", "PH", "HK", "TW",
		 "PK", "BD", "LK", "NP", "BT", "MM", "LA", "KH", "MN", "KP", "TJ", "TM",
		 "UZ", "KG", "KZ", "AF", "IR", "IQ", "SY", "JO", "LB", "IL", "PS", "SA",
		 "YE", "OM", "AE", "KW", "QA", "BH", "AZ", "GE", "AM", "BY", "UA", "MD",
		 "TR":
		return RegionAsia

	// Oceania
	case "AU", "NZ", "FJ", "PG", "SB", "VU", "NC", "PF", "CK", "NU", "TK", "TV",
		 "WF", "WS", "TO", "KI", "MH", "FM", "PW", "NR":
		return RegionOceania

	// South America
	case "BR", "AR", "CL", "CO", "PE", "VE", "EC", "BO", "PY", "UY", "GY", "SR",
		 "GF":
		return RegionSouthAmerica

	// Africa
	default:
		return RegionAfrica
	}
}

// getClientRegion determines the geographic region of a client IP
func (r *GeoRouter) getClientRegion(clientIP string) GeographicRegion {
	// Handle local/private IPs
	if clientIP == "" || clientIP == "127.0.0.1" || clientIP == "localhost" {
		return RegionNorthAmerica // Default for local development
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return RegionNorthAmerica // Default fallback
	}

	// Handle private IP ranges
	if ip.IsPrivate() || ip.IsLoopback() {
		return RegionNorthAmerica
	}

	// Try GeoIP lookup first
	if r.geoIPClient != nil {
		countryCode, err := r.geoIPClient.LookupCountry(clientIP)
		if err == nil && countryCode != "" {
			return countryToRegion(countryCode)
		}
		logrus.Debugf("GeoIP lookup failed for %s: %v, falling back to basic detection", clientIP, err)
	}

	// Fallback to basic IP-based region detection
	return r.basicRegionDetection(clientIP)
}

// basicRegionDetection provides simple IP-based region detection when GeoIP
// is unavailable or lookup fails. Uses first-octet heuristics only; for production
// configure GeoIP (GeoIPDatabasePath) so GeoIP lookup is used instead.
func (r *GeoRouter) basicRegionDetection(clientIP string) GeographicRegion {
	// Handle private networks
	if strings.HasPrefix(clientIP, "192.168.") || strings.HasPrefix(clientIP, "10.") ||
	   strings.HasPrefix(clientIP, "172.") {
		return RegionNorthAmerica // Local/private networks
	}

	// First-octet heuristic: rough geographic grouping by historic allocation
	octets := strings.Split(clientIP, ".")
	if len(octets) >= 1 {
		firstOctet := octets[0]
		switch {
		case firstOctet == "8" || firstOctet == "14" || firstOctet == "23": // Asia (China, Japan, etc.)
			return RegionAsia
		case firstOctet == "31" || firstOctet == "37" || firstOctet == "46" ||
		     firstOctet == "62" || firstOctet == "77" || firstOctet == "78" ||
		     firstOctet == "79" || firstOctet == "80" || firstOctet == "81" ||
		     firstOctet == "82" || firstOctet == "83" || firstOctet == "84" ||
		     firstOctet == "85" || firstOctet == "86" || firstOctet == "87" ||
		     firstOctet == "88" || firstOctet == "89" || firstOctet == "90" ||
		     firstOctet == "91" || firstOctet == "92" || firstOctet == "93" ||
		     firstOctet == "94" || firstOctet == "95": // Europe
			return RegionEurope
		case firstOctet == "189" || firstOctet == "190" || firstOctet == "191" ||
		     firstOctet == "200" || firstOctet == "201": // South America
			return RegionSouthAmerica
		}
	}

	return RegionNorthAmerica // Default fallback
}

// getHealthyBackendsInRegion returns healthy backends in a specific region
func (r *GeoRouter) getHealthyBackendsInRegion(region GeographicRegion) []*BackendLoad {
	backends := r.regions[region]
	if backends == nil {
		return nil
	}

	healthy := make([]*BackendLoad, 0)
	for _, backend := range backends {
		if backend.Healthy && backend.CurrentLoad < 0.9 { // Less than 90% load
			healthy = append(healthy, backend)
		}
	}

	return healthy
}

// getAllHealthyBackends returns all healthy backends across regions
func (r *GeoRouter) getAllHealthyBackends() []*BackendLoad {
	allHealthy := make([]*BackendLoad, 0)

	for _, regionBackends := range r.regions {
		for _, backend := range regionBackends {
			if backend.Healthy && backend.CurrentLoad < 0.9 {
				allHealthy = append(allHealthy, backend)
			}
		}
	}

	return allHealthy
}

// selectBackendByLoad selects a backend using weighted load balancing
func (r *GeoRouter) selectBackendByLoad(backends []*BackendLoad) *BackendLoad {
	if len(backends) == 0 {
		return nil
	}

	if len(backends) == 1 {
		return backends[0]
	}

	// Calculate weights based on load, latency, and error rate
	type backendWeight struct {
		backend *BackendLoad
		weight  float64
	}

	weights := make([]backendWeight, len(backends))

	for i, backend := range backends {
		// Weight calculation: lower load, lower latency, lower error rate = higher weight
		loadWeight := 1.0 - backend.CurrentLoad // Higher weight for lower load
		latencyWeight := math.Max(0, 1.0-(float64(backend.ResponseTime)/float64(5*time.Second))) // Higher weight for lower latency
		errorWeight := 1.0 - backend.ErrorRate // Higher weight for lower error rate

		// Use configurable weights from GeoRouter config
		totalWeight := (loadWeight * r.config.LoadBalancingWeights.LoadWeight) +
			(latencyWeight * r.config.LoadBalancingWeights.ResponseTimeWeight) +
			(errorWeight * r.config.LoadBalancingWeights.ErrorRateWeight)
		weights[i] = backendWeight{backend: backend, weight: totalWeight}
	}

	// Sort by weight (highest first)
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight > weights[j].weight
	})

	// Return the highest weighted backend
	return weights[0].backend
}

// parseRegionFromLocation converts a region string to GeographicRegion
func (r *GeoRouter) parseRegionFromLocation(location string) GeographicRegion {
	location = strings.ToLower(location)

	// Map common region names to GeographicRegion
	switch {
	case strings.Contains(location, "us-") || strings.Contains(location, "north") || strings.Contains(location, "america"):
		return RegionNorthAmerica
	case strings.Contains(location, "eu-") || strings.Contains(location, "europe") || strings.Contains(location, "frankfurt") || strings.Contains(location, "london"):
		return RegionEurope
	case strings.Contains(location, "asia") || strings.Contains(location, "tokyo") || strings.Contains(location, "singapore") || strings.Contains(location, "mumbai"):
		return RegionAsia
	case strings.Contains(location, "south") || strings.Contains(location, "america") || strings.Contains(location, "sao") || strings.Contains(location, "santiago"):
		return RegionSouthAmerica
	case strings.Contains(location, "africa") || strings.Contains(location, "johannesburg") || strings.Contains(location, "cape"):
		return RegionAfrica
	case strings.Contains(location, "oceania") || strings.Contains(location, "australia") || strings.Contains(location, "sydney"):
		return RegionOceania
	default:
		return RegionNorthAmerica // Default fallback
	}
}

// GetStats returns geographic routing statistics
func (r *GeoRouter) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})

	// Per-region statistics
	regionStats := make(map[string]interface{})
	for region, backends := range r.regions {
		total := len(backends)
		healthy := 0
		totalLoad := 0.0
		avgResponseTime := time.Duration(0)

		for _, backend := range backends {
			if backend.Healthy {
				healthy++
			}
			totalLoad += backend.CurrentLoad
			avgResponseTime += backend.ResponseTime
		}

		regionStats[string(region)] = map[string]interface{}{
			"total_backends":    total,
			"healthy_backends":  healthy,
			"average_load":      totalLoad / float64(total),
			"average_response_time_ms": avgResponseTime.Milliseconds() / int64(total),
		}
	}

	stats["regions"] = regionStats
	stats["total_backends"] = len(r.backends)

	return stats
}

// HealthCheck performs health checks on all backends
func (r *GeoRouter) HealthCheck(ctx context.Context) {
	r.mu.RLock()
	backends := make([]*BackendLoad, 0, len(r.backends))
	for _, backend := range r.backends {
		backends = append(backends, backend)
	}
	r.mu.RUnlock()

	// Perform health checks concurrently but limit concurrency
	semaphore := make(chan struct{}, 10) // Max 10 concurrent health checks
	var wg sync.WaitGroup

	for _, backend := range backends {
		wg.Add(1)
		go func(b *BackendLoad) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			r.checkBackendHealth(ctx, b)
		}(backend)
	}

	wg.Wait()
}

// checkBackendHealth performs a health check on a single backend
func (r *GeoRouter) checkBackendHealth(ctx context.Context, backend *BackendLoad) {
	// Validate URL
	if backend.URL == "" {
		r.updateHealthStatus(backend, false, time.Duration(0), fmt.Errorf("empty backend URL"))
		return
	}

	healthURL := strings.TrimSuffix(backend.URL, "/") + r.config.HealthCheckConfig.HealthPath

	// Create request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		r.updateHealthStatus(backend, false, time.Duration(0), err)
		return
	}

	// Add headers that might be expected by backends
	req.Header.Set("User-Agent", "FunctionFly-HealthChecker/1.0")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := r.healthClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		r.updateHealthStatus(backend, false, duration, err)
		return
	}
	defer resp.Body.Close()

	// Check if response is successful
	healthy := resp.StatusCode == r.config.HealthCheckConfig.ExpectedCode

	// Read and discard response body to free connection
	io.Copy(io.Discard, resp.Body)

	r.updateHealthStatus(backend, healthy, duration, nil)
}

// updateHealthStatus updates the health status and metrics for a backend
func (r *GeoRouter) updateHealthStatus(backend *BackendLoad, healthy bool, responseTime time.Duration, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	backend.LastHealthCheck = now

	// Update response time using exponential moving average
	if backend.ResponseTime == 0 {
		backend.ResponseTime = responseTime
	} else {
		// EMA with alpha = 0.3 (gives more weight to recent measurements)
		backend.ResponseTime = time.Duration(float64(backend.ResponseTime)*0.7 + float64(responseTime)*0.3)
	}

	// Update health status
	previousHealthy := backend.Healthy
	backend.Healthy = healthy

	// Update error rate
	if err != nil || !healthy {
		// Increase error rate
		backend.ErrorRate = math.Min(1.0, backend.ErrorRate+0.1)
	} else {
		// Decrease error rate
		backend.ErrorRate = math.Max(0.0, backend.ErrorRate-0.05)
	}

	// Log status changes
	if previousHealthy != healthy {
		logrus.WithFields(logrus.Fields{
			"backend_id":   backend.BackendID,
			"url":          backend.URL,
			"healthy":      healthy,
			"response_time": responseTime,
			"error_rate":    backend.ErrorRate,
		}).Info("Backend health status changed")
	}

	// Log errors for debugging
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"backend_id": backend.BackendID,
			"url":        backend.URL,
			"error":      err.Error(),
			"duration":   responseTime,
		}).Debug("Health check failed")
	}
}

// StartHealthChecks starts periodic health checking in a goroutine
func (r *GeoRouter) StartHealthChecks(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.config.HealthCheckConfig.Interval)
		defer ticker.Stop()

		// Perform initial health check
		r.HealthCheck(ctx)

		for {
			select {
			case <-ctx.Done():
				logrus.Info("Stopping health checks")
				return
			case <-ticker.C:
				r.HealthCheck(ctx)
			}
		}
	}()

	logrus.WithField("interval", r.config.HealthCheckConfig.Interval).Info("Started periodic health checks")
}

// GetHealthCheckConfig returns the current health check configuration
func (r *GeoRouter) GetHealthCheckConfig() *HealthCheckConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	config := *r.config.HealthCheckConfig
	return &config
}

// SetHealthCheckConfig updates the health check configuration
func (r *GeoRouter) SetHealthCheckConfig(config *HealthCheckConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.HealthCheckConfig = config

	// Update HTTP client timeout
	r.healthClient.Timeout = config.Timeout

	logrus.WithFields(logrus.Fields{
		"timeout":   config.Timeout,
		"interval":  config.Interval,
		"health_path": config.HealthPath,
		"expected_code": config.ExpectedCode,
	}).Info("Updated health check configuration")
}

// GetConfig returns the current GeoRouter configuration
func (r *GeoRouter) GetConfig() *GeoRouterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	config := *r.config
	// Deep copy nested structs
	config.HealthCheckConfig = &HealthCheckConfig{}
	*config.HealthCheckConfig = *r.config.HealthCheckConfig

	return &config
}

// UpdateConfig updates the GeoRouter configuration
func (r *GeoRouter) UpdateConfig(config *GeoRouterConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate config
	if config == nil {
		logrus.Error("Cannot update with nil configuration")
		return
	}

	r.config = config

	// Update HTTP client timeout
	r.healthClient.Timeout = config.HealthCheckConfig.Timeout

	logrus.WithFields(logrus.Fields{
		"geoip_db_path": config.GeoIPDatabasePath,
		"health_timeout": config.HealthCheckConfig.Timeout,
		"health_interval": config.HealthCheckConfig.Interval,
	}).Info("Updated GeoRouter configuration")
}
