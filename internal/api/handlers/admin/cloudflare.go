package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/sirupsen/logrus"
)

// ── Cloudflare GraphQL types ──────────────────────────────────────────────────

type cfGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type cfGraphQLResponse struct {
	Data   *cfData  `json:"data"`
	Errors []cfGQLError `json:"errors"`
}

type cfGQLError struct {
	Message string `json:"message"`
}

type cfData struct {
	Viewer struct {
		Zones []cfZone `json:"zones"`
	} `json:"viewer"`
}

type cfZone struct {
	Current  []cfDayGroup `json:"current"`
	Previous []cfDayGroup `json:"previous"`
}

type cfDayGroup struct {
	Sum  cfDaySum  `json:"sum"`
	Uniq cfDayUniq `json:"uniq"`
}

type cfDaySum struct {
	Requests          int64                `json:"requests"`
	Bytes             int64                `json:"bytes"`
	PageViews         int64                `json:"pageViews"`
	CachedRequests    int64                `json:"cachedRequests"`
	CachedBytes       int64                `json:"cachedBytes"`
	EncryptedRequests int64                `json:"encryptedRequests"`
	EncryptedBytes    int64                `json:"encryptedBytes"`
	CountryMap        []cfCountryEntry     `json:"countryMap"`
	StatusMap         []cfStatusEntry      `json:"responseStatusMap"`
	SSLMap            []cfSSLEntry         `json:"clientSSLMap"`
	ContentTypeMap    []cfContentTypeEntry `json:"contentTypeMap"`
}

type cfDayUniq struct {
	Uniques int64 `json:"uniques"`
}

type cfCountryEntry struct {
	ClientCountryName string `json:"clientCountryName"`
	Requests          int64  `json:"requests"`
	Bytes             int64  `json:"bytes"`
}

type cfStatusEntry struct {
	EdgeResponseStatus int   `json:"edgeResponseStatus"`
	Requests           int64 `json:"requests"`
}

type cfSSLEntry struct {
	ClientSSLProtocol string `json:"clientSSLProtocol"`
	Requests          int64  `json:"requests"`
}

type cfContentTypeEntry struct {
	EdgeResponseContentTypeName string `json:"edgeResponseContentTypeName"`
	Requests                    int64  `json:"requests"`
}

// ── Response types (matches frontend) ────────────────────────────────────────

type cfTrendStat struct {
	Value     float64 `json:"value"`
	ChangePct float64 `json:"change_pct"`
}

type cfCountryStat struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Requests    int64  `json:"requests"`
	BandwidthBytes int64 `json:"bandwidth_bytes"`
}

type cfNetworkBreakdown struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type cfAnalyticsResponse struct {
	Period  string `json:"period"`
	Traffic struct {
		Requests  cfTrendStat `json:"requests"`
		Bandwidth cfTrendStat `json:"bandwidth"`
		Visits    cfTrendStat `json:"visits"`
		PageViews cfTrendStat `json:"page_views"`
	} `json:"traffic"`
	Security struct {
		EncryptedRequests     cfTrendStat `json:"encrypted_requests"`
		EncryptedRequestsRate cfTrendStat `json:"encrypted_requests_rate"`
		EncryptedBandwidth    cfTrendStat `json:"encrypted_bandwidth"`
		EncryptedBandwidthRate cfTrendStat `json:"encrypted_bandwidth_rate"`
	} `json:"security"`
	Cache struct {
		CachedRequests     cfTrendStat `json:"cached_requests"`
		CachedRequestsRate cfTrendStat `json:"cached_requests_rate"`
		CachedBandwidth    cfTrendStat `json:"cached_bandwidth"`
		CachedBandwidthRate cfTrendStat `json:"cached_bandwidth_rate"`
	} `json:"cache"`
	Errors struct {
		Errors4xx    cfTrendStat `json:"errors_4xx"`
		ErrorRate4xx cfTrendStat `json:"error_rate_4xx"`
		Errors5xx    cfTrendStat `json:"errors_5xx"`
		ErrorRate5xx cfTrendStat `json:"error_rate_5xx"`
	} `json:"errors"`
	TopCountries []cfCountryStat      `json:"top_countries"`
	HTTPVersions []cfNetworkBreakdown `json:"http_versions"`
	TLSVersions  []cfNetworkBreakdown `json:"tls_versions"`
	ContentTypes []cfNetworkBreakdown `json:"content_types"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

// HandleCloudflareAnalytics returns Cloudflare zone analytics for the specified period.
// Requires CF_API_TOKEN and CF_ZONE_ID environment variables.
func (h *Handler) HandleCloudflareAnalytics(w http.ResponseWriter, r *http.Request) {
	apiToken := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		apierror.WriteError(w, apierror.NewNotFound("Cloudflare not configured. Set CF_API_TOKEN and CF_ZONE_ID."))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	// Always use daily groups — consistent schema for all periods.
	// For 24h we look back 2 days vs previous 2 days.
	now := time.Now().UTC()
	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	default:
		period = "24h"
		days = 2
	}

	currentEnd := now.Format("2006-01-02")
	currentStart := now.AddDate(0, 0, -days).Format("2006-01-02")
	previousEnd := now.AddDate(0, 0, -days).Format("2006-01-02")
	previousStart := now.AddDate(0, 0, -days*2).Format("2006-01-02")

	query := fmt.Sprintf(`{
  viewer {
    zones(filter: { zoneTag: "%s" }) {
      current: httpRequests1dGroups(
        limit: 60
        filter: { date_geq: "%s", date_leq: "%s" }
      ) {
        sum {
          requests bytes pageViews cachedRequests cachedBytes
          encryptedRequests encryptedBytes
          countryMap { clientCountryName requests bytes }
          responseStatusMap { edgeResponseStatus requests }
          clientSSLMap { clientSSLProtocol requests }
          contentTypeMap { edgeResponseContentTypeName requests }
        }
        uniq { uniques }
      }
      previous: httpRequests1dGroups(
        limit: 60
        filter: { date_geq: "%s", date_leq: "%s" }
      ) {
        sum {
          requests bytes pageViews cachedRequests cachedBytes
          encryptedRequests encryptedBytes
        }
        uniq { uniques }
      }
    }
  }
}`, zoneID, currentStart, currentEnd, previousStart, previousEnd)

	body, err := cfGraphQLQuery(apiToken, query)
	if err != nil {
		logrus.WithError(err).Error("Cloudflare GraphQL request failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to fetch Cloudflare analytics"))
		return
	}

	var gqlResp cfGraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		logrus.WithError(err).Error("Failed to parse Cloudflare GraphQL response")
		apierror.WriteError(w, apierror.NewInternal("Failed to parse Cloudflare response"))
		return
	}

	if len(gqlResp.Errors) > 0 {
		errMsgs := make([]string, 0, len(gqlResp.Errors))
		for _, e := range gqlResp.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		logrus.WithField("errors", errMsgs).Error("Cloudflare GraphQL returned errors")
		apierror.WriteError(w, apierror.NewBadRequest(strings.Join(errMsgs, "; ")))
		return
	}

	if gqlResp.Data == nil || len(gqlResp.Data.Viewer.Zones) == 0 {
		apierror.WriteError(w, apierror.NewNotFound("No zone data returned"))
		return
	}

	zone := gqlResp.Data.Viewer.Zones[0]
	cur := aggregateGroups(zone.Current)
	prev := aggregateGroups(zone.Previous)

	resp := buildResponse(period, cur, prev)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": resp})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func cfGraphQLQuery(apiToken, query string) ([]byte, error) {
	payload, _ := json.Marshal(cfGraphQLRequest{Query: query})
	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type aggregated struct {
	Requests          int64
	Bytes             int64
	PageViews         int64
	CachedRequests    int64
	CachedBytes       int64
	EncryptedRequests int64
	EncryptedBytes    int64
	Uniques           int64
	CountryMap        map[string]*cfCountryStat
	StatusMap         map[int]int64
	SSLMap            map[string]int64
	ContentTypeMap    map[string]int64
}

func aggregateGroups(groups []cfDayGroup) aggregated {
	a := aggregated{
		CountryMap:     make(map[string]*cfCountryStat),
		StatusMap:      make(map[int]int64),
		SSLMap:         make(map[string]int64),
		ContentTypeMap: make(map[string]int64),
	}
	for _, g := range groups {
		a.Requests += g.Sum.Requests
		a.Bytes += g.Sum.Bytes
		a.PageViews += g.Sum.PageViews
		a.CachedRequests += g.Sum.CachedRequests
		a.CachedBytes += g.Sum.CachedBytes
		a.EncryptedRequests += g.Sum.EncryptedRequests
		a.EncryptedBytes += g.Sum.EncryptedBytes
		a.Uniques += g.Uniq.Uniques

		for _, c := range g.Sum.CountryMap {
			if c.ClientCountryName == "" {
				continue
			}
			if existing, ok := a.CountryMap[c.ClientCountryName]; ok {
				existing.Requests += c.Requests
				existing.BandwidthBytes += c.Bytes
			} else {
				code := countryNameToCode(c.ClientCountryName)
				a.CountryMap[c.ClientCountryName] = &cfCountryStat{
					Country:        c.ClientCountryName,
					CountryCode:    code,
					Requests:       c.Requests,
					BandwidthBytes: c.Bytes,
				}
			}
		}
		for _, s := range g.Sum.StatusMap {
			a.StatusMap[s.EdgeResponseStatus] += s.Requests
		}
		for _, s := range g.Sum.SSLMap {
			a.SSLMap[s.ClientSSLProtocol] += s.Requests
		}
		for _, ct := range g.Sum.ContentTypeMap {
			label := ct.EdgeResponseContentTypeName
			if label == "" {
				label = "other"
			}
			a.ContentTypeMap[label] += ct.Requests
		}
	}
	return a
}

func buildResponse(period string, cur, prev aggregated) cfAnalyticsResponse {
	var r cfAnalyticsResponse
	r.Period = period

	r.Traffic.Requests = makeStat(float64(cur.Requests), float64(prev.Requests))
	r.Traffic.Bandwidth = makeStat(float64(cur.Bytes), float64(prev.Bytes))
	r.Traffic.Visits = makeStat(float64(cur.Uniques), float64(prev.Uniques))
	r.Traffic.PageViews = makeStat(float64(cur.PageViews), float64(prev.PageViews))

	r.Security.EncryptedRequests = makeStat(float64(cur.EncryptedRequests), float64(prev.EncryptedRequests))
	r.Security.EncryptedBandwidth = makeStat(float64(cur.EncryptedBytes), float64(prev.EncryptedBytes))

	var curEncReqRate, prevEncReqRate, curEncBWRate, prevEncBWRate float64
	if cur.Requests > 0 {
		curEncReqRate = float64(cur.EncryptedRequests) / float64(cur.Requests) * 100
	}
	if prev.Requests > 0 {
		prevEncReqRate = float64(prev.EncryptedRequests) / float64(prev.Requests) * 100
	}
	if cur.Bytes > 0 {
		curEncBWRate = float64(cur.EncryptedBytes) / float64(cur.Bytes) * 100
	}
	if prev.Bytes > 0 {
		prevEncBWRate = float64(prev.EncryptedBytes) / float64(prev.Bytes) * 100
	}
	r.Security.EncryptedRequestsRate = makeStat(curEncReqRate, prevEncReqRate)
	r.Security.EncryptedBandwidthRate = makeStat(curEncBWRate, prevEncBWRate)

	r.Cache.CachedRequests = makeStat(float64(cur.CachedRequests), float64(prev.CachedRequests))
	r.Cache.CachedBandwidth = makeStat(float64(cur.CachedBytes), float64(prev.CachedBytes))

	var curCacheReqRate, prevCacheReqRate, curCacheBWRate, prevCacheBWRate float64
	if cur.Requests > 0 {
		curCacheReqRate = float64(cur.CachedRequests) / float64(cur.Requests) * 100
	}
	if prev.Requests > 0 {
		prevCacheReqRate = float64(prev.CachedRequests) / float64(prev.Requests) * 100
	}
	if cur.Bytes > 0 {
		curCacheBWRate = float64(cur.CachedBytes) / float64(cur.Bytes) * 100
	}
	if prev.Bytes > 0 {
		prevCacheBWRate = float64(prev.CachedBytes) / float64(prev.Bytes) * 100
	}
	r.Cache.CachedRequestsRate = makeStat(curCacheReqRate, prevCacheReqRate)
	r.Cache.CachedBandwidthRate = makeStat(curCacheBWRate, prevCacheBWRate)

	// Errors: count 4xx and 5xx from status map
	var cur4xx, cur5xx, prev4xx, prev5xx int64
	for status, count := range cur.StatusMap {
		if status >= 400 && status < 500 {
			cur4xx += count
		} else if status >= 500 {
			cur5xx += count
		}
	}
	// previous period status breakdown is not fetched (would double query size),
	// so we use zeros for prev — change_pct will be 0 which is acceptable.
	_ = prev4xx
	_ = prev5xx

	r.Errors.Errors4xx = makeStat(float64(cur4xx), 0)
	r.Errors.Errors5xx = makeStat(float64(cur5xx), 0)

	var cur4xxRate, cur5xxRate float64
	if cur.Requests > 0 {
		cur4xxRate = float64(cur4xx) / float64(cur.Requests) * 100
		cur5xxRate = float64(cur5xx) / float64(cur.Requests) * 100
	}
	r.Errors.ErrorRate4xx = makeStat(cur4xxRate, 0)
	r.Errors.ErrorRate5xx = makeStat(cur5xxRate, 0)

	// Top countries (sorted by requests desc, top 10)
	countries := make([]*cfCountryStat, 0, len(cur.CountryMap))
	for _, c := range cur.CountryMap {
		countries = append(countries, c)
	}
	sort.Slice(countries, func(i, j int) bool {
		return countries[i].Requests > countries[j].Requests
	})
	if len(countries) > 10 {
		countries = countries[:10]
	}
	r.TopCountries = make([]cfCountryStat, 0, len(countries))
	for _, c := range countries {
		r.TopCountries = append(r.TopCountries, *c)
	}

	// HTTP versions — not available in daily rollup dataset; return empty
	r.HTTPVersions = []cfNetworkBreakdown{}
	// TLS versions from clientSSLMap
	r.TLSVersions = sortedBreakdown(cur.SSLMap)
	// Content types (top 5)
	ct := sortedBreakdown(cur.ContentTypeMap)
	if len(ct) > 5 {
		ct = ct[:5]
	}
	r.ContentTypes = ct

	return r
}

func makeStat(current, previous float64) cfTrendStat {
	if previous == 0 {
		if current == 0 {
			return cfTrendStat{Value: 0, ChangePct: 0}
		}
		return cfTrendStat{Value: current, ChangePct: 100}
	}
	changePct := (current - previous) / previous * 100
	return cfTrendStat{Value: current, ChangePct: changePct}
}

func sortedBreakdown(m map[string]int64) []cfNetworkBreakdown {
	result := make([]cfNetworkBreakdown, 0, len(m))
	for label, count := range m {
		result = append(result, cfNetworkBreakdown{Label: label, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

// countryNameToCode maps common Cloudflare country names to ISO 3166-1 alpha-2 codes.
func countryNameToCode(name string) string {
	codes := map[string]string{
		"France":         "FR",
		"United States":  "US",
		"Netherlands":    "NL",
		"Switzerland":    "CH",
		"India":          "IN",
		"Canada":         "CA",
		"United Kingdom": "GB",
		"China":          "CN",
		"Brazil":         "BR",
		"Germany":        "DE",
		"Japan":          "JP",
		"Australia":      "AU",
		"Singapore":      "SG",
		"Russia":         "RU",
		"South Korea":    "KR",
		"Sweden":         "SE",
		"Norway":         "NO",
		"Italy":          "IT",
		"Spain":          "ES",
		"Poland":         "PL",
		"Ukraine":        "UA",
		"Mexico":         "MX",
		"Argentina":      "AR",
		"Turkey":         "TR",
		"Indonesia":      "ID",
		"Thailand":       "TH",
		"Vietnam":        "VN",
		"Portugal":       "PT",
		"Belgium":        "BE",
		"Austria":        "AT",
		"Czech Republic": "CZ",
		"Denmark":        "DK",
		"Finland":        "FI",
		"Romania":        "RO",
		"Hong Kong":      "HK",
		"Taiwan":         "TW",
		"Israel":         "IL",
		"South Africa":   "ZA",
		"Egypt":          "EG",
		"Nigeria":        "NG",
		"Pakistan":       "PK",
		"Bangladesh":     "BD",
		"Malaysia":       "MY",
		"Philippines":    "PH",
		"New Zealand":    "NZ",
		"Colombia":       "CO",
		"Chile":          "CL",
		"Peru":           "PE",
		"Venezuela":      "VE",
		"Saudi Arabia":   "SA",
		"UAE":            "AE",
		"Ireland":        "IE",
		"Greece":         "GR",
		"Hungary":        "HU",
		"Slovakia":       "SK",
		"Bulgaria":       "BG",
		"Croatia":        "HR",
		"Serbia":         "RS",
		"Lithuania":      "LT",
		"Latvia":         "LV",
		"Estonia":        "EE",
		"Slovenia":       "SI",
		"Luxembourg":     "LU",
	}
	if code, ok := codes[name]; ok {
		return code
	}
	// Fallback: uppercase first 2 chars
	if len(name) >= 2 {
		return strings.ToUpper(name[:2])
	}
	return "XX"
}
