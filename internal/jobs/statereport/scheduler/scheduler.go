// Package statereportjob runs the monthly "State of AI Builders" report.
// The cron expression is `0 9 1 * *` (09:00 UTC on the 1st of each month)
// by default. The report file is written to the path in
// `STATE_REPORT_OUTPUT_DIR` (default `web/site/src/content/reports`)
// so the marketing site can pick it up at build time.
package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/jobs/statereport"
	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/functionfly/functionfly/internal/storage/universityranking"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Job is the cron entrypoint.
type Job struct {
	cityRepo *cityranking.Repository
	univRepo *universityranking.Repository
	outputDir string
	cron     *cron.Cron
	mu       sync.Mutex
	entry    cron.EntryID
	now      func() time.Time
}

// New wires a Job. outputDir is the directory the Markdown file is
// written to. If empty, defaults to "web/site/src/content/reports".
// now is injectable for tests.
func New(city *cityranking.Repository, univ *universityranking.Repository, outputDir string, now func() time.Time) *Job {
	if outputDir == "" {
		outputDir = "web/site/src/content/reports"
	}
	if now == nil {
		now = time.Now
	}
	return &Job{
		cityRepo:  city,
		univRepo:  univ,
		outputDir: outputDir,
		cron:      cron.New(),
		now:       now,
	}
}

// Start installs the cron schedule. Safe to call once.
func (j *Job) Start(ctx context.Context, expr string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if expr == "" {
		expr = "0 9 1 * *" // 09:00 UTC on the 1st of each month
	}
	id, err := j.cron.AddFunc(expr, func() {
		// The report covers the previous month, not the current one.
		// So we always subtract 1 day so that on the 1st of the month
		// we still pick up the last day of the previous month as the
		// "reference" — that gives the prior calendar month.
		ref := j.now().UTC().AddDate(0, 0, -1)
		if _, err := j.RunOnce(ctx, ref); err != nil {
			logrus.WithError(err).Error("state report failed")
		}
	})
	if err != nil {
		return fmt.Errorf("parse cron expr %q: %w", expr, err)
	}
	j.entry = id
	j.cron.Start()
	logrus.WithField("expr", expr).Info("State report cron started")
	return nil
}

// Stop halts the cron.
func (j *Job) Stop(ctx context.Context) error {
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

// RunOnce generates one report and writes it to disk. Returns the
// absolute path of the written file.
func (j *Job) RunOnce(ctx context.Context, reference time.Time) (string, error) {
	b := statereport.New(j.cityRepo, j.univRepo, j.now)
	rep, err := b.BuildForMonth(ctx, reference)
	if err != nil {
		return "", fmt.Errorf("build: %w", err)
	}
	if err := os.MkdirAll(j.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", j.outputDir, err)
	}
	path := filepath.Join(j.outputDir, rep.Slug+".md")
	body := rep.Render()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	logrus.WithFields(logrus.Fields{
		"slug":       rep.Slug,
		"path":       path,
		"metros":     rep.HeadlineStats.MetrosRanked,
		"universities": rep.HeadlineStats.UniversitiesRanked,
		"users":      rep.HeadlineStats.TotalActiveUsers,
	}).Info("State report written")
	return path, nil
}
