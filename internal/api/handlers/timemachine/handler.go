package timemachine

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	tmengine "github.com/functionfly/functionfly/internal/timemachine"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	tmRepo    *tmstorage.Repository
	repo      storage.Repository
	regRepo   *registry.RegistryRepository
	redis     *redis.Client
	engine    *tmengine.ReplayEngine
	recEngine *tmengine.ReconciliationEngine
	auditGen  *tmengine.AuditGenerator
	notifier  *notification.Service
	authSvc   *auth.AuthService
}

func NewHandler(
	tmRepo *tmstorage.Repository,
	repo storage.Repository,
	regRepo *registry.RegistryRepository,
	redis *redis.Client,
	platformKey ed25519.PrivateKey,
	notifier *notification.Service,
	authSvc *auth.AuthService,
) *Handler {
	engine := tmengine.NewReplayEngine(tmRepo, regRepo, repo, redis, nil)
	recEngine := tmengine.NewReconciliationEngine(tmRepo)
	auditGen := tmengine.NewAuditGenerator(tmRepo, platformKey, "functionfly-platform")

	return &Handler{
		tmRepo:    tmRepo,
		repo:      repo,
		regRepo:   regRepo,
		redis:     redis,
		engine:    engine,
		recEngine: recEngine,
		auditGen:  auditGen,
		notifier:  notifier,
		authSvc:   authSvc,
	}
}

func (h *Handler) SetExecutor(exec tmengine.Executor) {
	h.engine = tmengine.NewReplayEngine(h.tmRepo, h.regRepo, h.repo, h.redis, exec)
	h.engine.SetProgressPublisher(tmengine.NewRedisProgressPublisher(h.redis))
	h.engine.SetOnCompleteCallback(func(replay *tmstorage.Replay) {
		h.sendReplayNotification(replay, replay.Status)
		if replay.Status == "completed" {
			h.recordBillingUsage(replay)
		}
	})
	h.recEngine.SetRegistryRepository(h.regRepo)
}

func (h *Handler) GetEngine() *tmengine.ReplayEngine {
	return h.engine
}

func (h *Handler) requireAuthOrToken(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := middleware.GetUserFromContext(r)
	if claims != nil {
		return claims
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return nil
	}

	if h.authSvc == nil {
		h.writeError(w, http.StatusInternalServerError, "CONFIG_ERROR", "Auth service not available")
		return nil
	}

	parsedClaims, err := h.authSvc.ValidateToken(context.Background(), token)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
		return nil
	}

	return parsedClaims
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code string, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func (h *Handler) getTenantID(r *http.Request) uuid.UUID {
	claims := middleware.GetUserFromContext(r)
	if claims != nil {
		return claims.TenantID
	}
	return uuid.Nil
}

func (h *Handler) getUserID(r *http.Request) uuid.UUID {
	claims := middleware.GetUserFromContext(r)
	if claims != nil {
		return claims.UserID
	}
	return uuid.Nil
}

func (h *Handler) getPlan(r *http.Request) string {
	return middleware.GetTenantPlan(r)
}

func (h *Handler) getPagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return
}

func (h *Handler) sendReplayNotification(replay *tmstorage.Replay, eventType string) {
	if h.notifier == nil || replay.UserID == uuid.Nil {
		return
	}

	var notifType, title, body string
	var priority string

	switch eventType {
	case "completed":
		notifType = "timemachine.replay_completed"
		title = "Time Machine Replay Completed"
		body = fmt.Sprintf("Your replay of %s has completed. %d executions processed, %d changed.",
			replay.Reason, replay.TotalExecutionsReplayed, replay.TotalExecutionsChanged)
		priority = "normal"
	case "failed":
		notifType = "timemachine.replay_failed"
		title = "Time Machine Replay Failed"
		errMsg := ""
		if replay.ErrorMessage.Valid {
			errMsg = ": " + replay.ErrorMessage.String
		}
		body = fmt.Sprintf("Your replay of %s has failed%s.", replay.Reason, errMsg)
		priority = "high"
	default:
		return
	}

	go func() {
		ctx := context.Background()
		_, err := h.notifier.Send(ctx, notification.SendRequest{
			UserID:   replay.UserID,
			Type:     notifType,
			Category: "function",
			Title:    title,
			Body:     body,
			Data: notification.JSONMap{
				"replay_id":    replay.ID.String(),
				"function_id":  replay.FunctionID.String(),
				"reason":       replay.Reason,
				"status":       replay.Status,
				"found":        replay.TotalExecutionsFound,
				"replayed":     replay.TotalExecutionsReplayed,
				"changed":      replay.TotalExecutionsChanged,
				"failed":       replay.TotalExecutionsFailed,
			},
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
			Priority: priority,
		})
		if err != nil {
			logrus.WithError(err).WithField("replay_id", replay.ID).Warn("Failed to send replay notification")
		}
	}()
}

func (h *Handler) publishProgress(replayID uuid.UUID, data map[string]interface{}) {
	if h.redis == nil {
		return
	}
	channel := fmt.Sprintf("timemachine:progress:%s", replayID.String())
	payload, _ := json.Marshal(data)
	h.redis.Publish(context.Background(), channel, payload)
}

func (h *Handler) recordBillingUsage(replay *tmstorage.Replay) {
	if h.repo == nil || replay.TenantID == uuid.Nil {
		return
	}

	go func() {
		ctx := context.Background()

		usageEvent := &storage.UsageEvent{
			TenantID:  replay.TenantID,
			EventType: "time_machine_replay",
			Quantity:  replay.TotalExecutionsReplayed,
			Metadata: map[string]interface{}{
				"replay_id":     replay.ID.String(),
				"function_id":   replay.FunctionID.String(),
				"reason":        replay.Reason,
				"executions":    replay.TotalExecutionsReplayed,
				"changed":       replay.TotalExecutionsChanged,
				"duration_secs": int(time.Since(replay.CreatedAt).Seconds()),
			},
			Timestamp: time.Now().UTC(),
		}
		if err := h.repo.RecordUsageEvent(ctx, usageEvent); err != nil {
			logrus.WithError(err).WithField("replay_id", replay.ID).Warn("Failed to record billing usage for replay")
		}
	}()
}

func (h *Handler) RegisterRoutes(r *mux.Router, fm *middleware.FeatureMiddleware) {
	tm := r.PathPrefix("/time-machine").Subrouter()

	tm.HandleFunc("/replays", h.HandleCreateReplay).Methods("POST")
	tm.HandleFunc("/replays", h.HandleListReplays).Methods("GET")
	tm.HandleFunc("/replays/{id}", h.HandleGetReplay).Methods("GET")
	tm.HandleFunc("/replays/{id}", h.HandleCancelReplay).Methods("DELETE")
	tm.HandleFunc("/replays/{id}/progress", h.HandleReplayProgress).Methods("GET")
	tm.HandleFunc("/replays/{id}/stream", h.HandleReplayStream).Methods("GET")
	tm.HandleFunc("/replays/{id}/items", h.HandleListReplayItems).Methods("GET")
	tm.HandleFunc("/replays/{id}/items/{itemId}", h.HandleGetReplayItem).Methods("GET")

	tm.HandleFunc("/replays/{id}/diff-summary", h.HandleDiffSummary).Methods("GET")

	tm.HandleFunc("/replays/{id}/reconcile", h.HandleStartReconciliation).Methods("POST")
	tm.HandleFunc("/replays/{id}/reconciliations", h.HandleListReconciliations).Methods("GET")

	tm.HandleFunc("/replays/{id}/audit-certificate", h.HandleGetAuditCertificate).Methods("GET")

	tm.HandleFunc("/limits", h.HandleGetLimits).Methods("GET")
}

func (h *Handler) getStableVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error) {
	return h.regRepo.GetLatestFunctionVersion(functionID)
}

func (h *Handler) getPreviousVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error) {
	latest, err := h.regRepo.GetLatestFunctionVersion(functionID)
	if err != nil {
		return nil, err
	}
	return h.regRepo.GetPreviousVersion(functionID, latest.Version)
}
