package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// ScheduleConfig represents a function schedule configuration
type ScheduleConfig struct {
	ID          cron.EntryID `json:"-"` // cron entry id for removal
	Cron        string       `json:"cron"`
	Timezone    string       `json:"timezone"`
	Enabled     bool         `json:"enabled"`
	LastRun     time.Time    `json:"last_run"`
	NextRun     time.Time    `json:"next_run"`
	RunOnDeploy bool         `json:"run_on_deploy"`
}

// FunctionScheduler manages scheduled function executions
type FunctionScheduler struct {
	cron          *cron.Cron
	storage       storage.Repository
	functionCache map[uuid.UUID]*ScheduleConfig
	mu            sync.RWMutex
	executors     map[uuid.UUID]FunctionExecutor
	defaultExec   FunctionExecutor // used when no function-specific executor is registered
}

// FunctionExecutor defines the interface for executing scheduled functions
type FunctionExecutor interface {
	ExecuteFunction(ctx context.Context, functionID uuid.UUID, input []byte) ([]byte, error)
}

// NewFunctionScheduler creates a new function scheduler
func NewFunctionScheduler(repo storage.Repository) *FunctionScheduler {
	return &FunctionScheduler{
		cron:          cron.New(),
		storage:       repo,
		functionCache: make(map[uuid.UUID]*ScheduleConfig),
		executors:     make(map[uuid.UUID]FunctionExecutor),
	}
}

// Start starts the scheduler
func (s *FunctionScheduler) Start(ctx context.Context) error {
	logrus.Info("Starting function scheduler")
	s.cron.Start()
	return nil
}

// Stop stops the scheduler
func (s *FunctionScheduler) Stop(ctx context.Context) error {
	logrus.Info("Stopping function scheduler")
	<-s.cron.Stop().Done()
	return nil
}

// AddSchedule adds a schedule for a function
func (s *FunctionScheduler) AddSchedule(ctx context.Context, functionID uuid.UUID, config *ScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate cron expression
	if _, err := cron.ParseStandard(config.Cron); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Remove existing schedule if any
	if existing, ok := s.functionCache[functionID]; ok {
		s.cron.Remove(existing.ID)
	}

	// Add new schedule
	entryID, err := s.cron.AddFunc(config.Cron, func() {
		s.executeScheduledFunction(ctx, functionID)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	config.ID = entryID
	s.functionCache[functionID] = config

	// Calculate next run
	schedule, err := cron.ParseStandard(config.Cron)
	if err == nil {
		config.NextRun = schedule.Next(time.Now())
	}

	logrus.Infof("Added schedule for function %s: %s", functionID, config.Cron)
	return nil
}

// RemoveSchedule removes a schedule for a function
func (s *FunctionScheduler) RemoveSchedule(ctx context.Context, functionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config, ok := s.functionCache[functionID]; ok {
		s.cron.Remove(config.ID)
		delete(s.functionCache, functionID)
		logrus.Infof("Removed schedule for function %s", functionID)
	}

	return nil
}

// GetSchedule gets the schedule for a function
func (s *FunctionScheduler) GetSchedule(functionID uuid.UUID) (*ScheduleConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.functionCache[functionID]
	return config, ok
}

// ListSchedules lists all schedules
func (s *FunctionScheduler) ListSchedules() []*ScheduleConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules := make([]*ScheduleConfig, 0, len(s.functionCache))
	for _, config := range s.functionCache {
		schedules = append(schedules, config)
	}

	return schedules
}

// executeScheduledFunction executes a scheduled function
func (s *FunctionScheduler) executeScheduledFunction(ctx context.Context, functionID uuid.UUID) {
	logrus.Infof("Executing scheduled function %s", functionID)

	s.mu.RLock()
	config, ok := s.functionCache[functionID]
	s.mu.RUnlock()

	if !ok {
		logrus.Warnf("No schedule found for function %s", functionID)
		return
	}

	// Verify function exists in storage
	if _, err := s.storage.GetFunctionByID(ctx, functionID); err != nil {
		logrus.WithError(err).Errorf("Failed to get function %s", functionID)
		return
	}

	// Execute the function
	input := []byte(fmt.Sprintf(`{"trigger": "scheduled", "timestamp": "%s", "function_id": "%s"}`, time.Now().UTC().Format(time.RFC3339), functionID))

	// Get the actual executor from the map
	s.mu.RLock()
	fnExecutor, hasExecutor := s.executors[functionID]
	s.mu.RUnlock()

	if hasExecutor {
		_, err := fnExecutor.ExecuteFunction(ctx, functionID, input)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to execute scheduled function %s", functionID)
			// Update last run time even on error to avoid repeated failures
			s.updateLastRun(functionID, config)
			return
		}
	} else {
		logrus.Warnf("No executor registered for scheduled function %s", functionID)
		return
	}

	// Update last run time
	s.updateLastRun(functionID, config)

	logrus.Infof("Successfully executed scheduled function %s", functionID)
}

// updateLastRun updates the last run time for a schedule
func (s *FunctionScheduler) updateLastRun(functionID uuid.UUID, config *ScheduleConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cfg, ok := s.functionCache[functionID]; ok {
		cfg.LastRun = time.Now()
		schedule, err := cron.ParseStandard(cfg.Cron)
		if err == nil {
			cfg.NextRun = schedule.Next(time.Now())
		}
	}
}

// RegisterExecutor registers an executor for a function
func (s *FunctionScheduler) RegisterExecutor(functionID uuid.UUID, executor FunctionExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executors[functionID] = executor
}

// UnregisterExecutor unregisters an executor for a function
func (s *FunctionScheduler) UnregisterExecutor(functionID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.executors, functionID)
}

// ValidateCronExpression validates a cron expression
func ValidateCronExpression(expression string) error {
	_, err := cron.ParseStandard(expression)
	return err
}

// GetNextRunTime returns the next run time for a cron expression
func GetNextRunTime(expression string, timezone string) (time.Time, error) {
	schedule, err := cron.ParseStandard(expression)
	if err != nil {
		return time.Time{}, err
	}

	loc := time.UTC
	if timezone != "" {
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return time.Time{}, err
		}
	}

	return schedule.Next(time.Now().In(loc)), nil
}

// GetHumanReadableSchedule returns a human-readable description of the schedule
func GetHumanReadableSchedule(expression string) (string, error) {
	schedule, err := cron.ParseStandard(expression)
	if err != nil {
		return "", err
	}

	next := schedule.Next(time.Now())
	return fmt.Sprintf("Next run: %s", next.Format("Mon, Jan 2, 2006 at 3:04 PM")), nil
}

// Common schedule presets
var (
	// EveryMinute runs every minute
	EveryMinute = "* * * * *"
	// Every5Minutes runs every 5 minutes
	Every5Minutes = "*/5 * * * *"
	// Every15Minutes runs every 15 minutes
	Every15Minutes = "*/15 * * * *"
	// Every30Minutes runs every 30 minutes
	Every30Minutes = "*/30 * * * *"
	// EveryHour runs every hour
	EveryHour = "0 * * * *"
	// EveryDayAtMidnight runs every day at midnight
	EveryDayAtMidnight = "0 0 * * *"
	// EveryDayAtNoon runs every day at noon
	EveryDayAtNoon = "0 12 * * *"
	// EveryWeekday runs on weekdays at midnight
	EveryWeekday = "0 0 * * 1-5"
	// EveryWeek runs on Sunday at midnight
	EveryWeek = "0 0 * * 0"
	// EveryMonth runs on the first day of every month at midnight
	EveryMonth = "0 0 1 * *"
)

// SchedulePreset represents a schedule preset
type SchedulePreset struct {
	Name        string `json:"name"`
	Cron        string `json:"cron"`
	Description string `json:"description"`
}

// GetSchedulePresets returns common schedule presets
func GetSchedulePresets() []SchedulePreset {
	return []SchedulePreset{
		{Name: "Every minute", Cron: EveryMinute, Description: "Runs every minute"},
		{Name: "Every 5 minutes", Cron: Every5Minutes, Description: "Runs every 5 minutes"},
		{Name: "Every 15 minutes", Cron: Every15Minutes, Description: "Runs every 15 minutes"},
		{Name: "Every 30 minutes", Cron: Every30Minutes, Description: "Runs every 30 minutes"},
		{Name: "Every hour", Cron: EveryHour, Description: "Runs at the start of every hour"},
		{Name: "Every day at midnight", Cron: EveryDayAtMidnight, Description: "Runs once a day at midnight"},
		{Name: "Every day at noon", Cron: EveryDayAtNoon, Description: "Runs once a day at noon"},
		{Name: "Weekdays at midnight", Cron: EveryWeekday, Description: "Runs Monday through Friday at midnight"},
		{Name: "Every week", Cron: EveryWeek, Description: "Runs once a week on Sunday at midnight"},
		{Name: "Every month", Cron: EveryMonth, Description: "Runs on the first day of every month"},
	}
}
