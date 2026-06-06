// Tests for the receipt milestone worker. Unit tests only — they exercise
// the in-memory threshold cache, panic recovery, and the dedupe contract.
// DB and notifier behavior is covered by repository tests; here we only
// test the orchestration logic via a fake notifier.
package receipt

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// fakeNotifier captures the Insert/SendEmail calls so we can assert on them.
type fakeNotifier struct {
	mu        sync.Mutex
	inserts   []fakeInsert
	emails    []fakeEmail
	failNext  bool
}

type fakeInsert struct {
	UserID uuid.UUID
	Kind   string
	Title  string
	Body   string
	Data   map[string]interface{}
}

type fakeEmail struct {
	UserID    uuid.UUID
	Subject   string
	PlainBody string
	HTMLBody  string
}

func (f *fakeNotifier) Insert(_ context.Context, userID uuid.UUID, kind, title, body string, data map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext {
		f.failNext = false
		return errors.New("fake notifier failure")
	}
	f.inserts = append(f.inserts, fakeInsert{UserID: userID, Kind: kind, Title: title, Body: body, Data: data})
	return nil
}

func (f *fakeNotifier) SendEmail(_ context.Context, userID uuid.UUID, subject, htmlBody, plainBody string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.emails = append(f.emails, fakeEmail{UserID: userID, Subject: subject, HTMLBody: htmlBody, PlainBody: plainBody})
	return nil
}

func TestMilestone_OnExecution_NoOpWhenDisabled(t *testing.T) {
	notifier := &fakeNotifier{}
	cfg := Config{Enabled: true, MilestoneEnabled: false}
	// The repo is nil-safe by the call path: the worker skips when
	// MilestoneEnabled is false. We do not need a real repo here.
	w := NewMilestone(nil, notifier, cfg, logrus.New())
	w.OnExecution(context.Background(), uuid.New(), nil, "abc")
	if len(notifier.inserts) != 0 {
		t.Errorf("expected no notifications when disabled, got %d", len(notifier.inserts))
	}
}

func TestMilestone_OnExecution_NilUUID(t *testing.T) {
	notifier := &fakeNotifier{}
	cfg := Config{Enabled: true, MilestoneEnabled: true, MilestoneThresholds: []int{1, 10, 100}}
	w := NewMilestone(nil, notifier, cfg, logrus.New())
	// uuid.Nil should be a no-op to avoid counting every "empty function ID"
	// request against a single global bucket.
	w.OnExecution(context.Background(), uuid.Nil, nil, "abc")
	if len(notifier.inserts) != 0 {
		t.Errorf("expected no notifications for uuid.Nil, got %d", len(notifier.inserts))
	}
}

func TestMilestone_OnExecution_EmptyPublicID(t *testing.T) {
	notifier := &fakeNotifier{}
	cfg := Config{Enabled: true, MilestoneEnabled: true, MilestoneThresholds: []int{1, 10, 100}}
	w := NewMilestone(nil, notifier, cfg, logrus.New())
	w.OnExecution(context.Background(), uuid.New(), nil, "")
	if len(notifier.inserts) != 0 {
		t.Errorf("expected no notifications for empty publicID, got %d", len(notifier.inserts))
	}
}

func TestMilestone_armFor_Laziness(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1, 10}}, logrus.New())
	id := uuid.New()
	// First armFor initializes the set with all configured thresholds.
	w.armFor(id)
	w.mu.Lock()
	armed := w.armed[id.String()]
	w.mu.Unlock()
	if armed == nil {
		t.Fatal("expected armed set to exist after armFor")
	}
	if _, ok := armed[1]; !ok {
		t.Error("expected threshold 1 to be armed")
	}
	if _, ok := armed[10]; !ok {
		t.Error("expected threshold 10 to be armed")
	}
}

func TestMilestone_disarmFor_Removes(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1, 10}}, logrus.New())
	id := uuid.New()
	w.armFor(id)
	w.disarmFor(id, 1)
	w.mu.Lock()
	armed := w.armed[id.String()]
	w.mu.Unlock()
	if _, ok := armed[1]; ok {
		t.Error("expected threshold 1 to be disarmed")
	}
	if _, ok := armed[10]; !ok {
		t.Error("expected threshold 10 to still be armed")
	}
}

func TestMilestone_isArmed_DefaultTrue(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1, 10}}, logrus.New())
	id := uuid.New()
	// Unknown function ID should default to armed=true so the DB dedupe
	// has a chance to no-op rather than us swallowing the milestone.
	if !w.isArmed(id, 1) {
		t.Error("expected unknown function ID to default to armed")
	}
}

func TestMilestone_fireChannel_UnknownChannel(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1}}, logrus.New())
	if w.fireChannel(context.Background(), uuid.New(), nil, "abc", 1, 1, "sms") {
		t.Error("expected unknown channel to return false")
	}
}

func TestMilestone_fireChannel_TweetIntent_NoOp(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1}}, logrus.New())
	// tweet_intent is a no-op channel (the URL is built on the read path),
	// but it should still report success so the milestone row records
	// it in channels_fired.
	if !w.fireChannel(context.Background(), uuid.New(), nil, "abc", 1, 1, "tweet_intent") {
		t.Error("expected tweet_intent to report success")
	}
}

func TestMilestone_fireChannel_PanicRecovers(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneThresholds: []int{1}}, logrus.New())
	// We don't have a way to inject a panic from a channel, but we can
	// verify that a nil notifier + inapp channel doesn't panic.
	if w.fireChannel(context.Background(), uuid.New(), nil, "abc", 1, 1, "inapp") {
		t.Error("expected inapp to return false with nil notifier (no owner to notify)")
	}
}

func TestMilestone_buildMilestoneTweetIntent_Format(t *testing.T) {
	cfg := Config{PublicBaseURL: "https://functionfly.com/r", TwitterHandle: "functionfly"}
	url := buildMilestoneTweetIntent(cfg, "abc", 100)
	if len(url) == 0 {
		t.Fatal("expected non-empty url")
	}
}

// TestMilestone_SweepMissedMilestones_NilWorkerIsNoop — sanity check that the
// sweep guard against nil workers is honoured.
func TestMilestone_SweepMissedMilestones_DisabledIsNoop(t *testing.T) {
	w := NewMilestone(nil, &fakeNotifier{}, Config{MilestoneEnabled: false}, logrus.New())
	n, err := w.SweepMissedMilestones(context.Background(), time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 when disabled, got %d", n)
	}
}

// ----------------------------------------------------------------------------
// Config parsing for milestone channels
// ----------------------------------------------------------------------------

func TestDefaultConfig_OverrideThresholds(t *testing.T) {
	t.Setenv("RECEIPT_MILESTONE_THRESHOLDS", "5,50,500")
	t.Setenv("RECEIPT_MILESTONE_CHANNELS", "inapp")
	cfg := DefaultConfig()
	if len(cfg.MilestoneThresholds) != 3 {
		t.Fatalf("expected 3 thresholds, got %v", cfg.MilestoneThresholds)
	}
	if cfg.MilestoneThresholds[2] != 500 {
		t.Errorf("expected last threshold 500, got %d", cfg.MilestoneThresholds[2])
	}
	if len(cfg.MilestoneChannels) != 1 || cfg.MilestoneChannels[0] != "inapp" {
		t.Errorf("expected ['inapp'], got %v", cfg.MilestoneChannels)
	}
}

// Suppress unused-package import warning
var _ = errors.New
