// Package cityambassadorjob runs the hourly sync that promotes the top
// builder in each metro to ambassador. It runs *after* the city recompute
// job so it can read the latest rankings.
//
// Algorithm per cycle:
//  1. List metros with active_users >= MinActiveUsersForPublic (k=5).
//  2. For each, find the top-scoring active opted-in user (TopBuilderForMetro).
//  3. If no current ambassador → insert. If a different one is current →
//     revoke the old, insert the new. If the same user is already ambassador
//     → no-op (the partial unique index handles re-promotion).
//  4. For metros below the threshold, revoke any active ambassador (the
//     privacy contract says "no leaderboard, no ambassador").
package cityambassadorjob

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Sync is the cron entrypoint.
type Sync struct {
	repo  *cityranking.Repository
	cron  *cron.Cron
	mu    sync.Mutex
	now   func() time.Time
	entry cron.EntryID
}

// Config configures the scheduler.
type Config struct {
	Cron    string // e.g. "5 * * * *" (5 min after the city recompute)
	Enabled bool
}

// NewSync wires a Sync. now is injectable for tests.
func NewSync(repo *cityranking.Repository, now func() time.Time) *Sync {
	if now == nil {
		now = time.Now
	}
	return &Sync{repo: repo, cron: cron.New(), now: now}
}

// Start installs the cron schedule.
func (s *Sync) Start(ctx context.Context, cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !cfg.Enabled {
		logrus.Info("City ambassador sync disabled by config")
		return nil
	}
	if cfg.Cron == "" {
		cfg.Cron = "5 * * * *"
	}
	id, err := s.cron.AddFunc(cfg.Cron, func() {
		s.RunOnce(context.Background())
	})
	if err != nil {
		return err
	}
	s.entry = id
	s.cron.Start()
	logrus.WithField("cron", cfg.Cron).Info("City ambassador sync scheduler started")
	return nil
}

// Stop halts the cron.
func (s *Sync) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cron == nil {
		return nil
	}
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
	return nil
}

// RunOnce runs one sync cycle synchronously. Returns the number of metros
// promoted, revoked, and unchanged.
type RunResult struct {
	Start          time.Time
	MetrosEligible int
	Promoted       int
	Revoked        int
	Unchanged      int
	Duration       time.Duration
}

func (s *Sync) RunOnce(ctx context.Context) RunResult {
	start := s.now()
	res := RunResult{Start: start}
	defer func() {
		res.Duration = time.Since(start)
		logrus.WithFields(logrus.Fields{
			"eligible":      res.MetrosEligible,
			"promoted":      res.Promoted,
			"revoked":       res.Revoked,
			"unchanged":     res.Unchanged,
			"duration_secs": res.Duration.Seconds(),
		}).Info("City ambassador sync complete")
	}()

	eligible, err := s.repo.ListMetrosWithActiveBuilders(ctx, cityranking.MinActiveUsersForPublic)
	if err != nil {
		logrus.WithError(err).Error("Failed to list eligible metros")
		return res
	}
	res.MetrosEligible = len(eligible)

	for _, metroID := range eligible {
		top, err := s.repo.TopBuilderForMetro(ctx, metroID)
		if err != nil {
			logrus.WithError(err).WithField("metro_id", metroID).Debug("top builder lookup failed")
			continue
		}
		if top == nil {
			// No eligible user. Revoke any existing ambassador.
			if err := s.repo.RevokeAmbassador(ctx, metroID); err != nil {
				logrus.WithError(err).WithField("metro_id", metroID).Debug("revoke failed")
			}
			res.Revoked++
			continue
		}
		current, err := s.repo.GetAmbassadorForMetro(ctx, metroID)
		if err != nil {
			logrus.WithError(err).WithField("metro_id", metroID).Debug("current lookup failed")
			continue
		}
		if current != nil && current.UserID == top.UserID {
			res.Unchanged++
			continue
		}
		// Either no current ambassador, or a different one. Upsert.
		if err := s.repo.UpsertAmbassador(ctx, metroID, top.UserID, "auto"); err != nil {
			logrus.WithError(err).WithField("metro_id", metroID).Warn("upsert ambassador failed")
			continue
		}
		res.Promoted++
	}
	return res
}
