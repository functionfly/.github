package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/security"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo              HandlerRepo
	agentSearcher     AgentSearcher
	functionSearcher  FunctionSearcher
	agentRater        AgentRater
	functionRater     FunctionRater
}

type HandlerRepo interface {
	List(ctx context.Context, params ListParams) ([]Extension, error)
	Get(ctx context.Context, id string) (*Extension, error)
	Create(ctx context.Context, ext *Extension) error
	Update(ctx context.Context, ext *Extension) error
	Delete(ctx context.Context, id string) error
	IncrementInstallCount(ctx context.Context, id string) error
	GetInstallCounts(ctx context.Context, ids []string) (map[string]int, error)
	GetCategories(ctx context.Context) ([]CategoryCount, error)
	UpsertRating(ctx context.Context, rating *Rating) error
	GetRating(ctx context.Context, extensionID, tenantID string) (*Rating, error)
	ListRatings(ctx context.Context, extensionID string, limit int) ([]Rating, error)
	FindUpdates(ctx context.Context, installed []InstalledPlugin) ([]ExtensionUpdate, error)
	CreatePluginFromExtension(ctx context.Context, tenantID, extensionID string) (*Extension, error)
}

type AgentSearcher interface {
	SearchAgents(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedAgentResult, int64, error)
	SearchFunctions(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedFunctionResult, int64, error)
}

type AgentRater interface {
	RateAgent(ctx context.Context, agentID string, tenantID uuid.UUID, rating int, review string) error
	GetAgentRating(ctx context.Context, agentID string, tenantID uuid.UUID) (*AgentRatingResult, error)
	ListAgentRatings(ctx context.Context, agentID string, limit int) ([]AgentRatingResult, error)
}

type FunctionRater interface {
	RateFunction(ctx context.Context, functionID uuid.UUID, tenantID uuid.UUID, rating int, review string) error
	GetFunctionRating(ctx context.Context, functionID uuid.UUID, tenantID uuid.UUID) (*FunctionRatingResult, error)
	ListFunctionRatings(ctx context.Context, functionID uuid.UUID, limit int) ([]FunctionRatingResult, error)
}

type AgentRatingResult struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	TenantID  string    `json:"tenant_id"`
	Rating    int       `json:"rating"`
	Review    string    `json:"review,omitempty"`
	Username  string    `json:"username,omitempty"`
	UserName  string    `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FunctionRatingResult struct {
	ID         string    `json:"id"`
	FunctionID string    `json:"function_id"`
	TenantID   string    `json:"tenant_id"`
	Rating     int       `json:"rating"`
	Review     string    `json:"review,omitempty"`
	Username   string    `json:"username,omitempty"`
	UserName   string    `json:"user_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FunctionSearcher interface {
	SearchFunctions(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedFunctionResult, int, error)
}

type UnifiedSearchRequest struct {
	Query  string
	Limit  int
	Offset int
}

type UnifiedAgentResult struct {
	AgentID      string
	Name         string
	Description  string
	Capabilities []string
	PricingModel string
	PricePerCall *float64
	SubMonthly   *float64
	RatingScore  float64
	TotalCalls   int
	ROIScore     float64
	ListingType  string
	Verified     bool
	RankScore    float64
}

type UnifiedFunctionResult struct {
	FunctionID   string
	Name         string
	Description  string
	Author       string
	Category     string
	Runtime      string
	PricingModel string
	PricePerCall *float64
	SubMonthly   *float64
	RatingScore  float64
	CallVolume   int
	Verified     bool
	Tags         []string
}

func NewHandler(repo HandlerRepo) *Handler {
	return &Handler{repo: repo}
}

func NewHandlerWithSearchers(repo HandlerRepo, agentSearcher AgentSearcher, functionSearcher FunctionSearcher) *Handler {
	return &Handler{repo: repo, agentSearcher: agentSearcher, functionSearcher: functionSearcher}
}

func (h *Handler) SetAgentSearcher(searcher AgentSearcher) {
	h.agentSearcher = searcher
}

func (h *Handler) SetFunctionSearcher(searcher FunctionSearcher) {
	h.functionSearcher = searcher
}

func (h *Handler) SetAgentRater(rater AgentRater) {
	h.agentRater = rater
}

func (h *Handler) HandleRateAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.agentRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	agentID := mux.Vars(r)["id"]
	if agentID == "" {
		writeJSONError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	var req struct {
		Rating int    `json:"rating"`
		Review string `json:"review"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeJSONError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	if err := h.agentRater.RateAgent(r.Context(), agentID, tid, req.Rating, req.Review); err != nil {
		logrus.WithError(err).Error("marketplace: failed to rate agent")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Rating saved"})
}

func (h *Handler) HandleGetMyAgentRating(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.agentRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	agentID := mux.Vars(r)["id"]
	if agentID == "" {
		writeJSONError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	rating, err := h.agentRater.GetAgentRating(r.Context(), agentID, tid)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get agent rating")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"rating": rating})
}

func (h *Handler) HandleListAgentRatings(w http.ResponseWriter, r *http.Request) {
	if h.agentRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	agentID := mux.Vars(r)["id"]
	if agentID == "" {
		writeJSONError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	ratings, err := h.agentRater.ListAgentRatings(r.Context(), agentID, limit)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to list agent ratings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list ratings")
		return
	}

	if ratings == nil {
		ratings = []AgentRatingResult{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ratings": ratings})
}

func (h *Handler) SetFunctionRater(rater FunctionRater) {
	h.functionRater = rater
}

func (h *Handler) HandleRateFunction(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.functionRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	funcID := mux.Vars(r)["id"]
	if funcID == "" {
		writeJSONError(w, http.StatusBadRequest, "function id is required")
		return
	}

	fid, err := uuid.Parse(funcID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid function ID")
		return
	}

	var req struct {
		Rating int    `json:"rating"`
		Review string `json:"review"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeJSONError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	if err := h.functionRater.RateFunction(r.Context(), fid, tid, req.Rating, req.Review); err != nil {
		logrus.WithError(err).Error("marketplace: failed to rate function")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Rating saved"})
}

func (h *Handler) HandleGetMyFunctionRating(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.functionRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	funcID := mux.Vars(r)["id"]
	if funcID == "" {
		writeJSONError(w, http.StatusBadRequest, "function id is required")
		return
	}

	fid, err := uuid.Parse(funcID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid function ID")
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	rating, err := h.functionRater.GetFunctionRating(r.Context(), fid, tid)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get function rating")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"rating": rating})
}

func (h *Handler) HandleListFunctionRatings(w http.ResponseWriter, r *http.Request) {
	if h.functionRater == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Rating service not configured")
		return
	}

	funcID := mux.Vars(r)["id"]
	if funcID == "" {
		writeJSONError(w, http.StatusBadRequest, "function id is required")
		return
	}

	fid, err := uuid.Parse(funcID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid function ID")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	ratings, err := h.functionRater.ListFunctionRatings(r.Context(), fid, limit)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to list function ratings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list ratings")
		return
	}

	if ratings == nil {
		ratings = []FunctionRatingResult{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ratings": ratings})
}

func (h *Handler) HandleListExtensions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	featured := r.URL.Query().Get("featured") == "true"
	tagsParam := r.URL.Query().Get("tags")
	sortBy := r.URL.Query().Get("sort")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	params := ListParams{
		Category: categoryPtr,
		Status:   statusPtr,
		Search:   searchPtr,
		Featured: &featured,
		Tags:     tags,
		SortBy:   sortBy,
		Limit:    limit,
		Offset:   offset,
	}

	extensions, err := h.repo.List(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("marketplace: failed to list extensions")
		writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": []Extension{}})
		return
	}

	if extensions == nil {
		extensions = []Extension{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": extensions})
}

func (h *Handler) HandleGetExtension(w http.ResponseWriter, r *http.Request) {
	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleCreateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateExtensionRequest(req.Name, req.Version, req.Manifest); err != nil {
		logrus.WithError(err).WithField("name", req.Name).Info("extension validation failed")
		writeJSONError(w, http.StatusBadRequest, "Invalid extension request")
		return
	}

	ext := &Extension{
		CreatorID:     tenantID,
		Name:          req.Name,
		Version:       req.Version,
		Description:   req.Description,
		Category:      req.Category,
		IconURL:       req.IconURL,
		Manifest:      req.Manifest,
		Status:        "draft",
		Featured:      false,
		InstallCount:  0,
		RatingAverage: 0,
		RatingCount:   0,
		TrustScore:    0,
	}

	if err := h.repo.Create(r.Context(), ext); err != nil {
		logrus.WithError(err).Error("marketplace: failed to create extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create extension")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleUpdateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	var req UpdateExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name != "" {
		ext.Name = req.Name
	}
	if req.Version != "" {
		ext.Version = req.Version
	}
	if req.Description != "" {
		ext.Description = req.Description
	}
	if req.Category != "" {
		ext.Category = req.Category
	}
	if req.IconURL != "" {
		ext.IconURL = req.IconURL
	}
	if req.Manifest != nil {
		ext.Manifest = req.Manifest
	}
	if req.Status != "" {
		ext.Status = req.Status
	}
	if req.Featured != nil {
		ext.Featured = *req.Featured
	}

	if err := h.repo.Update(r.Context(), ext); err != nil {
		logrus.WithError(err).Error("marketplace: failed to update extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleDeleteExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	if ext.CreatorID != tenantID {
		writeJSONError(w, http.StatusForbidden, "Not authorized to delete this extension")
		return
	}

	if err := h.repo.Delete(r.Context(), extID); err != nil {
		logrus.WithError(err).Error("marketplace: failed to delete extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension deleted"})
}

func (h *Handler) HandleInstallExtension(w http.ResponseWriter, r *http.Request) {
	h.HandleInstallExtensionWithPlugin(w, r)
}

func (h *Handler) HandleGetInstallCounts(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		writeJSONError(w, http.StatusBadRequest, "ids parameter is required")
		return
	}

	ids := strings.Split(idsParam, ",")
	if len(ids) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"install_counts": map[string]int{}})
		return
	}

	counts, err := h.repo.GetInstallCounts(r.Context(), ids)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get install counts")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get install counts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"install_counts": counts})
}

func (h *Handler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get categories")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	if categories == nil {
		categories = []CategoryCount{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": categories})
}

type UnifiedItem struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	IconURL      *string                `json:"icon_url"`
	Rating       float64                `json:"rating"`
	InstallCount int                    `json:"install_count"`
	Price        string                 `json:"price"`
	PricingModel string                 `json:"pricing_model"`
	Tags         []string               `json:"tags"`
	Verified     bool                   `json:"verified"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type UnifiedSearchResponse struct {
	Items   []UnifiedItem `json:"items"`
	Total   int           `json:"total"`
	HasMore bool          `json:"has_more"`
}

func (h *Handler) HandleUnifiedSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	typeFilter := r.URL.Query().Get("type")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	searchReq := UnifiedSearchRequest{Query: q, Limit: limit, Offset: offset}
	var items []UnifiedItem
	var total int

	searchExtensions := typeFilter == "" || typeFilter == "extension"
	searchAgents := typeFilter == "" || typeFilter == "agent"
	searchFunctions := typeFilter == "" || typeFilter == "function"

	type extResult struct {
		items []UnifiedItem
		total int
		err   error
	}
	type agentResult struct {
		items []UnifiedItem
		total int64
		err   error
	}
	type funcResult struct {
		items []UnifiedItem
		total int
		err   error
	}

	extCh := make(chan extResult, 1)
	agentCh := make(chan agentResult, 1)
	funcCh := make(chan funcResult, 1)

	if searchExtensions {
		go func() {
			status := "active"
			exts, err := h.repo.List(r.Context(), ListParams{
				Search: &q,
				Status: &status,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				extCh <- extResult{err: err}
				return
			}
			var unified []UnifiedItem
			for _, ext := range exts {
				price := "Free"
				iconURL := &ext.IconURL
				if ext.IconURL == "" {
					iconURL = nil
				}
				tags := ext.Tags
				if tags == nil {
					tags = []string{}
				}
				unified = append(unified, UnifiedItem{
					Type:         "extension",
					ID:           ext.ID,
					Name:         ext.Name,
					Description:  ext.Description,
					IconURL:      iconURL,
					Rating:       ext.RatingAverage,
					InstallCount: ext.InstallCount,
					Price:        price,
					PricingModel: "free",
					Tags:         tags,
					Verified:     ext.Verified,
					Metadata: map[string]interface{}{
						"version":      ext.Version,
						"category":     ext.Category,
						"author_id":    ext.CreatorID,
						"trust_score":  ext.TrustScore,
						"status":       ext.Status,
						"rating_count": ext.RatingCount,
					},
				})
			}
			extCh <- extResult{items: unified, total: len(exts)}
		}()
	} else {
		extCh <- extResult{}
	}

	if searchAgents && h.agentSearcher != nil {
		go func() {
			agents, totalAgents, err := h.agentSearcher.SearchAgents(r.Context(), searchReq)
			if err != nil {
				agentCh <- agentResult{err: err}
				return
			}
			var unified []UnifiedItem
			for _, a := range agents {
				price := formatAgentPrice(a.PricingModel, a.PricePerCall, a.SubMonthly)
				tags := a.Capabilities
				if tags == nil {
					tags = []string{}
				}
				unified = append(unified, UnifiedItem{
					Type:         "agent",
					ID:           a.AgentID,
					Name:         a.Name,
					Description:  a.Description,
					IconURL:      nil,
					Rating:       a.RatingScore,
					InstallCount: a.TotalCalls,
					Price:        price,
					PricingModel: a.PricingModel,
					Tags:         tags,
					Verified:     a.Verified,
					Metadata: map[string]interface{}{
						"listing_type": a.ListingType,
						"roi_score":    a.ROIScore,
						"rank_score":   a.RankScore,
					},
				})
			}
			agentCh <- agentResult{items: unified, total: totalAgents}
		}()
	} else {
		agentCh <- agentResult{}
	}

	if searchFunctions && h.agentSearcher != nil {
		go func() {
			funcs, totalFuncs, err := h.agentSearcher.SearchFunctions(r.Context(), searchReq)
			if err != nil {
				funcCh <- funcResult{err: err}
				return
			}
			var unified []UnifiedItem
			for _, f := range funcs {
				price := formatFuncPrice(f.PricingModel, f.PricePerCall, f.SubMonthly)
				tags := f.Tags
				if tags == nil {
					tags = []string{}
				}
				if f.Runtime != "" {
					tags = append(tags, f.Runtime)
				}
				unified = append(unified, UnifiedItem{
					Type:         "function",
					ID:           f.FunctionID,
					Name:         f.Name,
					Description:  f.Description,
					IconURL:      nil,
					Rating:       f.RatingScore,
					InstallCount: f.CallVolume,
					Price:        price,
					PricingModel: f.PricingModel,
					Tags:         tags,
					Verified:     f.Verified,
					Metadata: map[string]interface{}{
						"author":   f.Author,
						"category": f.Category,
						"runtime":  f.Runtime,
					},
				})
			}
			funcCh <- funcResult{items: unified, total: int(totalFuncs)}
		}()
	} else {
		funcCh <- funcResult{}
	}

	er := <-extCh
	ar := <-agentCh
	fr := <-funcCh

	if er.err != nil {
		logrus.WithError(er.err).Warn("marketplace: unified search extension query failed")
	}
	if ar.err != nil {
		logrus.WithError(ar.err).Warn("marketplace: unified search agent query failed")
	}
	if fr.err != nil {
		logrus.WithError(fr.err).Warn("marketplace: unified search function query failed")
	}

	items = append(items, er.items...)
	items = append(items, ar.items...)
	items = append(items, fr.items...)
	total = len(items) + int(ar.total) + int(fr.total) - len(ar.items) - len(fr.items)
	if total < len(items) {
		total = len(items)
	}

	writeJSON(w, http.StatusOK, UnifiedSearchResponse{
		Items:   items,
		Total:   total,
		HasMore: total > offset+limit,
	})
}

func formatFuncPrice(pricingModel string, pricePerCall *float64, subMonthly *float64) string {
	if pricingModel == "" || pricingModel == "free" {
		if pricePerCall != nil && *pricePerCall > 0 {
			return fmt.Sprintf("$%.4f/call", *pricePerCall)
		}
		return "Free"
	}
	return formatAgentPrice(pricingModel, pricePerCall, subMonthly)
}

func formatAgentPrice(pricingModel string, pricePerCall, subMonthly *float64) string {
	switch pricingModel {
	case "free":
		return "Free"
	case "per_call":
		if pricePerCall != nil {
			return fmt.Sprintf("$%.4f/call", *pricePerCall)
		}
		return "Per Call"
	case "subscription":
		if subMonthly != nil {
			return fmt.Sprintf("$%.2f/mo", *subMonthly)
		}
		return "Subscription"
	case "revenue_share":
		return "Revenue Share"
	default:
		if pricingModel != "" {
			return pricingModel
		}
		return "Free"
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getTenantID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.TenantID.String()
}

type CreateExtensionRequest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	IconURL     string                 `json:"icon_url"`
	Manifest    map[string]interface{} `json:"manifest"`
	Tags        []string               `json:"tags"`
}

type UpdateExtensionRequest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	IconURL     string                 `json:"icon_url"`
	Manifest    map[string]interface{} `json:"manifest"`
	Status      string                 `json:"status"`
	Featured    *bool                  `json:"featured"`
	Tags        []string               `json:"tags"`
}

type Rating struct {
	ID          string    `json:"id"`
	ExtensionID string    `json:"extension_id"`
	TenantID    string    `json:"tenant_id"`
	Rating      int       `json:"rating"`
	Review      string    `json:"review,omitempty"`
	Username    string    `json:"username,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RateRequest struct {
	Rating int    `json:"rating"`
	Review string `json:"review"`
}

type InstalledPlugin struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ExtensionUpdate struct {
	InstalledPluginID string                 `json:"installed_plugin_id"`
	InstalledVersion  string                 `json:"installed_version"`
	ExtensionID       string                 `json:"extension_id"`
	LatestVersion     string                 `json:"latest_version"`
	Changelog         string                 `json:"changelog"`
	Manifest          map[string]interface{} `json:"manifest"`
}

func (h *Handler) HandleRateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to rate extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		writeJSONError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	rating := &Rating{
		ExtensionID: extID,
		TenantID:    tenantID,
		Rating:      req.Rating,
		Review:      req.Review,
	}
	if err := h.repo.UpsertRating(r.Context(), rating); err != nil {
		logrus.WithError(err).Error("marketplace: failed to upsert rating")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Rating saved", "rating": rating})
}

func (h *Handler) HandleGetMyRating(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	rating, err := h.repo.GetRating(r.Context(), extID, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get rating")
		return
	}

	if rating == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"rating": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"rating": rating})
}

func (h *Handler) HandleListRatings(w http.ResponseWriter, r *http.Request) {
	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	ratings, err := h.repo.ListRatings(r.Context(), extID, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list ratings")
		return
	}

	if ratings == nil {
		ratings = []Rating{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ratings": ratings})
}

func (h *Handler) HandleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var installed []InstalledPlugin
	if err := json.NewDecoder(r.Body).Decode(&installed); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates, err := h.repo.FindUpdates(r.Context(), installed)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to find updates")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check for updates")
		return
	}

	if updates == nil {
		updates = []ExtensionUpdate{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"updates": updates})
}

func (h *Handler) HandleInstallExtensionWithPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension for install")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	analyzer := security.NewPluginAnalyzer()
	manifest := &security.PluginManifest{}
	if err := mapToStruct(ext.Manifest, manifest); err == nil && manifest.Name != "" {
		if result, err := analyzer.AnalyzeManifest(r.Context(), manifest); err == nil {
			ext.TrustScore = result.Score
			ext.SecurityScore = result.Score
		}
	}

	created, err := h.repo.CreatePluginFromExtension(r.Context(), tenantID, extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to install extension as plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install extension")
		return
	}
	if created == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":         "Extension installed",
		"extension":       created,
		"extension_id":    extID,
		"plugin_manifest": created.Manifest,
	})
}

func validateExtensionRequest(name, version string, manifest map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if version == "" {
		return fmt.Errorf("version is required")
	}
	if !isValidSemver(version) {
		return fmt.Errorf("version must be valid semver (e.g. 1.0.0)")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	analyzer := security.NewPluginAnalyzer()
	parsed := &security.PluginManifest{}
	_ = mapToStruct(manifest, parsed)
	parsed.Name = name
	parsed.Version = version

	result, err := analyzer.AnalyzeManifest(context.Background(), parsed)
	if err != nil {
		return fmt.Errorf("manifest analysis failed: %w", err)
	}
	if !result.Safe {
		return fmt.Errorf("manifest failed security analysis (score: %.1f): %v", result.Score, result.Issues)
	}
	return nil
}

func isValidSemver(v string) bool {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func mapToStruct(m map[string]interface{}, dst interface{}) error {
	if m == nil {
		return fmt.Errorf("nil map")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
