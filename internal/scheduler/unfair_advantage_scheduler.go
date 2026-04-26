package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type UnfairAdvantageScheduler struct {
	cron     *cron.Cron
	engine   *swarm.UnfairAdvantageEngine
	mu       sync.RWMutex
	scheduleID cron.EntryID
	config   *UnfairAdvantageScheduleConfig
	isRunning bool
	lastRun  time.Time
	nextRun  time.Time
}

type UnfairAdvantageScheduleConfig struct {
	Enabled  bool   `json:"enabled"`
	Cron     string `json:"cron"`
	Timezone string `json:"timezone"`
}

func NewUnfairAdvantageScheduler(engine *swarm.UnfairAdvantageEngine) *UnfairAdvantageScheduler {
	return &UnfairAdvantageScheduler{
		cron:   cron.New(),
		engine: engine,
		config: &UnfairAdvantageScheduleConfig{
			Enabled:  false,
			Cron:     "0 2 * * *",
			Timezone: "UTC",
		},
	}
}

func (s *UnfairAdvantageScheduler) Start(ctx context.Context, config UnfairAdvantageScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		logrus.Info("Unfair advantage scheduler already running")
		return nil
	}

	s.config = &config

	if _, err := cron.ParseStandard(config.Cron); err != nil {
		return err
	}

	if config.Enabled {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.runRDLabCycle(ctx)
		})
		if err != nil {
			return err
		}
		s.scheduleID = entryID

		schedule, _ := cron.ParseStandard(config.Cron)
		s.nextRun = schedule.Next(time.Now())
	}

	s.cron.Start()
	s.isRunning = true

	logrus.Infof("Unfair advantage scheduler started (enabled=%v, cron=%s)", config.Enabled, config.Cron)
	return nil
}

func (s *UnfairAdvantageScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping unfair advantage scheduler")
	<-s.cron.Stop().Done()
	s.isRunning = false
	return nil
}

func (s *UnfairAdvantageScheduler) UpdateConfig(ctx context.Context, config UnfairAdvantageScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := cron.ParseStandard(config.Cron); err != nil {
		return err
	}

	if s.scheduleID > 0 {
		s.cron.Remove(s.scheduleID)
	}

	s.config = &config

	if config.Enabled {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.runRDLabCycle(ctx)
		})
		if err != nil {
			return err
		}
		s.scheduleID = entryID

		schedule, _ := cron.ParseStandard(config.Cron)
		s.nextRun = schedule.Next(time.Now())
	} else {
		s.scheduleID = 0
	}

	return nil
}

func (s *UnfairAdvantageScheduler) GetStatus() UnfairAdvantageScheduleStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return UnfairAdvantageScheduleStatus{
		IsEnabled: s.config.Enabled,
		Cron:      s.config.Cron,
		Timezone:  s.config.Timezone,
		IsRunning: s.isRunning,
		LastRun:   s.lastRun,
		NextRun:   s.nextRun,
	}
}

func (s *UnfairAdvantageScheduler) runRDLabCycle(ctx context.Context) {
	logrus.Info("Starting scheduled RD Lab cycle")

	run, err := s.engine.RunInternalRDLab(ctx)
	if err != nil {
		logrus.WithError(err).Error("Scheduled RD Lab cycle failed")
		return
	}

	s.mu.Lock()
	s.lastRun = time.Now()
	if s.config.Enabled {
		schedule, _ := cron.ParseStandard(s.config.Cron)
		s.nextRun = schedule.Next(time.Now())
	}
	s.mu.Unlock()

	logrus.Infof("Scheduled RD Lab cycle completed: ideas_scouted=%d, ideas_funded=%d, value_tracked=$%.2f",
		run.IdeasScouted, run.IdeasFunded, run.TotalValueTracked)
}

type UnfairAdvantageScheduleStatus struct {
	IsEnabled bool      `json:"is_enabled"`
	Cron      string    `json:"cron"`
	Timezone  string    `json:"timezone"`
	IsRunning bool      `json:"is_running"`
	LastRun   time.Time `json:"last_run"`
	NextRun   time.Time `json:"next_run"`
}