package leader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// Election manages leader election for multi-region control plane
type Election struct {
	repo       storage.LeaderRepository
	instanceID string
	region     string
	isLeader   bool
	ticker     *time.Ticker
	leaderCh   chan bool
	stopCh     chan struct{}
	mu         sync.RWMutex
	config     *Config
}

// Config holds election configuration
type Config struct {
	// ElectionInterval is how often to attempt election
	ElectionInterval time.Duration
	// LeadershipTTL is the time after which leadership expires if not renewed
	LeadershipTTL time.Duration
	// OnPromotion is called when this instance becomes leader
	OnPromotion func(ctx context.Context) error
	// OnDemotion is called when this instance loses leadership
	OnDemotion func(ctx context.Context) error
}

// DefaultConfig returns default election configuration
func DefaultConfig() *Config {
	return &Config{
		ElectionInterval: 5 * time.Second,
		LeadershipTTL:    30 * time.Second,
		OnPromotion:      nil,
		OnDemotion:       nil,
	}
}

// NewElection creates a new leader election instance
func NewElection(repo storage.LeaderRepository, instanceID, region string, config *Config) *Election {
	if config == nil {
		config = DefaultConfig()
	}

	return &Election{
		repo:       repo,
		instanceID: instanceID,
		region:     region,
		isLeader:   false,
		ticker:     time.NewTicker(config.ElectionInterval),
		leaderCh:   make(chan bool, 1),
		stopCh:     make(chan struct{}),
		config:     config,
	}
}

// Start begins the leader election process
func (e *Election) Start(ctx context.Context) {
	logrus.Infof("Starting leader election for instance %s in region %s", e.instanceID, e.region)

	// Register this region if not already registered
	if err := e.registerRegion(ctx); err != nil {
		logrus.Warnf("Failed to register region: %v", err)
	}

	go e.run(ctx)
}

// Stop stops the leader election process
func (e *Election) Stop(ctx context.Context) {
	logrus.Infof("Stopping leader election for instance %s", e.instanceID)

	close(e.stopCh)
	e.ticker.Stop()

	// Release leadership if we are the leader
	if e.IsLeader() {
		if err := e.repo.ReleaseLeadership(ctx, e.instanceID); err != nil {
			logrus.Warnf("Failed to release leadership: %v", err)
		}
	}
}

// IsLeader returns true if this instance is the leader
func (e *Election) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}

// Region returns the region this instance is running in
func (e *Election) Region() string {
	return e.region
}

// InstanceID returns the instance ID
func (e *Election) InstanceID() string {
	return e.instanceID
}

// LeaderCh returns a channel that receives true when this instance becomes leader
func (e *Election) LeaderCh() <-chan bool {
	return e.leaderCh
}

func (e *Election) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Leader election context cancelled")
			return
		case <-e.stopCh:
			logrus.Info("Leader election stopped")
			return
		case <-e.ticker.C:
			e.elect(ctx)
		}
	}
}

func (e *Election) elect(ctx context.Context) {
	// Attempt to acquire leadership
	leader, err := e.repo.GetLeader(ctx)
	if err != nil {
		logrus.Errorf("Failed to get leader: %v", err)
		return
	}

	wasLeader := e.IsLeader()

	if leader == nil || leader.InstanceID == e.instanceID {
		// No leader or we are leader, claim leadership
		err := e.repo.ClaimLeadership(ctx, e.instanceID, e.region, e.config.LeadershipTTL)
		if err != nil {
			logrus.Warnf("Failed to claim leadership: %v", err)
			if wasLeader {
				e.handleDemotion(ctx)
			}
			return
		}

		if !wasLeader {
			logrus.Infof("This instance (%s/%s) became leader", e.instanceID, e.region)
			e.setLeader(true)
			e.handlePromotion(ctx)
		}
	} else {
		if wasLeader {
			logrus.Infof("Another instance (%s/%s) is now leader", leader.InstanceID, leader.Region)
			e.setLeader(false)
			e.handleDemotion(ctx)
		}
	}
}

func (e *Election) setLeader(leader bool) {
	e.mu.Lock()
	e.isLeader = leader
	e.mu.Unlock()

	// Non-blocking send to leader channel
	select {
	case e.leaderCh <- leader:
	default:
	}
}

func (e *Election) handlePromotion(ctx context.Context) {
	if e.config.OnPromotion != nil {
		if err := e.config.OnPromotion(ctx); err != nil {
			logrus.Errorf("Error in OnPromotion handler: %v", err)
		}
	}
}

func (e *Election) handleDemotion(ctx context.Context) {
	if e.config.OnDemotion != nil {
		if err := e.config.OnDemotion(ctx); err != nil {
			logrus.Errorf("Error in OnDemotion handler: %v", err)
		}
	}
}

func (e *Election) registerRegion(ctx context.Context) error {
	return e.repo.RegisterRegion(ctx, e.region)
}

// ElectionMetrics holds election metrics
type ElectionMetrics struct {
	IsLeader       bool
	CurrentLeader  string
	CurrentRegion  string
	LastElection   time.Time
	ElectionErrors int
	TotalElections int
}

// GetMetrics returns current election metrics
func (e *Election) GetMetrics() *ElectionMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &ElectionMetrics{
		IsLeader:      e.isLeader,
		CurrentLeader: e.instanceID,
		CurrentRegion: e.region,
	}
}

// String returns a string representation of the election state
func (e *Election) String() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	role := "follower"
	if e.isLeader {
		role = "leader"
	}

	return fmt.Sprintf("Election(instance=%s, region=%s, role=%s)", e.instanceID, e.region, role)
}
