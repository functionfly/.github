package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/sirupsen/logrus"
)

const (
	nominatimURL     = "https://nominatim.openstreetmap.org/search"
	nominatimUA      = "FunctionFly/1.0 (https://functionfly.io; admin@functionfly.io)"
	nominatimTimeout = 5 * time.Second
	maxResults       = 5
	minQueryLen      = 2
	maxQueryLen      = 200
)

// locationRateLimiter enforces Nominatim's 1 req/sec policy globally.
type locationRateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
}

var locationLimiter = &locationRateLimiter{}

func (l *locationRateLimiter) Wait() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.lastCall.IsZero() {
		elapsed := time.Since(l.lastCall)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
	}
	l.lastCall = time.Now()
}

type nominatimAddress struct {
	City        string `json:"city,omitempty"`
	Town        string `json:"town,omitempty"`
	Village     string `json:"village,omitempty"`
	Municipality string `json:"municipality,omitempty"`
	State       string `json:"state,omitempty"`
	StateDistrict string `json:"state_district,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type nominatimResult struct {
	PlaceID     int              `json:"place_id"`
	DisplayName string           `json:"display_name"`
	Address     *nominatimAddress `json:"address,omitempty"`
}

type locationResult struct {
	PlaceID     int    `json:"placeId"`
	Label       string `json:"label"`
	DisplayName string `json:"displayName"`
}

// HandleSearchLocations proxies GET /v1/locations/search?q=... to Nominatim
// with proper User-Agent, rate limiting, and request timeout.
func (h *Handler) HandleSearchLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < minQueryLen {
		writeJSON(w, http.StatusOK, map[string]interface{}{"locations": []interface{}{}})
		return
	}
	if len(q) > maxQueryLen {
		q = q[:maxQueryLen]
	}

	// Enforce Nominatim's 1 req/sec globally
	locationLimiter.Wait()

	params := url.Values{
		"q":              {q},
		"format":         {"json"},
		"limit":          {fmt.Sprintf("%d", maxResults)},
		"addressdetails": {"1"},
		"accept-language": {"en"},
	}

	reqURL := fmt.Sprintf("%s?%s", nominatimURL, params.Encode())

	client := &http.Client{Timeout: nominatimTimeout}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, reqURL, nil)
	if err != nil {
		logrus.WithError(err).Error("locations: failed to build nominatim request")
		apierror.WriteError(w, apierror.NewInternal("Location search failed"))
		return
	}
	req.Header.Set("User-Agent", nominatimUA)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("locations: nominatim request failed")
		writeJSON(w, http.StatusOK, map[string]interface{}{"locations": []interface{}{}})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status", resp.StatusCode).Warn("locations: nominatim returned non-200")
		writeJSON(w, http.StatusOK, map[string]interface{}{"locations": []interface{}{}})
		return
	}

	var nominatimResults []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&nominatimResults); err != nil {
		logrus.WithError(err).Warn("locations: failed to decode nominatim response")
		writeJSON(w, http.StatusOK, map[string]interface{}{"locations": []interface{}{}})
		return
	}

	out := make([]locationResult, 0, len(nominatimResults))
	for _, nr := range nominatimResults {
		out = append(out, locationResult{
			PlaceID:     nr.PlaceID,
			Label:       formatNominatimLabel(nr),
			DisplayName: nr.DisplayName,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"locations": out})
}

func formatNominatimLabel(r nominatimResult) string {
	if r.Address == nil {
		return r.DisplayName
	}
	a := r.Address
	city := a.City
	if city == "" {
		city = a.Town
	}
	if city == "" {
		city = a.Village
	}
	if city == "" {
		city = a.Municipality
	}
	state := a.State
	if state == "" {
		state = a.StateDistrict
	}
	parts := []string{}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return r.DisplayName
}
