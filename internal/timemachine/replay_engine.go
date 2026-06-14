package timemachine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	maxReplayDuration  = 4 * time.Hour
	maxItemRetries     = 2
	scanBatchSize      = 500
	progressPubMinGap  = 2 * time.Second
)

type Executor interface {
	Execute(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, timeoutMs int) (json.RawMessage, int, error)
}

type ProgressPublisher interface {
	PublishProgress(ctx context.Context, replayID uuid.UUID, data map[string]interface{})
}

type ReplayCompleteCallback func(replay *tmstorage.Replay)

type ReplayEngine struct {
	tmRepo  *tmstorage.Repository
	regRepo *registry.RegistryRepository
	repo    storage.Repository
	redis   *redis.Client
	exec    Executor

	progress   ProgressPublisher
	onComplete ReplayCompleteCallback

	mu         sync.Mutex
	activeJobs map[uuid.UUID]context.CancelFunc
}

func NewReplayEngine(
	tmRepo *tmstorage.Repository,
	regRepo *registry.RegistryRepository,
	repo storage.Repository,
	redis *redis.Client,
	exec Executor,
) *ReplayEngine {
	return &ReplayEngine{
		tmRepo:     tmRepo,
		regRepo:    regRepo,
		repo:       repo,
		redis:      redis,
		exec:       exec,
		activeJobs: make(map[uuid.UUID]context.CancelFunc),
	}
}

func (e *ReplayEngine) SetProgressPublisher(pub ProgressPublisher) {
	e.progress = pub
}

func (e *ReplayEngine) SetOnCompleteCallback(cb ReplayCompleteCallback) {
	e.onComplete = cb
}

func (e *ReplayEngine) StartReplay(replayID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), maxReplayDuration)
	e.mu.Lock()
	e.activeJobs[replayID] = cancel
	e.mu.Unlock()

	go e.processReplay(ctx, replayID)
}

func (e *ReplayEngine) CancelReplay(replayID uuid.UUID) {
	e.mu.Lock()
	cancel, ok := e.activeJobs[replayID]
	e.mu.Unlock()
	if ok {
		cancel()
	}
}

func (e *ReplayEngine) ActiveCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.activeJobs)
}

func (e *ReplayEngine) ShutdownAll() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for id, cancel := range e.activeJobs {
		logrus.WithField("replay_id", id).Info("Cancelling active replay due to shutdown")
		cancel()
	}
}

func (e *ReplayEngine) processReplay(ctx context.Context, replayID uuid.UUID) {
	defer func() {
		e.mu.Lock()
		delete(e.activeJobs, replayID)
		e.mu.Unlock()
		if r := recover(); r != nil {
			logrus.WithField("replay_id", replayID).Errorf("Replay engine panic: %v", r)
			_ = e.tmRepo.UpdateReplayStatus(replayID, "failed", 0, "")
		}
	}()

	replay, err := e.tmRepo.GetReplay(replayID)
	if err != nil || replay == nil {
		logrus.WithError(err).WithField("replay_id", replayID).Error("Failed to load replay")
		return
	}

	if err := e.tmRepo.UpdateReplayStatus(replayID, "running", 0, "scanning"); err != nil {
		logrus.WithError(err).Error("Failed to update replay status")
		return
	}

	items, scanErr := e.scanPhase(ctx, replay)
	if scanErr != nil {
		e.failReplay(replayID, fmt.Sprintf("Scan failed: %v", scanErr))
		return
	}

	if ctx.Err() != nil || e.isCancelled(replayID) {
		_ = e.tmRepo.UpdateReplayStatus(replayID, "cancelled", 0, "")
		return
	}

	replayed, changed, failed, replayErr := e.replayPhase(ctx, replay, items)
	if replayErr != nil {
		e.failReplay(replayID, fmt.Sprintf("Replay failed: %v", replayErr))
		return
	}

	if ctx.Err() != nil || e.isCancelled(replayID) {
		_ = e.tmRepo.UpdateReplayStatus(replayID, "cancelled", 0, "")
		return
	}

	diffChanged, diffErr := e.diffPhase(ctx, replay, items)
	if diffErr != nil {
		e.failReplay(replayID, fmt.Sprintf("Diff failed: %v", diffErr))
		return
	}

	if diffChanged > changed {
		changed = diffChanged
	}

	_ = e.tmRepo.UpdateReplayProgress(
		replayID, "completed", 100, "completed",
		len(items), replayed, changed, failed,
	)

	e.publishProgress(replayID, map[string]interface{}{
		"status":   "completed",
		"percent":  100,
		"found":    len(items),
		"replayed": replayed,
		"changed":  changed,
		"failed":   failed,
	})

	if e.onComplete != nil {
		completed, _ := e.tmRepo.GetReplay(replayID)
		if completed != nil {
			e.onComplete(completed)
		}
	}

	logrus.WithFields(logrus.Fields{
		"replay_id": replayID,
		"found":     len(items),
		"replayed":  replayed,
		"changed":   changed,
		"failed":    failed,
	}).Info("Time Machine replay completed")
}

func (e *ReplayEngine) isCancelled(replayID uuid.UUID) bool {
	replay, err := e.tmRepo.GetReplay(replayID)
	if err != nil || replay == nil {
		return true
	}
	return replay.Status == "cancelled"
}

func (e *ReplayEngine) failReplay(replayID uuid.UUID, msg string) {
	logrus.WithField("replay_id", replayID).Error(msg)
	_ = e.tmRepo.UpdateReplayProgress(replayID, "failed", 0, "", 0, 0, 0, 0)

	e.publishProgress(replayID, map[string]interface{}{
		"status":  "failed",
		"percent": 0,
		"error":   msg,
	})

	if e.onComplete != nil {
		completed, _ := e.tmRepo.GetReplay(replayID)
		if completed != nil {
			e.onComplete(completed)
		}
	}
}

func (e *ReplayEngine) publishProgress(replayID uuid.UUID, data map[string]interface{}) {
	if e.progress == nil {
		return
	}
	e.progress.PublishProgress(context.Background(), replayID, data)
}

func (e *ReplayEngine) scanPhase(ctx context.Context, replay *tmstorage.Replay) ([]tmstorage.ReplayItem, error) {
	e.publishProgress(replay.ID, map[string]interface{}{
		"status":  "running",
		"phase":   "scanning",
		"percent": 5,
	})

	_ = e.tmRepo.UpdateReplayProgress(
		replay.ID, "running", 5, "scanning",
		0, 0, 0, 0,
	)

	allItems := make([]tmstorage.ReplayItem, 0)
	batchLimit := scanBatchSize
	if replay.MaxExecutions > 0 && replay.MaxExecutions < scanBatchSize {
		batchLimit = replay.MaxExecutions
	}
	remaining := replay.MaxExecutions

	cursor := replay.WindowStart
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		limit := batchLimit
		if remaining > 0 && remaining < limit {
			limit = remaining
		}

		executions, err := e.regRepo.GetPublicExecutionsInWindow(ctx, 
			replay.FunctionID, cursor, replay.WindowEnd, limit,
		)
		if err != nil {
			return nil, fmt.Errorf("query executions: %w", err)
		}

		if len(executions) == 0 {
			break
		}

		batch := make([]tmstorage.ReplayItem, 0, len(executions))
		for _, exec := range executions {
			item := tmstorage.ReplayItem{
				ID:                  uuid.New(),
				ReplayID:            replay.ID,
				OriginalExecutionID: exec.ID,
				OriginalInput:       exec.InputJSON,
				OriginalOutput:      exec.OutputJSON,
				OriginalVersion:     exec.Version,
				OriginalDurationMs:  exec.DurationMs,
				OriginalTimestamp:   exec.CreatedAt,
				Status:              "pending",
			}
			batch = append(batch, item)
		}

		if len(batch) > 0 {
			if err := e.tmRepo.CreateReplayItems(batch); err != nil {
				return nil, fmt.Errorf("create replay items batch: %w", err)
			}
			allItems = append(allItems, batch...)
		}

		cursor = executions[len(executions)-1].CreatedAt.Add(time.Millisecond)

		if remaining > 0 {
			remaining -= len(executions)
			if remaining <= 0 {
				break
			}
		}

		if len(executions) < limit {
			break
		}

		if len(allItems)%1000 == 0 {
			progress := 5 + (10 * float64(len(allItems)) / float64(maxInt(replay.MaxExecutions, 1)))
			if progress > 15 {
				progress = 15
			}
			_ = e.tmRepo.UpdateReplayProgress(
				replay.ID, "running", progress, "scanning",
				len(allItems), 0, 0, 0,
			)
			e.publishProgress(replay.ID, map[string]interface{}{
				"status":  "running",
				"phase":   "scanning",
				"percent": progress,
				"found":   len(allItems),
			})
		}
	}

	_ = e.tmRepo.UpdateReplayProgress(
		replay.ID, "running", 15, "scanning",
		len(allItems), 0, 0, 0,
	)

	e.publishProgress(replay.ID, map[string]interface{}{
		"status":  "running",
		"phase":   "scanning",
		"percent": 15,
		"found":   len(allItems),
	})

	return allItems, nil
}

func (e *ReplayEngine) replayPhase(ctx context.Context, replay *tmstorage.Replay, items []tmstorage.ReplayItem) (replayed, changed, failed int, err error) {
	_ = e.tmRepo.UpdateReplayProgress(
		replay.ID, "running", 20, "replaying",
		len(items), 0, 0, 0,
	)

	e.publishProgress(replay.ID, map[string]interface{}{
		"status":  "running",
		"phase":   "replaying",
		"percent": 20,
		"found":   len(items),
	})

	fnVersion, err := e.regRepo.GetFunctionVersionByID(replay.TargetVersionID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get target version: %w", err)
	}

	timeoutMs := fnVersion.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}

	lastPublish := time.Now()
	for i := range items {
		if ctx.Err() != nil {
			return replayed, changed, failed, ctx.Err()
		}

		item := &items[i]
		_ = e.tmRepo.UpdateReplayItemStatus(item.ID, "replaying")

		progress := 20 + (60 * float64(i) / float64(len(items)))
		if i%10 == 0 || i == len(items)-1 {
			_ = e.tmRepo.UpdateReplayProgress(
				replay.ID, "running", progress, "replaying",
				len(items), replayed, changed, failed,
			)
		}

		if time.Since(lastPublish) > progressPubMinGap {
			e.publishProgress(replay.ID, map[string]interface{}{
				"status":   "running",
				"phase":    "replaying",
				"percent":  progress,
				"found":    len(items),
				"replayed": replayed,
				"failed":   failed,
			})
			lastPublish = time.Now()
		}

		newOutput, durationMs, execErr := e.exec.Execute(fnVersion, item.OriginalInput, timeoutMs)

		retries := 0
		for execErr != nil && retries < maxItemRetries {
			retries++
			backoff := time.Duration(retries) * 100 * time.Millisecond
			if backoff > 2*time.Second {
				backoff = 2 * time.Second
			}
			time.Sleep(backoff)
			logrus.WithFields(logrus.Fields{
				"item_id": item.ID,
				"attempt": retries + 1,
				"error":   execErr.Error(),
				"backoff": backoff.String(),
			}).Warn("Retrying failed replay item")
			newOutput, durationMs, execErr = e.exec.Execute(fnVersion, item.OriginalInput, timeoutMs)
		}

		if execErr != nil {
			failed++
			errMsg := execErr.Error()
			_ = e.tmRepo.UpdateReplayItemResult(
				item.ID, nil, 0, 0, "failed",
			)
			_ = e.tmRepo.UpdateReplayItemError(item.ID, errMsg, "EXECUTION_ERROR")
			continue
		}

		replayed++
		_ = e.tmRepo.UpdateReplayItemResult(
			item.ID, newOutput, durationMs, 200, "completed",
		)
	}

	return replayed, changed, failed, nil
}

func (e *ReplayEngine) diffPhase(ctx context.Context, replay *tmstorage.Replay, items []tmstorage.ReplayItem) (int, error) {
	_ = e.tmRepo.UpdateReplayProgress(
		replay.ID, "running", 85, "diffing",
		len(items), len(items), 0, 0,
	)

	e.publishProgress(replay.ID, map[string]interface{}{
		"status":  "running",
		"phase":   "diffing",
		"percent": 85,
		"found":   len(items),
	})

	changed := 0
	lastPublish := time.Now()
	for i := range items {
		if ctx.Err() != nil {
			return changed, ctx.Err()
		}

		item := &items[i]
		if item.Status == "failed" {
			diffResult := CompareOutputsForError(item.OriginalOutput, "Replay execution failed")
			summary, detail := DiffSummaryToJSON(diffResult)
			_ = e.tmRepo.UpdateReplayItemDiff(item.ID, true, string(diffResult.Type), summary, detail)
			continue
		}

		if item.NewOutput == nil || len(item.NewOutput) == 0 {
			_ = e.tmRepo.UpdateReplayItemDiff(item.ID, false, string(DiffTypeIdentical), "No new output to compare", nil)
			continue
		}

		diffResult := CompareOutputs(item.OriginalOutput, item.NewOutput)

		if diffResult.Changed {
			changed++
		}

		summary, detail := DiffSummaryToJSON(diffResult)
		_ = e.tmRepo.UpdateReplayItemDiff(
			item.ID,
			diffResult.Changed,
			string(diffResult.Type),
			summary,
			detail,
		)

		if time.Since(lastPublish) > progressPubMinGap {
			progress := 85 + (10 * float64(i) / float64(len(items)))
			e.publishProgress(replay.ID, map[string]interface{}{
				"status":  "running",
				"phase":   "diffing",
				"percent": progress,
				"found":   len(items),
				"changed": changed,
			})
			lastPublish = time.Now()
		}
	}

	_ = e.tmRepo.UpdateReplayProgress(
		replay.ID, "running", 95, "diffing",
		len(items), len(items), changed, 0,
	)

	return changed, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
