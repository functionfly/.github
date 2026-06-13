package vault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// LeaderElector implements a Redis SETNX-based leader election so
// background workers (secret expiry sweep, dynamic-lease cleanup,
// SIEM dispatch, etc.) only run on one node at a time. The
// candidate key is renewed periodically; if a candidate fails to
// renew, another node will claim the lock within the next attempt
// window.
//
// Key:   vault:leader:{namespace}
// Value: instance_id (UUID)
// TTL:   30s; renew interval 10s.
//
// This is the standard pattern documented at
//
//	https://redis.io/docs/manual/patterns/distributed-locks/
//
// (single-instance simplified lease).
type LeaderElector struct {
	redis      *redis.Client
	namespace  string
	instanceID string
	logger     *logrus.Logger

	// ttl is the SET key TTL. Lower bound: 2*renewInterval.
	ttl time.Duration
	// renewInterval is how often we extend the lease while leader.
	renewInterval time.Duration
	// acquireInterval is how often we attempt to acquire when not
	// the leader.
	acquireInterval time.Duration

	// isLeader is updated by the run loop and read by IsLeader.
	isLeader atomic.Bool
	// lastRenew tracks the time of the last successful renew;
	// callers can inspect it for diagnostics.
	lastRenew atomic.Int64
	// acquired tracks whether the current process holds the lock
	// (vs. another instance does).
	holds atomic.Bool
}

// LeaderElectorConfig configures the elector.
type LeaderElectorConfig struct {
	// Namespace is the lock scope (e.g. "vault-sweep"). Two
	// electors in the same namespace contend; different namespaces
	// do not.
	Namespace string
	// InstanceID is the candidate ID stored in Redis. Default: a
	// random UUID generated at construction time.
	InstanceID string
	// TTL is the SET key TTL. Default 30s.
	TTL time.Duration
	// RenewInterval is the cadence for keep-alive attempts. Default 10s.
	RenewInterval time.Duration
	// AcquireInterval is the cadence for re-attempts when not leader.
	// Default 5s.
	AcquireInterval time.Duration
	// Logger is the structured logger; nil means logrus.New().
	Logger *logrus.Logger
}

// NewLeaderElector constructs an elector.
func NewLeaderElector(redisClient *redis.Client, cfg LeaderElectorConfig) *LeaderElector {
	if cfg.Namespace == "" {
		cfg.Namespace = "vault-sweep"
	}
	if cfg.InstanceID == "" {
		if host, err := os.Hostname(); err == nil && host != "" {
			cfg.InstanceID = fmt.Sprintf("%s-%s", host, uuid.New().String())
		} else {
			cfg.InstanceID = uuid.New().String()
		}
	}
	if cfg.TTL == 0 {
		cfg.TTL = 30 * time.Second
	}
	if cfg.RenewInterval == 0 {
		cfg.RenewInterval = 10 * time.Second
	}
	if cfg.AcquireInterval == 0 {
		cfg.AcquireInterval = 5 * time.Second
	}
	if cfg.RenewInterval*2 > cfg.TTL {
		// Defensive: TTL must be at least 2x renew.
		cfg.TTL = cfg.RenewInterval * 3
	}
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return &LeaderElector{
		redis:           redisClient,
		namespace:       cfg.Namespace,
		instanceID:      cfg.InstanceID,
		logger:          logger,
		ttl:             cfg.TTL,
		renewInterval:   cfg.RenewInterval,
		acquireInterval: cfg.AcquireInterval,
	}
}

// IsLeader reports whether the current process is the leader.
// Safe to call from any goroutine.
func (l *LeaderElector) IsLeader() bool {
	if l == nil {
		return false
	}
	return l.isLeader.Load()
}

// InstanceID returns the candidate's stable ID.
func (l *LeaderElector) InstanceID() string { return l.instanceID }

// LeaderKey returns the Redis key used for election.
func (l *LeaderElector) LeaderKey() string { return "vault:leader:" + l.namespace }

// Run blocks until ctx is cancelled, participating in the
// election. Callers should `go elector.Run(ctx)` to start the loop
// and `elector.IsLeader()` to gate exclusive work.
func (l *LeaderElector) Run(ctx context.Context) {
	if l == nil || l.redis == nil {
		l.logger.Warn("LeaderElector has no Redis client — IsLeader will always be false")
		<-ctx.Done()
		return
	}
	l.logger.WithFields(logrus.Fields{
		"namespace":   l.namespace,
		"instance_id": l.instanceID,
		"ttl":         l.ttl.String(),
	}).Info("Starting leader election loop")

	renew := time.NewTicker(l.renewInterval)
	defer renew.Stop()
	acquire := time.NewTicker(l.acquireInterval)
	defer acquire.Stop()

	for {
		select {
		case <-ctx.Done():
			l.releaseIfHeld()
			l.isLeader.Store(false)
			return
		case <-renew.C:
			if l.holds.Load() {
				l.renewLease(ctx)
			}
		case <-acquire.C:
			if !l.holds.Load() {
				l.tryAcquire(ctx)
			}
		}
	}
}

// tryAcquire attempts to claim the leader key with SETNX.
func (l *LeaderElector) tryAcquire(ctx context.Context) {
	ok, err := l.redis.SetNX(ctx, l.LeaderKey(), l.instanceID, l.ttl).Result()
	if err != nil {
		l.logger.WithError(err).Warn("Leader SETNX failed")
		return
	}
	if ok {
		l.holds.Store(true)
		l.isLeader.Store(true)
		l.lastRenew.Store(time.Now().Unix())
		l.logger.WithField("instance_id", l.instanceID).Info("Acquired leader lock")
	}
}

// renewLease extends the lock only if we still own it. We do this
// with a Lua script for atomicity: GET-and-EXPIRE-if-mine.
func (l *LeaderElector) renewLease(ctx context.Context) {
	const renewScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
  return 0
end`
	ttlMs := l.ttl.Milliseconds()
	res, err := l.redis.Eval(ctx, renewScript, []string{l.LeaderKey()}, l.instanceID, ttlMs).Result()
	if err != nil {
		l.logger.WithError(err).Warn("Leader renew failed")
		// If the script errored we may have lost the lock; check.
		l.checkStillLeader(ctx)
		return
	}
	n, _ := res.(int64)
	if n > 0 {
		l.lastRenew.Store(time.Now().Unix())
		return
	}
	// Renew returned 0: someone else owns the lock now.
	l.holds.Store(false)
	l.isLeader.Store(false)
	l.logger.Warn("Lost leader lock during renew")
}

// checkStillLeader does a final GET to confirm whether we still hold
// the lock after an error. Defensive: rarely needed.
func (l *LeaderElector) checkStillLeader(ctx context.Context) {
	val, err := l.redis.Get(ctx, l.LeaderKey()).Result()
	if err != nil {
		l.holds.Store(false)
		l.isLeader.Store(false)
		return
	}
	if val == l.instanceID {
		l.holds.Store(true)
		l.isLeader.Store(true)
	} else {
		l.holds.Store(false)
		l.isLeader.Store(false)
	}
}

// releaseIfHeld deletes the key if and only if we still own it. Best
// effort — on shutdown we don't block on Redis.
func (l *LeaderElector) releaseIfHeld() {
	if !l.holds.Load() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end`
	_, err := l.redis.Eval(ctx, releaseScript, []string{l.LeaderKey()}, l.instanceID).Result()
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		l.logger.WithError(err).Debug("Leader release failed")
	}
}

// LastRenewAt returns the timestamp of the last successful renew.
func (l *LeaderElector) LastRenewAt() time.Time {
	ts := l.lastRenew.Load()
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// Holds reports whether this instance is the one that acquired
// the lock. Distinct from IsLeader: an instance can briefly see
// IsLeader=false during a renew blip but still Holds=true.
func (l *LeaderElector) Holds() bool {
	if l == nil {
		return false
	}
	return l.holds.Load()
}
