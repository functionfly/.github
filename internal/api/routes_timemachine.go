package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/timemachine"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/dre"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func registerTimeMachineRoutes(
	api *mux.Router,
	tmHandler *timemachine.Handler,
	authMiddleware *middleware.AuthMiddleware,
	fm *middleware.FeatureMiddleware,
) {
	tm := api.PathPrefix("/time-machine").Subrouter()

	// Basic Time Machine (time_machine_basic) - available to Free+ plans
	tm.HandleFunc("/replays", authMiddleware.RequireAuth(tmHandler.HandleCreateReplay)).Methods("POST", "OPTIONS")
	tm.HandleFunc("/replays", authMiddleware.RequireAuth(tmHandler.HandleListReplays)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}", authMiddleware.RequireAuth(tmHandler.HandleGetReplay)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}", authMiddleware.RequireAuth(tmHandler.HandleCancelReplay)).Methods("DELETE", "OPTIONS")
	tm.HandleFunc("/replays/{id}/progress", authMiddleware.RequireAuth(tmHandler.HandleReplayProgress)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}/stream", authMiddleware.RequireAuth(tmHandler.HandleReplayStream)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}/items", authMiddleware.RequireAuth(tmHandler.HandleListReplayItems)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}/items/{itemId}", authMiddleware.RequireAuth(tmHandler.HandleGetReplayItem)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/replays/{id}/diff-summary", authMiddleware.RequireAuth(tmHandler.HandleDiffSummary)).Methods("GET", "OPTIONS")
	tm.HandleFunc("/limits", authMiddleware.RequireAuth(tmHandler.HandleGetLimits)).Methods("GET", "OPTIONS")

	// Pro features (time_machine_pro) - reconciliation
	tm.HandleFunc("/replays/{id}/reconcile",
		fm.RequireFeature(plans.FeatureTimeMachinePro)(http.HandlerFunc(authMiddleware.RequireAuth(tmHandler.HandleStartReconciliation))).ServeHTTP,
	).Methods("POST", "OPTIONS")
	tm.HandleFunc("/replays/{id}/reconciliations",
		fm.RequireFeature(plans.FeatureTimeMachinePro)(http.HandlerFunc(authMiddleware.RequireAuth(tmHandler.HandleListReconciliations))).ServeHTTP,
	).Methods("GET", "OPTIONS")

	// Enterprise features (time_machine_enterprise) - audit certificates
	tm.HandleFunc("/replays/{id}/audit-certificate",
		fm.RequireFeature(plans.FeatureTimeMachineEnterprise)(http.HandlerFunc(authMiddleware.RequireAuth(tmHandler.HandleGetAuditCertificate))).ServeHTTP,
	).Methods("GET", "OPTIONS")
}

func newTimeMachineHandler(
	tmRepo *tmstorage.Repository,
	repo storage.Repository,
	redisClient *redis.Client,
	realtimeUsageTracker *services.RealtimeUsageTracker,
	notifier *notification.Service,
	authSvc *auth.AuthService,
) *timemachine.Handler {
	platformKey, _ := dre.LoadPlatformKeyFromEnv()
	return timemachine.NewHandler(tmRepo, repo, nil, redisClient, platformKey, notifier, authSvc)
}
