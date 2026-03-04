package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var dayLabels = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
var monthLabels = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// HandleDashboardActivity returns platform activity for the last 7 days (new users, function calls).
// GET /v1/admin/dashboard/activity?days=7
func (h *Handler) HandleDashboardActivity(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 90 {
			days = n
		}
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	filters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}
	events, err := h.repo.ListAuditEventsFiltered(2000, 0, filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to list audit events for dashboard activity")
		http.Error(w, "Failed to get activity", http.StatusInternalServerError)
		return
	}

	byDay := make(map[string]struct{ NewUsers int; FunctionCalls int })
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		key := d.UTC().Format("2006-01-02")
		byDay[key] = struct{ NewUsers int; FunctionCalls int }{}
	}

	for _, e := range events {
		key := e.Timestamp.UTC().Format("2006-01-02")
		entry, ok := byDay[key]
		if !ok {
			continue
		}
		rt := e.ResourceType
		action := e.Action
		isUser := rt == "user" || strings.Contains(action, "user") || strings.Contains(action, "signup") || strings.Contains(action, "login")
		isFunction := rt == "function" || rt == "app" || strings.Contains(action, "function") || strings.Contains(action, "deploy")
		if isUser {
			entry.NewUsers++
		}
		if isFunction {
			entry.FunctionCalls++
		}
		if !isUser && !isFunction {
			entry.FunctionCalls++
		}
		byDay[key] = entry
	}

	type point struct {
		Date          string `json:"date"`
		DayLabel      string `json:"day_label"`
		NewUsers      int    `json:"new_users"`
		FunctionCalls int    `json:"function_calls"`
	}
	series := make([]point, 0, days)
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		key := d.UTC().Format("2006-01-02")
		entry := byDay[key]
		series = append(series, point{
			Date:          key,
			DayLabel:      dayLabels[d.UTC().Weekday()],
			NewUsers:      entry.NewUsers,
			FunctionCalls: entry.FunctionCalls,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"series": series})
}

// HandleDashboardRevenue returns revenue by month and MRR for the admin dashboard.
// GET /v1/admin/dashboard/revenue?months=7
func (h *Handler) HandleDashboardRevenue(w http.ResponseWriter, r *http.Request) {
	months := 7
	if m := r.URL.Query().Get("months"); m != "" {
		if n, err := strconv.Atoi(m); err == nil && n > 0 && n <= 24 {
			months = n
		}
	}

	invoices, err := h.repo.ListAllInvoices(5000, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list invoices for dashboard revenue")
		http.Error(w, "Failed to get revenue", http.StatusInternalServerError)
		return
	}

	byMonth := make(map[string]int) // "2006-01" -> revenue_cents
	now := time.Now()
	for i := 0; i < months; i++ {
		t := now.AddDate(0, -months+1+i, 0)
		byMonth[t.UTC().Format("2006-01")] = 0
	}

	for _, inv := range invoices {
		var t time.Time
		if inv.PaidAt != nil {
			t = *inv.PaidAt
		} else {
			t = inv.CreatedAt
		}
		key := t.UTC().Format("2006-01")
		if _, ok := byMonth[key]; ok {
			byMonth[key] += inv.AmountPaidCents
		}
	}

	type point struct {
		Month        string `json:"month"`
		RevenueCents int    `json:"revenue_cents"`
	}
	series := make([]point, 0, months)
	for i := 0; i < months; i++ {
		t := now.AddDate(0, -months+1+i, 0)
		key := t.UTC().Format("2006-01")
		series = append(series, point{
			Month:        monthLabels[t.Month()-1],
			RevenueCents: byMonth[key],
		})
	}

	mrrCents := 0
	if len(series) > 0 {
		mrrCents = series[len(series)-1].RevenueCents
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"series":     series,
		"mrr_cents":  mrrCents,
	})
}

// HandleDashboardQuickStats returns quick stats for the admin dashboard footer.
// GET /v1/admin/dashboard/quick-stats
func (h *Handler) HandleDashboardQuickStats(w http.ResponseWriter, r *http.Request) {
	// Placeholder values; can be wired to Prometheus/metrics later
	stats := map[string]interface{}{
		"platform_uptime_percent": 99.97,
		"avg_response_time_ms":    45,
		"functions_executed":      "0",
		"data_processed":           "0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

