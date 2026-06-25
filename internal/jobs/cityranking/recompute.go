package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CityRankingRecompute runs the hourly city-ranking recomputation. It is
// deliberately simple: a cron-scheduled job that iterates every active metro,
// computes its signals, writes a new city_rankings row, and invalidates the
// read-side cache.
type CityRankingRecompute struct {
	repo  *cityranking.Repository
	cache *cityranking.Cache
	cron  *cron.Cron
	entry cron.EntryID
	mu    sync.Mutex
	now   func() time.Time
}

// CityRankingConfig configures the scheduler.
type CityRankingConfig struct {
	Cron    string // e.g. "0 * * * *" (hourly at minute 0)
	Enabled bool
	Workers int // parallel workers for the recompute pass
}

// NewCityRankingRecompute creates a new recompute scheduler. `now` is
// injectable for tests; pass nil to use time.Now.
func NewCityRankingRecompute(repo *cityranking.Repository, cache *cityranking.Cache, now func() time.Time) *CityRankingRecompute {
	if now == nil {
		now = time.Now
	}
	return &CityRankingRecompute{
		repo:  repo,
		cache: cache,
		cron:  cron.New(),
		now:   now,
	}
}

// Start installs the cron schedule. Safe to call once.
func (j *CityRankingRecompute) Start(ctx context.Context, cfg CityRankingConfig) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if cfg.Cron == "" {
		cfg.Cron = "0 * * * *"
	}
	id, err := j.cron.AddFunc(cfg.Cron, func() {
		j.runCycle(context.Background(), cfg.Workers)
	})
	if err != nil {
		return err
	}
	j.entry = id
	j.cron.Start()
	logrus.WithField("cron", cfg.Cron).Info("City ranking recompute scheduler started")
	return nil
}

// Stop halts the cron scheduler. Idempotent.
func (j *CityRankingRecompute) Stop(ctx context.Context) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.cron == nil {
		return nil
	}
	stopCtx := j.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
	return nil
}

// TriggerImmediate runs one recompute cycle right now in a goroutine.
func (j *CityRankingRecompute) TriggerImmediate(ctx context.Context, workers int) {
	go j.runCycle(ctx, workers)
}

// RunOnce runs one recompute cycle synchronously, blocking until it
// completes. Useful for tests and ad-hoc CLI tooling.
func (j *CityRankingRecompute) RunOnce(ctx context.Context, workers int) {
	j.runCycle(ctx, workers)
}

// runCycle is the actual work. Errors per metro are logged and skipped; the
// pass always completes.
func (j *CityRankingRecompute) runCycle(ctx context.Context, workers int) {
	start := j.now().UTC()
	periodEnd := cityranking.TruncateHour(start)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	categories := cityranking.AllCategories
	log := logrus.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
		"categories":   len(categories),
	})
	log.Info("Starting city ranking recompute cycle")

	metros, err := j.repo.ListMetros(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to list metros; aborting cycle")
		return
	}
	if len(metros) == 0 {
		log.Warn("No metros found; skipping recompute (run seed first)")
		return
	}

	if workers <= 0 {
		workers = 10
	}
	// One job per (metro, category) so signals are computed once and the
	// same Signals struct is fed into every category's Compute(). The repo
	// handles per-category upsert; ranks are assigned per category after
	// every metro has been scored.
	type job struct {
		idx      int
		category cityranking.Category
	}
	totalJobs := len(metros) * len(categories)
	jobs := make(chan job, totalJobs)
	for i := range metros {
		for _, c := range categories {
			jobs <- job{idx: i, category: c}
		}
	}
	close(jobs)

	// Group metros by index so each worker can compute signals once and
	// then score every category for that metro. The outer worker pool is
	// fixed; the inner "per-metro" loop serializes the signal query.
	var wg sync.WaitGroup
	var errCount int
	var errMu sync.Mutex
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seen := map[int]bool{}
			pending := []cityranking.Category{}
			signalsCache := map[int]cityranking.Signals{}
			for jj := range jobs {
				m := metros[jj.idx]
				if !seen[jj.idx] {
					s, err := j.repo.MetroSignals(ctx, m.ID, periodStart, periodEnd)
					if err != nil {
						errMu.Lock()
						errCount++
						errMu.Unlock()
						log.WithError(err).WithField("metro", m.Slug).Warn("signals failed")
						continue
					}
					signalsCache[jj.idx] = s
					seen[jj.idx] = true
				}
				signals := signalsCache[jj.idx]
				weights := cityranking.CategoryWeights(jj.category)
				score := cityranking.Compute(signals, m.Population, weights)
				if err := j.repo.UpsertRanking(ctx, m.ID, jj.category, score, periodStart, periodEnd); err != nil {
					errMu.Lock()
					errCount++
					errMu.Unlock()
					log.WithError(err).
						WithField("metro", m.Slug).
						WithField("category", jj.category).
						Warn("upsert failed")
				}
				_ = pending // reserved for future per-category post-processing
			}
		}()
	}
	wg.Wait()

	for _, c := range categories {
		if err := j.repo.AssignRanks(ctx, periodEnd, c); err != nil {
			log.WithError(err).WithField("category", c).Error("Failed to assign ranks")
		}
	}

	if j.cache != nil {
		if err := j.cache.InvalidateAll(ctx); err != nil {
			log.WithError(err).Warn("Failed to invalidate cache after recompute")
		}
	}

	log.WithFields(logrus.Fields{
		"metros":    len(metros),
		"jobs":      totalJobs,
		"errored":   errCount,
		"duration":  j.now().Sub(start),
	}).Info("City ranking recompute cycle complete")
}

func (j *CityRankingRecompute) scoreOne(ctx context.Context, m cityranking.MetroArea, periodStart, periodEnd time.Time) error {
	signals, err := j.repo.MetroSignals(ctx, m.ID, periodStart, periodEnd)
	if err != nil {
		return err
	}
	score := cityranking.Compute(signals, m.Population, cityranking.DefaultWeights())
	return j.repo.UpsertRanking(ctx, m.ID, cityranking.CategoryComposite, score, periodStart, periodEnd)
}
