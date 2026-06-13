package vault

import (
	"context"
	"testing"
	"time"
)

func TestLeaderElector_DisabledWhenRedisNil(t *testing.T) {
	e := NewLeaderElector(nil, LeaderElectorConfig{})
	if e.IsLeader() {
		t.Fatal("expected IsLeader=false when redis is nil")
	}
	if e.Holds() {
		t.Fatal("expected Holds=false when redis is nil")
	}
}

func TestLeaderElector_LeaderKey(t *testing.T) {
	e := NewLeaderElector(nil, LeaderElectorConfig{Namespace: "vault-foo"})
	if e.LeaderKey() != "vault:leader:vault-foo" {
		t.Fatalf("got %q", e.LeaderKey())
	}
}

func TestLeaderElector_DefaultInstanceID(t *testing.T) {
	e := NewLeaderElector(nil, LeaderElectorConfig{})
	if e.InstanceID() == "" {
		t.Fatal("expected non-empty instance ID")
	}
	// Two electors should pick different IDs.
	e2 := NewLeaderElector(nil, LeaderElectorConfig{})
	if e.InstanceID() == e2.InstanceID() {
		t.Fatal("default instance IDs collided")
	}
}

func TestLeaderElector_DefensiveTTLFloor(t *testing.T) {
	// TTL shorter than 2*RenewInterval gets bumped to 3*RenewInterval.
	e := NewLeaderElector(nil, LeaderElectorConfig{
		RenewInterval: 10 * time.Second,
		TTL:           1 * time.Second, // explicitly too small
	})
	if e.ttl != 30*time.Second {
		t.Fatalf("ttl=%s, want 30s", e.ttl)
	}
}

func TestLeaderElector_LastRenewAt_InitiallyZero(t *testing.T) {
	e := NewLeaderElector(nil, LeaderElectorConfig{})
	if !e.LastRenewAt().IsZero() {
		t.Fatal("expected zero last-renew at init")
	}
}

// TestLeaderElector_TwoContendingInstancesWithRedis is an integration
// test that requires a real Redis. We exercise the SetNX behaviour
// by simulating two electors attempting to acquire the same key.
func TestLeaderElector_TwoContendingInstancesWithRedis(t *testing.T) {
	rc := newTestRedis(t)
	ctx := context.Background()
	key := "vault:leader:test-contention"
	defer rc.Del(ctx, key)

	// First elector wins.
	ok1, err := rc.SetNX(ctx, key, "instance-1", 30*time.Second).Result()
	if err != nil {
		t.Fatal(err)
	}
	if !ok1 {
		t.Fatal("first SetNX must succeed")
	}
	// Second loses.
	ok2, err := rc.SetNX(ctx, key, "instance-2", 30*time.Second).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ok2 {
		t.Fatal("second SetNX must fail while first holds the key")
	}
	// Release first.
	rc.Del(ctx, key)
	ok3, _ := rc.SetNX(ctx, key, "instance-3", 30*time.Second).Result()
	if !ok3 {
		t.Fatal("third SetNX must succeed after first releases")
	}
}
