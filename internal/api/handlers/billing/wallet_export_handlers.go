package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WalletExportHandler handles wallet data export API endpoints
type WalletExportHandler struct {
	walletService *wallet.Service
	logger        *logrus.Logger
}

// NewWalletExportHandler creates a new wallet export handler
func NewWalletExportHandler(walletService *wallet.Service) *WalletExportHandler {
	return &WalletExportHandler{
		walletService: walletService,
		logger:        logrus.New(),
	}
}

// ExportWalletsRequest represents the request for wallet export
//
// swagger:model ExportWalletsRequest
type ExportWalletsRequest struct {
	// Format can be "csv", "json", or "parquet"
	Format string `json:"format"`

	// IncludeTransactions includes the full transaction history
	IncludeTransactions bool `json:"include_transactions"`

	// IncludeBalanceHistory includes the balance history time-series
	IncludeBalanceHistory bool `json:"include_balance_history"`

	// Date range filter (for transactions and balance history)
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`

	// User filters
	UserIDs    []uuid.UUID `json:"user_ids,omitempty"`
	WalletIDs  []uuid.UUID `json:"wallet_ids,omitempty"`
	AgentIDs   []uuid.UUID `json:"agent_ids,omitempty"`

	// Status filters
	SuspendedOnly  bool `json:"suspended_only,omitempty"`
	ActiveOnly     bool `json:"active_only,omitempty"`
	LowBalanceOnly bool `json:"low_balance_only,omitempty"`

	// Fields to include
	Fields []string `json:"fields,omitempty"`
}

// ExportWallets exports wallet data for the specified users/agents
//
// POST /api/v1/wallets/export
//
// Request body:
//
//	{
//	  "format": "csv",
//	  "include_transactions": true,
//	  "include_balance_history": false,
//	  "date_from": "2024-01-01T00:00:00Z",
//	  "date_to": "2024-12-31T23:59:59Z",
//	  "agent_ids": ["uuid1", "uuid2"],
//	  "fields": ["wallet_id", "balance_usd", "suspended", "created_at"]
//	}
func (h *WalletExportHandler) ExportWallets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.extractUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req ExportWalletsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	// Default format to CSV if not specified
	if req.Format == "" {
		req.Format = "csv"
	}

	// Validate format
	if req.Format != "csv" && req.Format != "json" {
		h.writeError(w, http.StatusBadRequest, "Invalid Format", "Format must be 'csv' or 'json'")
		return
	}

	// Load wallet data
	wallets, err := h.loadWalletData(ctx, &req, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load wallet data for export")
		h.writeError(w, http.StatusInternalServerError, "Export Failed", "Failed to load wallet data")
		return
	}

	if len(wallets) == 0 {
		h.writeError(w, http.StatusNotFound, "No Data", "No wallet data found for the specified criteria")
		return
	}

	// Export based on format
	switch req.Format {
	case "csv":
		h.exportAsCSV(w, wallets, &req)
	case "json":
		h.exportAsJSON(w, wallets, &req)
	default:
		h.writeError(w, http.StatusBadRequest, "Invalid Format", "Unsupported export format")
	}
}

// ExportOwnWallet exports the authenticated user's own wallet data
//
// GET /api/v1/wallets/export/own
//
// Query parameters:
//   - format: csv or json (default: csv)
//   - include_transactions: true/false (default: true)
//   - include_balance_history: true/false (default: false)
//   - date_from: RFC3339 timestamp
//   - date_to: RFC3339 timestamp
func (h *WalletExportHandler) ExportOwnWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.extractUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	// Parse query parameters
	req := ExportWalletsRequest{
		Format:                r.URL.Query().Get("format"),
		IncludeTransactions:   r.URL.Query().Get("include_transactions") == "true",
		IncludeBalanceHistory: r.URL.Query().Get("include_balance_history") == "true",
		UserIDs:               []uuid.UUID{userID},
	}

	// Default format
	if req.Format == "" {
		req.Format = "csv"
	}

	// Parse dates if provided
	if from := r.URL.Query().Get("date_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			req.DateFrom = &t
		}
	}
	if to := r.URL.Query().Get("date_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			req.DateTo = &t
		}
	}

	// Load wallet data
	wallets, err := h.loadWalletData(ctx, &req, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load wallet data for export")
		h.writeError(w, http.StatusInternalServerError, "Export Failed", "Failed to load wallet data")
		return
	}

	if len(wallets) == 0 {
		h.writeError(w, http.StatusNotFound, "No Data", "No wallet data found")
		return
	}

	// Export based on format
	switch req.Format {
	case "csv":
		h.exportAsCSV(w, wallets, &req)
	case "json":
		h.exportAsJSON(w, wallets, &req)
	default:
		h.writeError(w, http.StatusBadRequest, "Invalid Format", "Unsupported export format")
	}
}

// ExportWalletByID exports a specific wallet by ID (admin only)
//
// GET /api/v1/admin/wallets/{walletId}/export
//
// Query parameters:
//   - format: csv or json (default: csv)
//   - include_transactions: true/false
//   - include_balance_history: true/false
func (h *WalletExportHandler) ExportWalletByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	walletIDStr := vars["walletId"]
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid ID", "Invalid wallet ID")
		return
	}

	// Parse query parameters
	req := ExportWalletsRequest{
		Format:                r.URL.Query().Get("format"),
		IncludeTransactions:   r.URL.Query().Get("include_transactions") == "true",
		IncludeBalanceHistory: r.URL.Query().Get("include_balance_history") == "true",
		WalletIDs:             []uuid.UUID{walletID},
	}

	// Default format
	if req.Format == "" {
		req.Format = "csv"
	}

	// Parse dates if provided
	if from := r.URL.Query().Get("date_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			req.DateFrom = &t
		}
	}
	if to := r.URL.Query().Get("date_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			req.DateTo = &t
		}
	}

	// Load wallet data
	wallets, err := h.loadWalletData(ctx, &req, uuid.Nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load wallet data for export")
		h.writeError(w, http.StatusInternalServerError, "Export Failed", "Failed to load wallet data")
		return
	}

	if len(wallets) == 0 {
		h.writeError(w, http.StatusNotFound, "Not Found", "Wallet not found")
		return
	}

	// Export based on format
	switch req.Format {
	case "csv":
		h.exportAsCSV(w, wallets, &req)
	case "json":
		h.exportAsJSON(w, wallets, &req)
	default:
		h.writeError(w, http.StatusBadRequest, "Invalid Format", "Unsupported export format")
	}
}

// WalletExportData represents a single wallet's data for export
type WalletExportData struct {
	WalletID              uuid.UUID                  `json:"wallet_id"`
	UserID                *uuid.UUID                 `json:"user_id,omitempty"`
	AgentID               *uuid.UUID                 `json:"agent_id,omitempty"`
	BalanceUSD            float64                    `json:"balance_usd"`
	BalanceLocal          *float64                   `json:"balance_local,omitempty"`
	Currency              string                     `json:"currency"`
	ExchangeRateToUSD     *float64                   `json:"exchange_rate_to_usd,omitempty"`
	LifetimeEarningsUSD   float64                    `json:"lifetime_earnings_usd"`
	LifetimeSpentUSD      float64                    `json:"lifetime_spent_usd"`
	SpendCapDailyUSD      *float64                   `json:"spend_cap_daily_usd,omitempty"`
	SpendCapMonthlyUSD    *float64                   `json:"spend_cap_monthly_usd,omitempty"`
	Suspended             bool                       `json:"suspended"`
	SuspendedAt           *time.Time                 `json:"suspended_at,omitempty"`
	SuspensionReason      *string                    `json:"suspension_reason,omitempty"`
	AutoTopupEnabled      bool                       `json:"auto_topup_enabled"`
	AutoTopupThresholdUSD float64                    `json:"auto_topup_threshold_usd"`
	CreatedAt             time.Time                  `json:"created_at"`
	UpdatedAt             time.Time                  `json:"updated_at"`

	// Related data
	Transactions    []wallet.WalletTransaction `json:"transactions,omitempty"`
	BalanceHistory  []wallet.BalanceHistoryEntry `json:"balance_history,omitempty"`
}

func (h *WalletExportHandler) loadWalletData(ctx context.Context, req *ExportWalletsRequest, requestingUserID uuid.UUID) ([]WalletExportData, error) {
	var wallets []WalletExportData

	// Build filter
	filter := wallet.WalletFilter{}
	if len(req.UserIDs) > 0 {
		filter.OwnerIDs = make([]string, len(req.UserIDs))
		for i, id := range req.UserIDs {
			filter.OwnerIDs[i] = id.String()
		}
	}
	if len(req.WalletIDs) > 0 {
		filter.WalletIDs = make([]string, len(req.WalletIDs))
		for i, id := range req.WalletIDs {
			filter.WalletIDs[i] = id.String()
		}
	}
	if req.SuspendedOnly {
		filter.Status = wallet.WalletStatusSuspended
	}
	if req.ActiveOnly {
		filter.Status = wallet.WalletStatusActive
	}

	// Query wallets
	result, err := h.walletService.ListWallets(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	for _, w := range result.Wallets {
		// Check low balance filter
		if req.LowBalanceOnly && w.BalanceUSD > 5.0 { // Assuming $5 is low balance threshold
			continue
		}

		data := WalletExportData{
			WalletID:              w.ID,
			UserID:                w.UserID,
			BalanceUSD:            w.BalanceUSD,
			BalanceLocal:          w.BalanceLocal,
			Currency:              w.Currency,
			ExchangeRateToUSD:     w.ExchangeRateToUSD,
			LifetimeEarningsUSD:   w.LifetimeEarningsUSD,
			LifetimeSpentUSD:      w.LifetimeSpentUSD,
			SpendCapDailyUSD:      w.SpendCapDailyUSD,
			SpendCapMonthlyUSD:    w.SpendCapMonthlyUSD,
			Suspended:             w.Suspended,
			SuspendedAt:           w.SuspendedAt,
			SuspensionReason:      w.SuspensionReason,
			AutoTopupEnabled:      w.AutoTopupEnabled,
			AutoTopupThresholdUSD: w.AutoTopupThresholdUSD,
			CreatedAt:             w.CreatedAt,
			UpdatedAt:             w.UpdatedAt,
		}

		// Load transactions if requested
		if req.IncludeTransactions {
			transactions, err := h.loadWalletTransactions(ctx, w.ID, req.DateFrom, req.DateTo)
			if err != nil {
				h.logger.WithError(err).WithField("wallet_id", w.ID).Warn("Failed to load transactions")
			} else {
				data.Transactions = transactions
			}
		}

		// Load balance history if requested
		if req.IncludeBalanceHistory {
			history, err := h.loadBalanceHistory(ctx, w.ID, req.DateFrom, req.DateTo)
			if err != nil {
				h.logger.WithError(err).WithField("wallet_id", w.ID).Warn("Failed to load balance history")
			} else {
				data.BalanceHistory = history
			}
		}

		wallets = append(wallets, data)
	}

	return wallets, nil
}

func (h *WalletExportHandler) loadWalletTransactions(ctx context.Context, walletID uuid.UUID, dateFrom, dateTo *time.Time) ([]wallet.WalletTransaction, error) {
	// Use the service method to get transaction history
	transactions, _, err := h.walletService.GetTransactionHistory(ctx, walletID, 1000, 0)
	return transactions, err
}

func (h *WalletExportHandler) loadBalanceHistory(ctx context.Context, walletID uuid.UUID, dateFrom, dateTo *time.Time) ([]wallet.BalanceHistoryEntry, error) {
	query := wallet.BalanceHistoryQuery{
		WalletID:  &walletID,
		StartDate: dateFrom,
		EndDate:   dateTo,
		Limit:     1000,
	}

	result, err := h.walletService.GetBalanceHistory(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Entries, nil
}

func (h *WalletExportHandler) exportAsCSV(w http.ResponseWriter, wallets []WalletExportData, req *ExportWalletsRequest) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Build header
	header := []string{
		"wallet_id", "user_id", "balance_usd", "balance_local", "currency",
		"exchange_rate_to_usd", "lifetime_earnings_usd", "lifetime_spent_usd",
		"spend_cap_daily_usd", "spend_cap_monthly_usd", "suspended",
		"suspension_reason", "auto_topup_enabled", "auto_topup_threshold_usd",
		"created_at", "updated_at",
	}

	if err := writer.Write(header); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Export Failed", "Failed to write CSV header")
		return
	}

	// Write wallet rows
	for _, data := range wallets {
		// Format UserID (pointer)
		userIDStr := ""
		if data.UserID != nil {
			userIDStr = data.UserID.String()
		}
		// Format BalanceLocal
		balanceLocalStr := ""
		if data.BalanceLocal != nil {
			balanceLocalStr = strconv.FormatFloat(*data.BalanceLocal, 'f', 4, 64)
		}
		// Format ExchangeRateToUSD
		rateStr := ""
		if data.ExchangeRateToUSD != nil {
			rateStr = strconv.FormatFloat(*data.ExchangeRateToUSD, 'f', 6, 64)
		}
		// Format SpendCapDailyUSD
		dailyCapStr := ""
		if data.SpendCapDailyUSD != nil {
			dailyCapStr = strconv.FormatFloat(*data.SpendCapDailyUSD, 'f', 4, 64)
		}
		// Format SpendCapMonthlyUSD
		monthlyCapStr := ""
		if data.SpendCapMonthlyUSD != nil {
			monthlyCapStr = strconv.FormatFloat(*data.SpendCapMonthlyUSD, 'f', 4, 64)
		}
		// Format SuspensionReason
		suspReasonStr := ""
		if data.SuspensionReason != nil {
			suspReasonStr = *data.SuspensionReason
		}

		row := []string{
			data.WalletID.String(),
			userIDStr,
			strconv.FormatFloat(data.BalanceUSD, 'f', 4, 64),
			balanceLocalStr,
			data.Currency,
			rateStr,
			strconv.FormatFloat(data.LifetimeEarningsUSD, 'f', 4, 64),
			strconv.FormatFloat(data.LifetimeSpentUSD, 'f', 4, 64),
			dailyCapStr,
			monthlyCapStr,
			strconv.FormatBool(data.Suspended),
			suspReasonStr,
			strconv.FormatBool(data.AutoTopupEnabled),
			strconv.FormatFloat(data.AutoTopupThresholdUSD, 'f', 4, 64),
			data.CreatedAt.Format(time.RFC3339),
			data.UpdatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			h.logger.WithError(err).Error("Failed to write CSV row")
			continue
		}

		// Write transaction rows if included
		if req.IncludeTransactions && len(data.Transactions) > 0 {
			// Write transaction header
			if err := writer.Write([]string{
				"", "", "transaction_id", "tx_type", "amount_usd", "amount_local",
				"description", "reference", "status", "created_at",
			}); err != nil {
				h.logger.WithError(err).Error("Failed to write transaction header")
			}

			for _, tx := range data.Transactions {
				ref := ""
				if tx.Reference != nil {
					ref = *tx.Reference
				}
				txRow := []string{
					"", // wallet_id spacer
					"", // user_id spacer
					tx.ID.String(),
					tx.TransactionType,
					strconv.FormatFloat(tx.AmountUSD, 'f', 4, 64),
					"", // AmountLocal not available
					"", // Description not available
					ref,
					tx.Status,
					tx.CreatedAt.Format(time.RFC3339),
				}
				if err := writer.Write(txRow); err != nil {
					h.logger.WithError(err).Error("Failed to write transaction row")
					continue
				}
			}
		}

		// Write balance history rows if included
		if req.IncludeBalanceHistory && len(data.BalanceHistory) > 0 {
			if err := writer.Write([]string{
				"", "", "history_id", "balance_usd", "change_amount_usd",
				"snapshot_type", "recorded_at",
			}); err != nil {
				h.logger.WithError(err).Error("Failed to write balance history header")
			}

		for _, entry := range data.BalanceHistory {
			histRow := []string{
				"", // wallet_id spacer
				"", // user_id spacer
				entry.ID.String(),
				strconv.FormatFloat(entry.BalanceUSD, 'f', 4, 64),
				strconv.FormatFloat(entry.ChangeAmountUSD, 'f', 4, 64),
				string(entry.SnapshotType),
				entry.RecordedAt.Format(time.RFC3339),
			}
			if err := writer.Write(histRow); err != nil {
				h.logger.WithError(err).Error("Failed to write balance history row")
				continue
			}
		}
		}
	}

	writer.Flush()

	// Set headers and write response
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=wallets_%s.csv", time.Now().Format("20060102_150405")))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *WalletExportHandler) exportAsJSON(w http.ResponseWriter, wallets []WalletExportData, req *ExportWalletsRequest) {
	response := map[string]interface{}{
		"exported_at":   time.Now().Format(time.RFC3339),
		"format":        "json",
		"count":         len(wallets),
		"filters":       req,
		"wallets":       wallets,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=wallets_%s.json", time.Now().Format("20060102_150405")))
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON export")
	}
}

func (h *WalletExportHandler) extractUserID(r *http.Request) (uuid.UUID, error) {
	// Try to get from context (set by auth middleware)
	if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
		return uuid.Parse(userID)
	}
	if userID, ok := r.Context().Value("userID").(uuid.UUID); ok {
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("user ID not found in request context")
}

func (h *WalletExportHandler) writeError(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"status": status,
			"title":  title,
			"detail": detail,
		},
	}

	json.NewEncoder(w).Encode(response)
}
