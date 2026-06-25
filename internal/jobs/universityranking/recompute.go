// Package universityrankingjob runs the hourly recompute that materializes
// the per-university ranking rows. It mirrors the city recompute so the
// two leaderboards stay on the same time grid (period_end = top of the
// hour) and the same k-anonymity threshold.
package universityrankingjob

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/universityranking"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Recompute is the cron entrypoint.
type Recompute struct {
	repo  *universityranking.Repository
	cache *universityranking.Cache
	cron  *cron.Cron
	mu    sync.Mutex
	entry cron.EntryID
	now   func() time.Time
}

// Config configures the scheduler.
type Config struct {
	Cron    string // e.g. "0 * * * *"
	Enabled bool
	Workers int
}

// NewRecompute creates a new recompute scheduler. `now` is injectable for
// tests; pass nil to use time.Now.
func NewRecompute(repo *universityranking.Repository, cache *universityranking.Cache, now func() time.Time) *Recompute {
	if now == nil {
		now = time.Now
	}
	return &Recompute{
		repo:  repo,
		cache: cache,
		cron:  cron.New(),
		now:   now,
	}
}

// Start installs the cron schedule. Safe to call once.
func (r *Recompute) Start(ctx context.Context, cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !cfg.Enabled {
		logrus.Info("University ranking recompute disabled by config")
		return nil
	}
	if cfg.Cron == "" {
		cfg.Cron = "0 * * * *"
	}
	workers := cfg.Workers
	if workers <= 0 {
		workers = 8
	}
	id, err := r.cron.AddFunc(cfg.Cron, func() {
		r.RunOnce(context.Background(), workers)
	})
	if err != nil {
		return err
	}
	r.entry = id
	r.cron.Start()
	logrus.WithField("cron", cfg.Cron).Info("University ranking recompute scheduler started")
	return nil
}

// Stop halts the cron scheduler. Idempotent.
func (r *Recompute) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cron == nil {
		return nil
	}
	stopCtx := r.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
	return nil
}

// TriggerImmediate runs one recompute cycle right now in a goroutine.
func (r *Recompute) TriggerImmediate(ctx context.Context, workers int) {
	go r.RunOnce(ctx, workers)
}

// RunOnce runs one recompute cycle synchronously. Used by tests and the
// cron callback.
func (r *Recompute) RunOnce(ctx context.Context, workers int) {
	start := r.now().UTC()
	periodEnd := universityranking.TruncateHour(start)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	categories := universityranking.AllCategories()
	log := logrus.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
		"categories":   len(categories),
	})
	log.Info("Starting university ranking recompute cycle")

	unis, err := r.repo.ListAll(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to list universities; aborting cycle")
		return
	}
	if len(unis) == 0 {
		log.Warn("No universities found; skipping recompute (run seed first)")
		return
	}
	log.WithField("universities", len(unis)).Info("Listed universities; starting workers")

	if workers <= 0 {
		workers = 8
	}

	// Fan out per (university, category) across `workers` goroutines. The
	// channel is closed after all jobs are sent so the workers' `for j :=
	// range work` loop exits naturally — without the close, the workers
	// block forever on a never-closed channel.
	work := make(chan recomputeJob, len(unis)*len(categories))
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]scoredRow, 0, len(unis)*len(categories))
		errored int
	)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range work {
				rk, err := j.score(ctx, r.repo, periodStart, periodEnd)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"university_id": j.university.ID,
						"category":      j.category,
					}).Debug("score failed")
					mu.Lock()
					errored++
					mu.Unlock()
					continue
				}
				mu.Lock()
				results = append(results, scoredRow{rk: rk, cat: j.category})
				mu.Unlock()
			}
		}()
	}
	for _, u := range unis {
		for _, cat := range categories {
			work <- recomputeJob{university: u, category: cat}
		}
	}
	close(work)
	wg.Wait()
	log.WithField("results", len(results)).Info("Workers finished")

	prevPositions, _ := r.fetchPrevPositions(ctx, periodEnd)
	byCategory := map[universityranking.Category][]universityranking.Ranking{}
	for _, res := range results {
		if res.rk.ActiveUsers < universityranking.MinActiveUsersForPublic {
			continue
		}
		if prev, ok := prevPositions[res.rk.UniversityID]; ok {
			prevCopy := prev
			res.rk.PrevRank = &prevCopy
			res.rk.RankDelta = prevCopy - res.rk.RankPosition
		}
		byCategory[res.cat] = append(byCategory[res.cat], res.rk)
	}
	written := 0
	for cat, rs := range byCategory {
		universityranking.SortRankings(rs)
		for i := range rs {
			rs[i].RankPosition = i + 1
			if err := r.repo.UpsertRanking(ctx, rs[i], cat); err != nil {
				logrus.WithError(err).WithField("category", cat).Warn("upsert ranking failed")
				errored++
				continue
			}
			written++
		}
		_ = r.cache.SetLeaderboard(ctx, "", cat, rs)
		for _, country := range []string{"US", "GB", "IN", "DE", "JP", "BR", "AU", "CN", "CA"} {
			filtered := filterByCountry(rs, country)
			if len(filtered) > 0 {
				_ = r.cache.SetLeaderboard(ctx, country, cat, filtered)
			}
		}
	}

	log.WithFields(logrus.Fields{
		"period_end":    periodEnd,
		"universities":  len(unis),
		"categories":    len(categories),
		"ranks":         written,
		"errored":       errored,
		"duration_secs": time.Since(start).Seconds(),
	}).Info("University ranking recompute cycle complete")
}

type recomputeJob struct {
	university universityranking.University
	category   universityranking.Category
}

type scoredRow struct {
	rk  universityranking.Ranking
	cat universityranking.Category
}

func (j recomputeJob) score(ctx context.Context, repo *universityranking.Repository, periodStart, periodEnd time.Time) (universityranking.Ranking, error) {
	signals, err := repo.SignalsFor(ctx, j.university.ID, periodStart, periodEnd)
	if err != nil {
		return universityranking.Ranking{}, err
	}
	wA, wD, wE, wF, wN := universityranking.CategoryWeights(j.category)
	score := universityranking.Compute(signals, j.university.StudentCount, wA, wD, wE, wF, wN)
	return universityranking.Ranking{
		UniversityID:    j.university.ID,
		UniversitySlug:  j.university.Slug,
		UniversityName:  j.university.Name,
		ShortName:       j.university.ShortName,
		CountryCode:     j.university.CountryCode,
		StateCode:       j.university.StateCode,
		CitySlug:        "",
		ActiveUsers:     score.ActiveUsers,
		Deployments:     score.Deployments,
		Executions30d:   score.Executions30d,
		FounderEarnings: signals.FounderEarnings,
		NewUsers30d:     score.NewUsers30d,
		ScoreRaw:        score.Raw,
		ScorePerCapita:  score.PerCapita,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
	}, nil
}

func (r *Recompute) fetchPrevPositions(ctx context.Context, currentEnd time.Time) (map[int64]int, error) {
	prevEnd := currentEnd.Add(-1 * time.Hour)
	rows, err := r.repo.Pool().Query(ctx, `
		SELECT university_id, rank_position
		FROM university_rankings
		WHERE period_end = $1
	`, prevEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int{}
	for rows.Next() {
		var id int64
		var pos int
		if err := rows.Scan(&id, &pos); err != nil {
			return nil, err
		}
		out[id] = pos
	}
	return out, rows.Err()
}

func filterByCountry(rs []universityranking.Ranking, country string) []universityranking.Ranking {
	out := make([]universityranking.Ranking, 0, len(rs))
	for _, r := range rs {
		if r.CountryCode == country {
			out = append(out, r)
		}
	}
	return out
}
