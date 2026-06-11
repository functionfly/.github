package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
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

	type point struct {
		Date          string `json:"date"`
		DayLabel      string `json:"day_label"`
		NewUsers      int    `json:"new_users"`
		FunctionCalls int    `json:"function_calls"`
	}

	filters := map[string]interface{}{
		"start_time": start,
		"end_time":   end,
	}
	events, err := h.repo.ListAuditEventsFiltered(2000, 0, filters)
	if err != nil {
		logrus.WithError(err).Warn("Failed to list audit events for dashboard activity; returning empty series")
		// Return empty data so the dashboard still loads (e.g. if audit_events table is missing or DB error)
		series := make([]point, 0, days)
		for i := 0; i < days; i++ {
			d := start.AddDate(0, 0, i)
			key := d.UTC().Format("2006-01-02")
			series = append(series, point{
				Date:          key,
				DayLabel:      dayLabels[d.UTC().Weekday()],
				NewUsers:      0,
				FunctionCalls: 0,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      []interface{}{},
			"series":    series,
			"success":   true,
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	byDay := make(map[string]struct {
		NewUsers      int
		FunctionCalls int
	})
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		key := d.UTC().Format("2006-01-02")
		byDay[key] = struct {
			NewUsers      int
			FunctionCalls int
		}{}
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

	// Recent activity list for the dashboard feed (frontend expects data array)
	recentLimit := 50
	if len(events) < recentLimit {
		recentLimit = len(events)
	}
	recentData := make([]map[string]interface{}, 0, recentLimit)
	for i := 0; i < recentLimit; i++ {
		e := events[i]
		recentData = append(recentData, map[string]interface{}{
			"timestamp":     e.Timestamp.Format(time.RFC3339),
			"action":        e.Action,
			"user_email":    e.ActorEmail,
			"resource_type": e.ResourceType,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      recentData,
		"series":    series,
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	})
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
		apierror.WriteError(w, apierror.NewInternal("Failed to get revenue"))
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
		"series":    series,
		"mrr_cents": mrrCents,
	})
}

// HandleDashboardQuickStats returns quick stats for the admin dashboard (real counts from DB).
// GET /v1/admin/dashboard/quick-stats
func (h *Handler) HandleDashboardQuickStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	stats := map[string]interface{}{
		"total_tenants":     0,
		"total_users":       0,
		"active_sessions":   0,
		"pending_incidents": 0,
	}

	if tenants, err := h.repo.ListTenants(); err == nil {
		stats["total_tenants"] = len(tenants)
	}
	if users, err := h.repo.ListUsers(); err == nil {
		stats["total_users"] = len(users)
	}

	if incidents, err := h.repo.ListIncidents(ctx, 1000, 0, nil); err == nil {
		pending := 0
		for _, inc := range incidents {
			if inc.Status != "resolved" {
				pending++
			}
		}
		stats["pending_incidents"] = pending
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      stats,
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
