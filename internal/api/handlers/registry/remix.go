package registry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RemixCost is the fee charged for remixing a function (in USD)
const RemixCostUSD = 0.50

// RemixRequest represents a request to remix/fork a function
type RemixRequest struct {
	SourceAuthor    string `json:"source_author"`
	SourceName      string `json:"source_name"`
	TargetTenantID  string `json:"target_tenant_id"`
	NewName         string `json:"new_name,omitempty"`
	Customization   string `json:"customization,omitempty"`    // Optional description of customizations
	PrivateFunction bool   `json:"private_function,omitempty"` // Whether to create as private function (default: true)
}

// RemixResponse represents the response from a remix operation
type RemixResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	RemixID       string  `json:"remix_id"`
	NewFunctionID string  `json:"new_function_id,omitempty"`
	NewAuthor     string  `json:"new_author,omitempty"`
	NewName       string  `json:"new_name,omitempty"`
	CostUSD       float64 `json:"cost_usd,omitempty"`
	NewBalanceUSD float64 `json:"new_balance_usd,omitempty"`
}

// HandleRemix handles remixing/forking a public gallery function to a user's private functions
func (h *Handler) HandleRemix(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	var req RemixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use URL params as defaults if not in body
	if req.SourceAuthor == "" {
		req.SourceAuthor = author
	}
	if req.SourceName == "" {
		req.SourceName = name
	}
	if req.TargetTenantID == "" {
		req.TargetTenantID = user.TenantID.String()
	}

	// Validate source function exists and is public
	sourceFn, err := h.repo.GetFunctionByAuthorName(req.SourceAuthor, req.SourceName)
	if err != nil {
		if isRecordNotFound(err) {
			http.Error(w, "Source function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get source function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Only allow remixing public functions
	if sourceFn.Visibility != "public" {
		http.Error(w, "Cannot remix non-public functions", http.StatusForbidden)
		return
	}

	// Check wallet balance and charge for remix
	if h.platformFeeRepo == nil {
		http.Error(w, "Billing system unavailable", http.StatusServiceUnavailable)
		return
	}

	wallet, err := h.platformFeeRepo.GetOrCreateWallet(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user wallet for remix")
		http.Error(w, "Failed to check wallet balance", http.StatusInternalServerError)
		return
	}

	if wallet.BalanceUSD < RemixCostUSD {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        "Insufficient balance",
			"message":      fmt.Sprintf("Remixing costs $%.2f. Your balance is $%.2f. Please add funds to your wallet.", RemixCostUSD, wallet.BalanceUSD),
			"cost_usd":     RemixCostUSD,
			"balance_usd":  wallet.BalanceUSD,
			"required_usd": RemixCostUSD - wallet.BalanceUSD,
		})
		return
	}

	// Charge for the remix
	if err := h.platformFeeRepo.DebitWallet(r.Context(), user.UserID, RemixCostUSD,
		fmt.Sprintf("Remix of %s/%s", req.SourceAuthor, req.SourceName)); err != nil {
		logrus.WithError(err).Error("Failed to charge wallet for remix")
		http.Error(w, "Failed to process payment for remix", http.StatusInternalServerError)
		return
	}

	// Record the platform fee
	fee := &storageregistry.PlatformFee{
		ID:              uuid.New(),
		FunctionID:      sourceFn.ID,
		UserID:          user.UserID,
		FeeType:         "remix",
		AmountUSD:       RemixCostUSD,
		ChargedAt:       time.Now(),
		Status:          "completed",
		StripePaymentID: "",
	}
	if err := h.platformFeeRepo.RecordPlatformFee(r.Context(), fee); err != nil {
		logrus.WithError(err).Warn("Failed to record remix platform fee")
		// Don't fail the request, just log the error
	}

	// Get the latest version
	sourceVersion, err := h.repo.GetLatestFunctionVersion(sourceFn.ID)
	if err != nil {
		http.Error(w, "No versions available for source function", http.StatusNotFound)
		return
	}

	// Get the target tenant ID
	targetTenantID, err := uuid.Parse(req.TargetTenantID)
	if err != nil {
		http.Error(w, "Invalid target tenant ID", http.StatusBadRequest)
		return
	}

	// Verify user has access to target tenant (primary tenant or via membership)
	hasAccess, err := h.backendRepo.IsUserInTenant(r.Context(), user.UserID, targetTenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check tenant access")
		http.Error(w, "Failed to verify tenant access", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Cannot remix to tenant you don't have access to", http.StatusForbidden)
		return
	}

	// Generate new name if not provided
	newName := req.NewName
	if newName == "" {
		// Generate a remix name: "original-remix-{shortid}"
		newName = fmt.Sprintf("%s-remix-%s", req.SourceName, uuid.New().String()[:8])
	}

	// Get the source manifest
	var sourceManifest functionregistry.FunctionManifest
	if err := json.Unmarshal(sourceVersion.Manifest, &sourceManifest); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal source manifest")
		http.Error(w, "Invalid source manifest", http.StatusInternalServerError)
		return
	}

	// Get source code
	var sourceCode string
	if sourceVersion.SourceCode.Valid {
		sourceCode = sourceVersion.SourceCode.String
	}

	// Determine if we're creating a private function or a public one
	private := req.PrivateFunction
	if !private {
		// Default to creating a private function
		private = true
	}

	var newFnID uuid.UUID

	if private {
		// Create as private function using the function repository
		// Import storage package here
		remixID := uuid.New()

		fn := &storage.FunctionConfig{
			ID:           remixID,
			TenantID:     targetTenantID,
			Name:         newName,
			Providers:    []string{"functionfly"}, // Default provider
			Region:       "auto",
			Code:         sourceCode,
			EnvVars:      []storage.EnvironmentVariable{},
			Version:      "1.0.0",
			Status:       "draft",
			Capabilities: []string{},
		}

		ctx := r.Context()
		_, err := h.functionRepo.CreateFunction(ctx, fn)
		if err != nil {
			logrus.WithError(err).Error("Failed to create remixed function")
			http.Error(w, "Failed to create function", http.StatusInternalServerError)
			return
		}
		newFnID = fn.ID
	} else {
		// Create as a new public registry function
		// This is a "remix to public" scenario
		newAuthor := user.Email
		if newAuthor == "" {
			newAuthor = user.Username
		}

		// Prepare tags
		tagsJSON, _ := json.Marshal([]string{"remix", fmt.Sprintf("original:%s/%s", req.SourceAuthor, req.SourceName)})

		// Create description noting it's a remix
		description := sourceFn.Description.String
		if req.Customization != "" {
			description = fmt.Sprintf("%s\n\nRemixed from %s/%s with customizations: %s",
				description, req.SourceAuthor, req.SourceName, req.Customization)
		} else {
			description = fmt.Sprintf("%s\n\nRemixed from %s/%s",
				description, req.SourceAuthor, req.SourceName)
		}

		fn := &storageregistry.RegistryFunction{
			Author:      newAuthor,
			Name:        newName,
			Title:       sourceFn.Title,
			Description: sql.NullString{String: description, Valid: true},
			Category:    sourceFn.Category,
			Tags:        tagsJSON,
			Visibility:  "public",
			TenantID:    &targetTenantID,
			OwnerUserID: &user.UserID,
		}

		if err := h.repo.CreateFunction(fn); err != nil {
			logrus.WithError(err).Error("Failed to create remixed function")
			http.Error(w, "Failed to create function", http.StatusInternalServerError)
			return
		}
		newFnID = fn.ID

		// Create the version
		now := time.Now()
		version := &storageregistry.RegistryFunctionVersion{
			ID:            uuid.New(),
			FunctionID:    newFnID,
			Version:       "1.0.0",
			Manifest:      sourceVersion.Manifest,
			Runtime:       sourceVersion.Runtime,
			MemoryMB:      sourceVersion.MemoryMB,
			TimeoutMs:     sourceVersion.TimeoutMs,
			Deterministic: sourceVersion.Deterministic,
			SideEffects:   sourceVersion.SideEffects,
			Idempotent:    sourceVersion.Idempotent,
			SourceCode:    sourceVersion.SourceCode,
			PublishedAt:   now,
		}

		if err := h.repo.CreateFunctionVersion(version); err != nil {
			logrus.WithError(err).Error("Failed to create function version")
			http.Error(w, "Failed to create function version", http.StatusInternalServerError)
			return
		}
	}

	// Record the remix relationship in the remix_history table
	if err := h.repo.RecordRemix(sourceFn.ID, newFnID, user.UserID, req.Customization, RemixCostUSD); err != nil {
		logrus.WithError(err).Warn("Failed to record remix history")
		// Don't fail the request, but log the error
	}

	// Get actual remix count from database
	remixCount, err := h.repo.CountRemixesForFunction(sourceFn.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to count remixes, using popularity as fallback")
		remixCount = int64(sourceFn.PopularityScore)
	}

	// Update source function's remix count (using popularity_score as remix_count for now)
	if _, err := h.repo.UpdateRegistryFunction(sourceFn.ID, map[string]interface{}{
		"popularity_score": int(remixCount),
	}); err != nil {
		logrus.WithError(err).Warn("Failed to update source function remix count")
	}

	// Get the new author name
	newAuthor := user.Email
	if newAuthor == "" {
		newAuthor = user.Username
	}

	// Get updated wallet balance
	updatedWallet, _ := h.platformFeeRepo.GetWallet(r.Context(), user.UserID)
	newBalance := wallet.BalanceUSD - RemixCostUSD
	if updatedWallet != nil {
		newBalance = updatedWallet.BalanceUSD
	}

	response := RemixResponse{
		Success:       true,
		Message:       fmt.Sprintf("Successfully remixed %s/%s", req.SourceAuthor, req.SourceName),
		RemixID:       uuid.New().String(),
		NewFunctionID: newFnID.String(),
		NewAuthor:     newAuthor,
		NewName:       newName,
		CostUSD:       RemixCostUSD,
		NewBalanceUSD: newBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetRemixHistory handles getting the remix history for a function
func (h *Handler) HandleGetRemixHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Get remix history from database
	history, err := h.repo.GetRemixHistoryForFunction(fn.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get remix history from database")
		// Return empty history but don't fail the request
		history = []storageregistry.RemixHistory{}
	}

	// Count remixes
	remixCount, err := h.repo.CountRemixesForFunction(fn.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to count remixes")
		remixCount = int64(len(history))
	}

	// Format history for response
	remixHistory := make([]map[string]interface{}, 0, len(history))
	for _, h := range history {
		entry := map[string]interface{}{
			"remix_id":      h.ID.String(),
			"remixed_at":    h.RemixedAt.Format(time.RFC3339),
			"customization": h.Customization,
			"remixed_by":    h.RemixedByUserID.String(),
			"cost_usd":      h.CostUSD,
		}
		// Include target function info if available
		if h.TargetFunction != nil {
			entry["target_author"] = h.TargetFunction.Author
			entry["target_name"] = h.TargetFunction.Name
		}
		remixHistory = append(remixHistory, entry)
	}

	// Check if this function is a remix by looking for it as a target in remix_history
	// or by checking tags (for backward compatibility)
	isRemix := false
	var sourceAuthor, sourceName string

	// Check tags for remix indicator
	var tags []string
	if err := json.Unmarshal(fn.Tags, &tags); err == nil {
		for _, tag := range tags {
			if len(tag) > 9 && tag[:9] == "original:" {
				// Extract original author/name from tag
				parts := strings.Split(tag[9:], "/")
				if len(parts) == 2 {
					isRemix = true
					sourceAuthor = parts[0]
					sourceName = parts[1]
					break
				}
			}
		}
	}

	// Check if this function appears as a target in remix_history
	// This indicates it was remixed from something else
	if !isRemix {
		if isRemixResult, history, err := h.repo.IsFunctionRemix(fn.ID); err == nil && isRemixResult && history != nil {
			isRemix = true
			// Get source function info
			if history.SourceFunction != nil {
				sourceAuthor = history.SourceFunction.Author
				sourceName = history.SourceFunction.Name
			}
		}
	}

	response := map[string]interface{}{
		"function_id":   fn.ID.String(),
		"author":        author,
		"name":          name,
		"remix_count":   remixCount,
		"is_remix":      isRemix,
		"source_author": sourceAuthor,
		"source_name":   sourceName,
		"remix_history": remixHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetRemixCost returns the cost to remix a function and the user's balance
func (h *Handler) HandleGetRemixCost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Verify function exists and is public
	sourceFn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}
	if sourceFn.Visibility != "public" {
		http.Error(w, "Cannot remix non-public functions", http.StatusForbidden)
		return
	}

	// Get wallet balance
	balance := 0.0
	if h.platformFeeRepo != nil {
		wallet, err := h.platformFeeRepo.GetWallet(r.Context(), user.UserID)
		if err == nil && wallet != nil {
			balance = wallet.BalanceUSD
		}
	}

	response := map[string]interface{}{
		"cost_usd":        RemixCostUSD,
		"balance_usd":     balance,
		"can_remix":       balance >= RemixCostUSD,
		"required_usd":    RemixCostUSD,
		"function_author": author,
		"function_name":   name,
		"is_own_function": sourceFn.Author == user.Username || sourceFn.Author == user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTrendingFunctions returns trending/remixed functions
func (h *Handler) HandleGetTrendingFunctions(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if limit > 1000 {
		limit = 1000
	}

	// Get functions sorted by popularity score from database
	functions, total, err := h.repo.ListTrendingFunctions(limit, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list trending functions")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	funcInfos, err := h.buildRegistryFunctionInfos(functions)
	if err != nil {
		logrus.WithError(err).Error("Failed to build function infos")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	response := functionregistry.ListFunctionsResponse{
		Functions: convertToFunctionInfos(funcInfos),
		Total:     total,
		Limit:     limit,
		Offset:    0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleLikeFunction handles liking/unliking a function (social feature)
func (h *Handler) HandleLikeFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Toggle like in database
	liked, likeCount, err := h.repo.ToggleLike(fn.ID, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to toggle like")
		http.Error(w, "Failed to process like", http.StatusInternalServerError)
		return
	}

	// Update function's like count (store in popularity_score for quick access)
	if _, err := h.repo.UpdateRegistryFunction(fn.ID, map[string]interface{}{
		"popularity_score": int(likeCount),
	}); err != nil {
		logrus.WithError(err).Warn("Failed to update function like count")
		// Don't fail the request
	}

	response := map[string]interface{}{
		"success":     true,
		"liked":       liked,
		"like_count":  likeCount,
		"function_id": fn.ID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunctionLikes returns the like count and users who liked a function
func (h *Handler) HandleGetFunctionLikes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Get actual like count from database
	likeCount, err := h.repo.CountLikesForFunction(fn.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to count likes")
		likeCount = 0
	}

	// Check if current user has liked (if authenticated)
	likedByUser := false
	if user := middleware.GetUserFromContext(r); user != nil {
		liked, err := h.repo.HasUserLiked(fn.ID, user.UserID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to check if user liked")
		}
		likedByUser = liked
	}

	response := map[string]interface{}{
		"function_id":   fn.ID.String(),
		"author":        author,
		"name":          name,
		"like_count":    likeCount,
		"liked_by_user": likedByUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
