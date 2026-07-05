package fly

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FlyWebhookHandler struct {
	microvmRepo *storage.MicroVMRepository
	webhookSecret string
}

func NewFlyWebhookHandler(microvmRepo *storage.MicroVMRepository) *FlyWebhookHandler {
	secret := os.Getenv("FLY_WEBHOOK_SECRET")
	if secret == "" {
		logrus.Warn("FLY_WEBHOOK_SECRET not set - webhook verification disabled")
	}
	return &FlyWebhookHandler{
		microvmRepo:   microvmRepo,
		webhookSecret: secret,
	}
}

func (h *FlyWebhookHandler) verifyWebhookSignature(r *http.Request, body []byte) bool {
	if h.webhookSecret == "" {
		return true
	}

	sig := r.Header.Get("FLY-WEBHOOK-SIGNATURE-256")
	if sig == "" {
		return false
	}

	sig = strings.TrimPrefix(sig, "sha256=")

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

type FlyMachineEvent struct {
	Event   string `json:"event"`
	ID      string `json:"id"`
	MachineID string `json:"machine_id"`
	AppName string `json:"app_name"`
	Org     string `json:"org"`
}

type FlyWebhookPayload struct {
	Event   string            `json:"event"`
	Machine FlyMachineEvent   `json:"machine"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

func (h *FlyWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read webhook body")
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if !h.verifyWebhookSignature(r, body) {
		logrus.Warn("Invalid webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload FlyWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal webhook payload")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"event":       payload.Event,
		"machine_id":  payload.Machine.MachineID,
		"app_name":    payload.Machine.AppName,
	}).Info("Received Fly webhook")

	ctx := context.Background()

	switch payload.Event {
	case "machine.stopped":
		h.handleMachineStopped(ctx, &payload.Machine)
	case "machine.destroyed":
		h.handleMachineDestroyed(ctx, &payload.Machine)
	case "machine.failed":
		h.handleMachineFailed(ctx, &payload.Machine)
	default:
		logrus.WithField("event", payload.Event).Debug("Ignoring unhandled Fly webhook event")
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func (h *FlyWebhookHandler) handleMachineStopped(ctx context.Context, machine *FlyMachineEvent) {
	if machine.MachineID == "" {
		return
	}

	exec, err := h.microvmRepo.GetExecutionByFlyMachineID(ctx, machine.MachineID)
	if err != nil {
		logrus.WithError(err).WithField("machine_id", machine.MachineID).Error("Failed to get execution by fly machine ID")
		return
	}
	if exec == nil {
		logrus.WithField("machine_id", machine.MachineID).Debug("No execution found for machine")
		return
	}

	now := time.Now()
	durationMs := int(now.Sub(exec.StartedAt).Milliseconds())

	outcome := "success"
	errorMsg := ""

	execStatus := "completed"
	if err := h.microvmRepo.UpdateExecutionStatus(ctx, exec.ID, execStatus, &outcome, &errorMsg, now, durationMs); err != nil {
		logrus.WithError(err).WithField("execution_id", exec.ID).Error("Failed to update execution status")
		return
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"machine_id":   machine.MachineID,
		"duration_ms":  durationMs,
	}).Info("Updated execution from machine stopped webhook")
}

func (h *FlyWebhookHandler) handleMachineDestroyed(ctx context.Context, machine *FlyMachineEvent) {
	if machine.MachineID == "" {
		return
	}

	exec, err := h.microvmRepo.GetExecutionByFlyMachineID(ctx, machine.MachineID)
	if err != nil {
		logrus.WithError(err).WithField("machine_id", machine.MachineID).Error("Failed to get execution by fly machine ID")
		return
	}
	if exec == nil {
		return
	}

	auditLog := &storage.MicroVMAuditLog{
		ID:           uuid.New(),
		TenantID:     exec.TenantID,
		Action:       "machine_destroyed",
		ResourceType: "machine",
		ResourceID:   uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Details:      json.RawMessage(fmt.Sprintf(`{"machine_id": %q}`, machine.MachineID)),
		CreatedAt:    time.Now(),
	}
	if err := h.microvmRepo.CreateAuditLog(ctx, auditLog); err != nil {
		logrus.WithError(err).Error("Failed to create audit log")
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"machine_id":   machine.MachineID,
	}).Debug("Machine destroyed webhook processed")
}

func (h *FlyWebhookHandler) handleMachineFailed(ctx context.Context, machine *FlyMachineEvent) {
	if machine.MachineID == "" {
		return
	}

	exec, err := h.microvmRepo.GetExecutionByFlyMachineID(ctx, machine.MachineID)
	if err != nil {
		logrus.WithError(err).WithField("machine_id", machine.MachineID).Error("Failed to get execution by fly machine ID")
		return
	}
	if exec == nil {
		return
	}

	now := time.Now()
	durationMs := int(now.Sub(exec.StartedAt).Milliseconds())

	outcome := "failed"
	errorMsg := "Machine entered failed state"

	execStatus := "failed"
	if err := h.microvmRepo.UpdateExecutionStatus(ctx, exec.ID, execStatus, &outcome, &errorMsg, now, durationMs); err != nil {
		logrus.WithError(err).WithField("execution_id", exec.ID).Error("Failed to update execution status")
		return
	}

	auditLog := &storage.MicroVMAuditLog{
		ID:           uuid.New(),
		TenantID:     exec.TenantID,
		Action:       "machine_failed",
		ResourceType: "machine",
		ResourceID:   uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Details:      json.RawMessage(fmt.Sprintf(`{"machine_id": %q, "execution_id": %q}`, machine.MachineID, exec.ID.String())),
		CreatedAt:    time.Now(),
	}
	if err := h.microvmRepo.CreateAuditLog(ctx, auditLog); err != nil {
		logrus.WithError(err).Error("Failed to create audit log")
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"machine_id":   machine.MachineID,
	}).Warn("Machine failed webhook processed")
}
