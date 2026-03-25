package degradation

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DegradationLevel defines the level of service degradation
type DegradationLevel int

const (
	// LevelNormal is full functionality
	LevelNormal DegradationLevel = iota
	// LevelDegraded is reduced functionality
	LevelDegraded
	// LevelCritical is minimal functionality
	LevelCritical
	// LevelEmergency is emergency mode
	LevelEmergency
)

// String returns the string representation of degradation level
func (l DegradationLevel) String() string {
	switch l {
	case LevelNormal:
		return "normal"
	case LevelDegraded:
		return "degraded"
	case LevelCritical:
		return "critical"
	case LevelEmergency:
		return "emergency"
	default:
		return "unknown"
	}
}

// Feature represents a feature that can be degraded
type Feature struct {
	// Name is the feature name
	Name string
	// Description describes the feature
	Description string
	// Enabled indicates if the feature is enabled
	Enabled bool
	// DegradationLevel is the minimum level for this feature
	MinLevel DegradationLevel
	// Fallback is the fallback function when degraded
	Fallback func(ctx context.Context, args ...interface{}) (interface{}, error)
}

// DegradationManager manages graceful degradation
type DegradationManager struct {
	currentLevel DegradationLevel
	features     map[string]*Feature
	mu           sync.RWMutex
	history      []DegradationEvent
	maxHistory   int
}

// DegradationEvent represents a degradation event
type DegradationEvent struct {
	Timestamp time.Time
	FromLevel DegradationLevel
	ToLevel   DegradationLevel
	Reason    string
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager() *DegradationManager {
	return &DegradationManager{
		currentLevel: LevelNormal,
		features:     make(map[string]*Feature),
		history:      make([]DegradationEvent, 0),
		maxHistory:   100,
	}
}

// RegisterFeature registers a feature with degradation support
func (m *DegradationManager) RegisterFeature(feature *Feature) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.features[feature.Name] = feature
	logrus.WithFields(logrus.Fields{
		"feature":         feature.Name,
		"min_level":       feature.MinLevel.String(),
		"has_fallback":    feature.Fallback != nil,
	}).Info("Registered feature for degradation management")
}

// SetLevel sets the current degradation level
func (m *DegradationManager) SetLevel(level DegradationLevel, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldLevel := m.currentLevel
	m.currentLevel = level

	// Record event
	event := DegradationEvent{
		Timestamp: time.Now(),
		FromLevel: oldLevel,
		ToLevel:   level,
		Reason:    reason,
	}

	m.history = append(m.history, event)
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}

	logrus.WithFields(logrus.Fields{
		"from_level": oldLevel.String(),
		"to_level":   level.String(),
		"reason":     reason,
	}).Warn("Degradation level changed")
}

// GetLevel returns the current degradation level
func (m *DegradationManager) GetLevel() DegradationLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentLevel
}

// IsFeatureEnabled checks if a feature is enabled at the current level
func (m *DegradationManager) IsFeatureEnabled(featureName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[featureName]
	if !exists {
		return false
	}

	if !feature.Enabled {
		return false
	}

	return m.currentLevel <= feature.MinLevel
}

// ExecuteFeature executes a feature with degradation support
func (m *DegradationManager) ExecuteFeature(ctx context.Context, featureName string, fn func(ctx context.Context) (interface{}, error), args ...interface{}) (interface{}, error) {
	m.mu.RLock()
	feature, exists := m.features[featureName]
	currentLevel := m.currentLevel
	m.mu.RUnlock()

	if !exists {
		logrus.WithField("feature", featureName).Warn("Feature not registered")
		return nil, nil
	}

	if !feature.Enabled {
		logrus.WithField("feature", featureName).Debug("Feature disabled")
		return nil, nil
	}

	// Check if feature is available at current level
	if currentLevel <= feature.MinLevel {
		// Feature is available, execute normally
		return fn(ctx)
	}

	// Feature is degraded, use fallback if available
	if feature.Fallback != nil {
		logrus.WithFields(logrus.Fields{
			"feature":      featureName,
			"current_level": currentLevel.String(),
			"min_level":    feature.MinLevel.String(),
		}).Info("Using fallback for degraded feature")

		return feature.Fallback(ctx, args...)
	}

	// No fallback available
	logrus.WithFields(logrus.Fields{
		"feature":      featureName,
		"current_level": currentLevel.String(),
		"min_level":    feature.MinLevel.String(),
	}).Warn("Feature degraded with no fallback")

	return nil, nil
}

// GetFeatureStatus returns the status of all features
func (m *DegradationManager) GetFeatureStatus() map[string]FeatureStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]FeatureStatus)
	for name, feature := range m.features {
		status[name] = FeatureStatus{
			Name:           name,
			Enabled:        feature.Enabled,
			Available:      m.currentLevel <= feature.MinLevel,
			MinLevel:       feature.MinLevel.String(),
			CurrentLevel:   m.currentLevel.String(),
			HasFallback:    feature.Fallback != nil,
		}
	}

	return status
}

// FeatureStatus represents the status of a feature
type FeatureStatus struct {
	Name         string
	Enabled      bool
	Available    bool
	MinLevel     string
	CurrentLevel string
	HasFallback  bool
}

// GetHistory returns degradation history
func (m *DegradationManager) GetHistory() []DegradationEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]DegradationEvent, len(m.history))
	copy(history, m.history)
	return history
}

// AutoDegradation automatically adjusts degradation level based on metrics
type AutoDegradation struct {
	manager           *DegradationManager
	errorThreshold    float64
	latencyThreshold  time.Duration
	checkInterval     time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

// NewAutoDegradation creates a new auto degradation manager
func NewAutoDegradation(manager *DegradationManager, errorThreshold float64, latencyThreshold time.Duration) *AutoDegradation {
	return &AutoDegradation{
		manager:          manager,
		errorThreshold:   errorThreshold,
		latencyThreshold: latencyThreshold,
		checkInterval:    30 * time.Second,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the auto degradation monitor
func (a *AutoDegradation) Start() {
	a.wg.Add(1)
	go a.monitorLoop()
}

// Stop stops the auto degradation monitor
func (a *AutoDegradation) Stop() {
	close(a.stopChan)
	a.wg.Wait()
}

// monitorLoop monitors metrics and adjusts degradation level
func (a *AutoDegradation) monitorLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.checkAndAdjust()
		}
	}
}

// checkAndAdjust checks metrics and adjusts degradation level
func (a *AutoDegradation) checkAndAdjust() {
	// This would integrate with your metrics system
	// For now, it's a placeholder that demonstrates the pattern

	currentLevel := a.manager.GetLevel()

	// Example logic:
	// - If error rate > threshold, increase degradation level
	// - If error rate < threshold/2, decrease degradation level

	// In a real implementation, you would:
	// 1. Query your metrics system for error rate and latency
	// 2. Compare against thresholds
	// 3. Adjust degradation level accordingly

	logrus.WithFields(logrus.Fields{
		"current_level":     currentLevel.String(),
		"error_threshold":   a.errorThreshold,
		"latency_threshold": a.latencyThreshold,
	}).Debug("Auto degradation check completed")
}

// Predefined features for FunctionFly
var (
	// RoutingFeature represents the routing feature
	RoutingFeature = &Feature{
		Name:           "routing",
		Description:    "Request routing to backends",
		Enabled:        true,
		MinLevel:       LevelNormal,
		Fallback:       nil, // No fallback, routing is critical
	}

	// HealthCheckFeature represents health checking
	HealthCheckFeature = &Feature{
		Name:           "health_check",
		Description:    "Backend health monitoring",
		Enabled:        true,
		MinLevel:       LevelNormal,
		Fallback:       nil, // No fallback, health checks are critical
	}

	// CachingFeature represents caching
	CachingFeature = &Feature{
		Name:           "caching",
		Description:    "Response caching",
		Enabled:        true,
		MinLevel:       LevelDegraded,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: skip cache, go directly to backend
			logrus.Info("Cache degraded, bypassing cache")
			return nil, nil
		},
	}

	// AnalyticsFeature represents analytics collection
	AnalyticsFeature = &Feature{
		Name:           "analytics",
		Description:    "Analytics and metrics collection",
		Enabled:        true,
		MinLevel:       LevelCritical,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: skip analytics, continue processing
			logrus.Info("Analytics degraded, skipping collection")
			return nil, nil
		},
	}

	// RateLimitingFeature represents rate limiting
	RateLimitingFeature = &Feature{
		Name:           "rate_limiting",
		Description:    "Request rate limiting",
		Enabled:        true,
		MinLevel:       LevelDegraded,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: allow all requests
			logrus.Info("Rate limiting degraded, allowing all requests")
			return true, nil
		},
	}
)

// Global degradation manager instance
var GlobalDegradationManager = NewDegradationManager()

// InitializeDegradation initializes degradation management
func InitializeDegradation() {
	GlobalDegradationManager.RegisterFeature(RoutingFeature)
	GlobalDegradationManager.RegisterFeature(HealthCheckFeature)
	GlobalDegradationManager.RegisterFeature(CachingFeature)
	GlobalDegradationManager.RegisterFeature(AnalyticsFeature)
	GlobalDegradationManager.RegisterFeature(RateLimitingFeature)
}

