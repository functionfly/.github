package agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UnfairAdvantageHandler struct {
	engine *swarm.UnfairAdvantageEngine
}

func NewUnfairAdvantageHandler(engine *swarm.UnfairAdvantageEngine) *UnfairAdvantageHandler {
	return &UnfairAdvantageHandler{engine: engine}
}

func (h *UnfairAdvantageHandler) RegisterRoutes(router *mux.Router, basePath string) {
	logrus.Info("UnfairAdvantageHandler: Registering unfair advantage engine routes (ADMIN ONLY)")

	advantage := router.PathPrefix(basePath + "/advantage").Subrouter()

	advantage.HandleFunc("/dashboard", h.HandleGetDashboard).Methods("GET", "OPTIONS")
	advantage.HandleFunc("/value-report", h.HandleGetValueReport).Methods("GET", "OPTIONS")

	advantage.HandleFunc("/opportunities", h.HandleListInternalOpportunities).Methods("GET", "OPTIONS")
	advantage.HandleFunc("/opportunities/seed", h.HandleSeedOpportunity).Methods("POST", "OPTIONS")
	advantage.HandleFunc("/opportunities/custom", h.HandleSeedCustomOpportunity).Methods("POST", "OPTIONS")

	advantage.HandleFunc("/rdlab/run", h.HandleRunRDLab).Methods("POST", "OPTIONS")

	advantage.HandleFunc("/functions/generate", h.HandleGenerateInternalFunction).Methods("POST", "OPTIONS")
	advantage.HandleFunc("/functions", h.HandleListInternalFunctions).Methods("GET", "OPTIONS")

	advantage.HandleFunc("/stealth/run", h.HandleRunStealthPipeline).Methods("POST", "OPTIONS")

	logrus.Info("UnfairAdvantageHandler: All unfair advantage routes registered (ADMIN ONLY)")
}

func (h *UnfairAdvantageHandler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.engine.GetDashboard(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"dashboard": dashboard,
		"retrieved_at": time.Now(),
	})
}

func (h *UnfairAdvantageHandler) HandleGetValueReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.engine.GetValueReport(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"report": report,
	})
}

func (h *UnfairAdvantageHandler) HandleListInternalOpportunities(w http.ResponseWriter, r *http.Request) {
	filter := &swarm.OpportunityFilter{
		Limit: 20,
		Offset: 0,
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if n := parseInt(l); n > 0 && n <= 100 {
			filter.Limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n := parseInt(o); n >= 0 {
			filter.Offset = n
		}
	}
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = category
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filter.Priority = priority
	}
	if r.URL.Query().Get("confidential_only") == "true" {
		filter.ConfidentialOnly = true
	}

	opportunities, total, err := h.engine.ListInternalOpportunities(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"opportunities": opportunities,
		"total":        total,
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	})
}

func (h *UnfairAdvantageHandler) HandleSeedOpportunity(w http.ResponseWriter, r *http.Request) {
	var req swarm.SeedOpportunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title is required")
		return
	}

	opportunity, err := h.engine.SeedInternalOpportunity(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":          true,
		"opportunity": opportunity,
		"message":     "Internal opportunity seeded successfully",
	})
}

func (h *UnfairAdvantageHandler) HandleSeedCustomOpportunity(w http.ResponseWriter, r *http.Request) {
	var seed swarm.CustomOpportunitySeed
	if err := json.NewDecoder(r.Body).Decode(&seed); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if seed.Title == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title is required")
		return
	}

	opportunity, err := h.engine.SeedCustomOpportunity(r.Context(), &seed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":          true,
		"opportunity": opportunity,
		"message":     "Custom opportunity seeded into factory pipeline",
	})
}

func (h *UnfairAdvantageHandler) HandleRunRDLab(w http.ResponseWriter, r *http.Request) {
	run, err := h.engine.RunInternalRDLab(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"run":    run,
		"message": "R&D Lab run completed",
	})
}

func (h *UnfairAdvantageHandler) HandleGenerateInternalFunction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OpportunityID string `json:"opportunity_id" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	oppID, err := uuid.Parse(req.OpportunityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid opportunity_id")
		return
	}

	function, err := h.engine.GenerateInternalFunction(r.Context(), oppID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":       true,
		"function": function,
		"message":  "Internal function generated from opportunity",
	})
}

func (h *UnfairAdvantageHandler) HandleListInternalFunctions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"functions": []interface{}{},
		"total":    0,
	})
}

func (h *UnfairAdvantageHandler) HandleRunStealthPipeline(w http.ResponseWriter, r *http.Request) {
	var cfg swarm.StealthConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if cfg.Mode == "" {
		cfg.Mode = "normal"
	}
	if cfg.ScanDepth == 0 {
		cfg.ScanDepth = 5
	}
	if cfg.GeneratorWorkers == 0 {
		cfg.GeneratorWorkers = 3
	}
	if cfg.QualityFloor == 0 {
		cfg.QualityFloor = 80
	}

	run, err := h.engine.RunStealthPipeline(r.Context(), &cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":     true,
		"run":    run,
		"message": "Stealth pipeline run initiated",
	})
}