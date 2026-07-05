package machines

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

type Watchdog struct {
	client       *fly.FlyMachinesClient
	microvmRepo *storage.MicroVMRepository
	interval     time.Duration
	startedAt   time.Time
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewWatchdog(client *fly.FlyMachinesClient, microvmRepo *storage.MicroVMRepository, interval time.Duration) *Watchdog {
	return &Watchdog{
		client:       client,
		microvmRepo: microvmRepo,
		interval:     interval,
		stopCh:      make(chan struct{}),
	}
}

func (w *Watchdog) Start(ctx context.Context) {
	w.startedAt = time.Now()
	w.wg.Add(1)

	go w.run(ctx)

	logrus.WithFields(logrus.Fields{
		"interval":  w.interval,
		"app_name": w.client.GetAppName(),
	}).Info("Fly Machines watchdog started")
}

func (w *Watchdog) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	logrus.Info("Fly Machines watchdog stopped")
}

func (w *Watchdog) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *Watchdog) poll(ctx context.Context) {
	machines, err := w.client.ListMachines(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Watchdog: failed to list machines")
		return
	}

	for _, machine := range machines {
		if machine.State != "started" && machine.State != "starting" {
			continue
		}

		if !w.shouldWatchMachine(&machine) {
			continue
		}

		exec, err := w.microvmRepo.GetExecutionByFlyMachineID(ctx, machine.ID)
		if err != nil {
			logrus.WithError(err).WithField("machine_id", machine.ID).Warn("Watchdog: failed to get execution")
			continue
		}
		if exec == nil {
			continue
		}

		elapsed := time.Since(exec.StartedAt)
		timeout := time.Duration(exec.MemoryMB) * time.Millisecond * 100

		if exec.Status == "starting" && elapsed > 30*time.Second {
			logrus.WithFields(logrus.Fields{
				"machine_id":    machine.ID,
				"execution_id":  exec.ID,
				"elapsed":       elapsed,
			}).Warn("Watchdog: Machine stuck in starting state, stopping")

			w.stopAndDeleteMachine(ctx, exec, machine.ID, "hung_in_starting")
		}

		if exec.Status == "running" && elapsed > timeout && timeout > 0 {
			logrus.WithFields(logrus.Fields{
				"machine_id":    machine.ID,
				"execution_id":  exec.ID,
				"elapsed":       elapsed,
				"timeout":       timeout,
			}).Warn("Watchdog: Machine exceeded timeout, stopping")

			w.stopAndDeleteMachine(ctx, exec, machine.ID, "timeout")
		}
	}
}

func (w *Watchdog) shouldWatchMachine(machine *fly.Machine) bool {
	if machine.Config.Metadata == nil {
		return false
	}

	if _, ok := machine.Config.Metadata["ff_execution_id"]; ok {
		return true
	}

	return false
}

func (w *Watchdog) stopAndDeleteMachine(ctx context.Context, exec *storage.MicroVMExecution, machineID, reason string) {
	if err := w.client.StopMachine(ctx, machineID); err != nil {
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			logrus.WithError(err).WithField("machine_id", machineID).Error("Watchdog: failed to stop machine")
		}
	}

	time.Sleep(500 * time.Millisecond)

	if err := w.client.DeleteMachine(ctx, machineID); err != nil {
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
			logrus.WithError(err).WithField("machine_id", machineID).Error("Watchdog: failed to delete machine")
		}
	}

	outcome := reason
	errorMsg := "Machine cleaned up by watchdog"
	now := time.Now()
	durationMs := int(now.Sub(exec.StartedAt).Milliseconds())

	if err := w.microvmRepo.UpdateExecutionStatus(ctx, exec.ID, "failed", &outcome, &errorMsg, now, durationMs); err != nil {
		logrus.WithError(err).WithField("execution_id", exec.ID).Error("Watchdog: failed to update execution status")
	}

	auditLog := &storage.MicroVMAuditLog{
		TenantID:     exec.TenantID,
		Action:       "watchdog_cleanup",
		ResourceType: "machine",
		Details:      []byte(`{"machine_id": "` + machineID + `", "reason": "` + reason + `"}`),
		CreatedAt:    time.Now(),
	}
	if err := w.microvmRepo.CreateAuditLog(ctx, auditLog); err != nil {
		logrus.WithError(err).Error("Watchdog: failed to create audit log")
	}

	logrus.WithFields(logrus.Fields{
		"machine_id":   machineID,
		"execution_id": exec.ID,
		"reason":       reason,
	}).Info("Watchdog: cleaned up hung machine")
}
