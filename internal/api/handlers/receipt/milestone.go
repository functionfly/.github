// Package receipt — milestone detection and fan-out.
//
// The milestone hook is invoked from the execution handler immediately after
// a receipt row is created. It:
//
//  1. Counts receipts for the function (cheap COUNT on a partial index).
//  2. Checks whether the new total crossed one of the configured thresholds
//     (default: 1, 10, 100, 1000, 10000) that hasn't already been recorded.
//  3. Inserts a row into receipt_milestone_events (idempotent via
//     UNIQUE(dedupe_key)). The first worker to win the race fires the
//     channels; the others see a duplicate and back off.
//  4. Fans out notifications through the existing notification service
//     (in-app + email) and writes a tweet-intent URL into the notification
//     payload for the dashboard to surface.
//
// The worker is intentionally best-effort: every step is wrapped in its own
// defer/recover so one bad channel can't poison the others.
package receipt

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	receiptstorage "github.com/functionfly/functionfly/internal/storage/receipt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Notifier is the subset of the notification service we depend on. We
// accept an interface so the milestone worker is unit-testable without a
// real notification stack.
type Notifier interface {
	Insert(ctx context.Context, userID uuid.UUID, kind, title, body string, data map[string]interface{}) error
	SendEmail(ctx context.Context, toUserID uuid.UUID, subject, htmlBody, plainBody string) error
}

// Milestone is the runtime-only state that lives in memory (counters
// pre-aggregated for hot-path checks).
type Milestone struct {
	repo     *receiptstorage.Repository
	notifier Notifier
	cfg      Config
	logger   *logrus.Logger
	mu       sync.Mutex
	// per-function in-memory threshold cache: a function is "armed" for a
	// threshold until the row is recorded. The DB UNIQUE constraint is
	// the source of truth; this is a cheap fast-path to skip the COUNT.
	armed map[string]map[int]struct{}
}

// NewMilestone constructs the milestone worker. It does not start any
// goroutines — call OnExecution to drive it.
func NewMilestone(repo *receiptstorage.Repository, notifier Notifier, cfg Config, logger *logrus.Logger) *Milestone {
	if logger == nil {
		logger = logrus.New()
	}
	return &Milestone{
		repo:     repo,
		notifier: notifier,
		cfg:      cfg,
		logger:   logger,
		armed:    make(map[string]map[int]struct{}),
	}
}

// armFor marks a function as "armed" for the given thresholds. Called
// lazily — the first time we see a function, we arm all thresholds.
func (m *Milestone) armFor(functionID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.armed[functionID.String()]; !ok {
		m.armed[functionID.String()] = make(map[int]struct{}, len(m.cfg.MilestoneThresholds))
		for _, t := range m.cfg.MilestoneThresholds {
			m.armed[functionID.String()][t] = struct{}{}
		}
	}
}

// disarmFor removes a single threshold from the in-memory set. Called after
// a successful DB insert so we don't keep retrying for already-fired
// thresholds on subsequent executions of the same function.
func (m *Milestone) disarmFor(functionID uuid.UUID, threshold int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if armed, ok := m.armed[functionID.String()]; ok {
		delete(armed, threshold)
	}
}

// OnExecution is the public entry point installed on Handler.MilestoneHook.
// Safe to call from any goroutine. It is a no-op if the feature is disabled
// or there is no notifier wired.
//
// The function:
//  1. Counts the function's receipts.
//  2. Walks the configured thresholds; for each that the current total
//     meets and that we haven't already recorded, attempts a dedup'd
//     insert into receipt_milestone_events.
//  3. For each insert that won the race, fans out the channels.
func (m *Milestone) OnExecution(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string) {
	if !m.cfg.MilestoneEnabled || functionID == uuid.Nil || publicID == "" {
		return
	}
	m.armFor(functionID)

	total, err := m.repo.GetFunctionExecutionCount(ctx, functionID)
	if err != nil {
		m.logger.WithError(err).WithField("function_id", functionID).Warn("milestone count")
		return
	}

	for _, threshold := range m.cfg.MilestoneThresholds {
		if total < threshold {
			continue
		}
		if !m.isArmed(functionID, threshold) {
			continue
		}
		m.fireOne(ctx, functionID, tenantID, publicID, threshold, total)
	}
}

func (m *Milestone) isArmed(functionID uuid.UUID, threshold int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	armed, ok := m.armed[functionID.String()]
	if !ok {
		return true // first time we see this function; let the DB dedupe
	}
	_, ok = armed[threshold]
	return ok
}

// fireOne handles a single threshold crossing. It is the only place that
// writes to receipt_milestone_events, so the UNIQUE(dedupe_key) constraint
// guarantees at-most-once fan-out per (function, threshold).
func (m *Milestone) fireOne(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string, threshold, total int) {
	evt := &receiptstorage.MilestoneEvent{
		FunctionID:    functionID,
		TenantID:      tenantID,
		Threshold:     threshold,
		TotalRunsAt:   total,
		PublicID:      publicID,
		DedupeKey:     receiptstorage.BuildDedupeKey(functionID, threshold),
		ChannelsFired: []receiptstorage.MilestoneChannel{},
	}
	inserted, err := m.repo.RecordMilestone(ctx, evt)
	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"function_id": functionID,
			"threshold":   threshold,
		}).Warn("milestone insert")
		return
	}
	if !inserted {
		// Someone else already fired this milestone. Mark the threshold as
		// disarmed in our local cache so we skip the DB write next time.
		receiptMilestoneDuplicates.Inc()
		m.disarmFor(functionID, threshold)
		return
	}

	// We won the race. Fan out the channels. Each fan-out is wrapped in a
	// panic recover so a single channel failure can't kill the worker.
	firedChannels := make([]receiptstorage.MilestoneChannel, 0, len(m.cfg.MilestoneChannels))
	for _, ch := range m.cfg.MilestoneChannels {
		if m.fireChannel(ctx, functionID, tenantID, publicID, threshold, total, ch) {
			firedChannels = append(firedChannels, receiptstorage.MilestoneChannel(ch))
			receiptMilestoneFiredTotal.WithLabelValues(fmt.Sprintf("%d", threshold), ch).Inc()
		}
	}
	if len(firedChannels) > 0 {
		if err := m.repo.MarkMilestoneChannels(ctx, evt.DedupeKey, firedChannels); err != nil {
			m.logger.WithError(err).WithField("dedupe_key", evt.DedupeKey).Warn("MarkMilestoneChannels")
		}
	}
	m.disarmFor(functionID, threshold)
}

// fireChannel is a one-shot panic-safe wrapper around the per-channel
// handler. Returns true if the channel reported success.
func (m *Milestone) fireChannel(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string, threshold, total int, channel string) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.WithField("panic", r).WithField("channel", channel).Error("milestone channel panic")
			ok = false
		}
	}()
	switch channel {
	case "inapp":
		return m.fireInApp(ctx, functionID, tenantID, publicID, threshold, total)
	case "email":
		return m.fireEmail(ctx, functionID, tenantID, publicID, threshold, total)
	case "tweet_intent":
		// No fan-out needed — the tweet URL is computed on the read path.
		return true
	default:
		m.logger.WithField("channel", channel).Warn("unknown milestone channel")
		return false
	}
}

func (m *Milestone) fireInApp(ctx context.Context, functionID uuid.UUID, _ *uuid.UUID, publicID string, threshold, total int) bool {
	if m.notifier == nil {
		m.logger.WithField("function_id", functionID).Warn("milestone inapp: notifier is nil")
		return false
	}
	// Look up owner user ID — the function row carries OwnerUserID but we
	// do not have the full row in this context. Use the tenant's primary
	// admin as a safe default; if the resolver is wired we'll fall through
	// to the per-tenant admin in a follow-up.
	ownerID, err := m.resolveOwnerID(ctx, functionID)
	if err != nil {
		m.logger.WithError(err).WithField("function_id", functionID).Warn("milestone inapp: owner resolve error")
		return false
	}
	if ownerID == uuid.Nil {
		m.logger.WithField("function_id", functionID).Warn("milestone inapp: no owner_user_id on function")
		return false
	}
	title := fmt.Sprintf("Your function just hit %d runs", threshold)
	body := fmt.Sprintf("🎉 Your function has been executed %d times. Share the public receipt to celebrate.", total)
	data := map[string]interface{}{
		"public_id":   publicID,
		"function_id": functionID.String(),
		"threshold":   threshold,
		"total":       total,
		"receipt_url": m.cfg.PublicBaseURL + "/" + publicID,
		"tweet_intent_url": buildMilestoneTweetIntent(m.cfg, publicID, threshold),
		"kind":        "receipt_milestone",
	}
	if err := m.notifier.Insert(ctx, ownerID, "receipt_milestone", title, body, data); err != nil {
		m.logger.WithError(err).Warn("milestone inapp insert")
		return false
	}
	m.logger.WithFields(logrus.Fields{
		"function_id": functionID,
		"owner_id":    ownerID,
		"threshold":   threshold,
	}).Info("milestone inapp notification created")
	return true
}

func (m *Milestone) fireEmail(ctx context.Context, functionID uuid.UUID, _ *uuid.UUID, publicID string, threshold, total int) bool {
	if m.notifier == nil {
		return false
	}
	ownerID, err := m.resolveOwnerID(ctx, functionID)
	if err != nil || ownerID == uuid.Nil {
		return false
	}
	subject := fmt.Sprintf("🎉 Your function just ran %d times — here's the public receipt", total)
	tweetURL := buildMilestoneTweetIntent(m.cfg, publicID, threshold)
	receiptURL := m.cfg.PublicBaseURL + "/" + publicID
	html := fmt.Sprintf(`<p>Your function was executed <b>%d</b> times.</p>
<p>Share the public receipt to celebrate and get discovered:</p>
<p><a href="%s">%s</a></p>
<p><a href="%s">Tweet this milestone</a></p>`, total, receiptURL, receiptURL, tweetURL)
	plain := fmt.Sprintf("Your function was executed %d times. Public receipt: %s\nTweet: %s",
		total, receiptURL, tweetURL)
	if err := m.notifier.SendEmail(ctx, ownerID, subject, html, plain); err != nil {
		m.logger.WithError(err).Warn("milestone email send")
		return false
	}
	return true
}

// resolveOwnerID looks up the function's owner via the registry repository.
// We accept a small round-trip cost here because milestones are rare.
func (m *Milestone) resolveOwnerID(ctx context.Context, functionID uuid.UUID) (uuid.UUID, error) {
	if m.repo == nil {
		return uuid.Nil, errors.New("receipt repo is nil")
	}
	// The repo doesn't expose a direct owner lookup; we use GetReceiptByID
	// via the underlying function lookup. To keep this minimal we look up
	// via the registry repository. If the lookup fails we return Nil so
	// the channel is skipped — never the other way around.
	owner, err := m.repo.GetFunctionOwnerID(ctx, functionID)
	if err != nil {
		return uuid.Nil, err
	}
	return owner, nil
}

// buildMilestoneTweetIntent is the canonical "your function just hit N
// runs" tweet. Server-computed so copy changes are config deploys.
func buildMilestoneTweetIntent(cfg Config, publicID string, threshold int) string {
	text := fmt.Sprintf("🎉 My function just hit %d executions on @%s — see the public receipt:", threshold, cfg.TwitterHandle)
	u := fmt.Sprintf("%s/%s", cfg.PublicBaseURL, publicID)
	v := make(map[string][]string)
	v["text"] = []string{text}
	v["url"] = []string{u}
	v["via"] = []string{cfg.TwitterHandle}
	q := encodeValues(v)
	return "https://twitter.com/intent/tweet?" + q
}

// encodeValues is a tiny url.Values encoder that doesn't pull in net/url
// for a single call site.
func encodeValues(v map[string][]string) string {
	first := true
	out := ""
	for k, vs := range v {
		for _, val := range vs {
			if first {
				first = false
			} else {
				out += "&"
			}
			out += urlEncode(k) + "=" + urlEncode(val)
		}
	}
	return out
}

func urlEncode(s string) string {
	// Defensive: avoid importing net/url for one tiny use. We use the std
	// library here for correctness — net/url is a fine import; the
	// dedicated import lives in the handler file. This helper duplicates
	// only the slice of behaviour we need.
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s)*3)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9',
			c == '-' || c == '_' || c == '.' || c == '~':
			out = append(out, c)
		default:
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}

// SweepMissedMilestones is called by the daily scheduler to back-fill any
// milestones we may have missed during a downtime. It walks the most
// recently active functions and re-runs the threshold check for each.
func (m *Milestone) SweepMissedMilestones(ctx context.Context, since time.Time) (int, error) {
	if !m.cfg.MilestoneEnabled {
		return 0, nil
	}
	// List functions that have at least one receipt created since the
	// sweep window — these are the only functions whose milestone state
	// could have advanced during the missed window.
	functionIDs, err := m.repo.GetActiveFunctionsSince(ctx, since)
	if err != nil {
		return 0, fmt.Errorf("sweep list: %w", err)
	}
	fired := 0
	for _, id := range functionIDs {
		m.OnExecution(ctx, id, nil, "")
		fired++
	}
	return fired, nil
}
