package cityranking

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler holds the dependencies for the city-ranking HTTP surface.
type Handler struct {
	repo    *cityranking.Repository
	cache   *cityranking.Cache
	ipgeo   *cityranking.IPGeoResolver
	geo     *cityranking.GeoResolver
	log     *logrus.Logger
}

// NewHandler wires a Handler with the repository and cache. cache may be nil
// (handler will skip caching). ipgeo may be nil; the IP-fallback handler will
// return 503 in that case. geo may be nil; city auto-creation will be skipped.
func NewHandler(repo *cityranking.Repository, cache *cityranking.Cache, ipgeo *cityranking.IPGeoResolver, geo *cityranking.GeoResolver, log *logrus.Logger) *Handler {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Handler{repo: repo, cache: cache, ipgeo: ipgeo, geo: geo, log: log}
}

// HandleListLeaderboard: GET /city-rankings
func (h *Handler) HandleListLeaderboard(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetLeaderboard(r.Context(), country, category); err == nil && hit {
		entries := ToEntries(cached)
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, LeaderboardResponse{
			PeriodEnd:   periodOrZero(cached),
			TotalRanked: len(cached),
			Entries:     entries,
			Country:     country,
			Category:    string(category),
			CacheHit:    true,
		})
		return
	}

	rows, err := h.repo.ListRankings(r.Context(), 500, country, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list rankings")
		http.Error(w, "failed to list rankings", http.StatusInternalServerError)
		return
	}
	// Best-effort cache warm.
	if len(rows) > 0 {
		_ = h.cache.SetLeaderboard(r.Context(), country, category, rows)
	}
	entries := ToEntries(rows)
	if len(entries) > limit {
		entries = entries[:limit]
	}
	periodEnd := time.Time{}
	if len(rows) > 0 {
		periodEnd = rows[0].PeriodEnd
	}
	writeJSON(w, LeaderboardResponse{
		PeriodEnd:   periodEnd,
		TotalRanked: len(rows),
		Entries:     entries,
		Country:     country,
		Category:    string(category),
	})
}

// HandleListCategories: GET /city-rankings/categories (public)
func (h *Handler) HandleListCategories(w http.ResponseWriter, r *http.Request) {
	categories := cityranking.AllCategories
	meta := make([]cityranking.CategoryMeta, 0, len(categories))
	for _, c := range categories {
		meta = append(meta, cityranking.CategoryMetaFor(c))
	}
	writeJSON(w, CategoriesResponse{Categories: meta})
}

// HandleListAmbassadors: GET /city-rankings/ambassadors (public)
//
// Lists the active ambassador for every metro that currently has one
// (i.e. passes the k=5 privacy threshold). Optional ?country=US narrows
// the list; ?limit caps it.
func (h *Handler) HandleListAmbassadors(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 200, 500)

	rows, err := h.repo.ListAmbassadors(r.Context(), country, limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to list ambassadors")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, AmbassadorsListResponse{
		Total:      len(rows),
		Entries:    ToAmbassadorEntries(rows),
		Country:    country,
		PrivacyMin: cityranking.MinActiveUsersForPublic,
	})
}

// HandleListCountries: GET /city-rankings/countries (public)
//
// Returns the list of countries that have active metros, for populating
// the country filter on the ambassadors page. The list is dynamic —
// new countries are automatically included when metros are added.
func (h *Handler) HandleListCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := h.repo.ListCountries(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to list countries")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	entries := make([]CountryEntry, 0, len(countries))
	for _, c := range countries {
		entries = append(entries, CountryEntry{Code: c.Code, Name: c.Name})
	}
	writeJSON(w, CountriesResponse{Countries: entries})
}

// HandleGetAmbassador: GET /city-rankings/{slug}/ambassador (public)
//
// Returns the active ambassador for a single metro, or 404 if the metro
// has no leaderboard row (and therefore no ambassador).
func (h *Handler) HandleGetAmbassador(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	slug = strings.TrimPrefix(slug, "/")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	metro, err := h.repo.GetMetroBySlug(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get metro for ambassador")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if metro == nil {
		http.Error(w, "metro not found", http.StatusNotFound)
		return
	}
	amb, err := h.repo.GetAmbassadorForMetro(r.Context(), metro.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get ambassador")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if amb == nil {
		http.Error(w, "no ambassador for this metro (privacy threshold not met)", http.StatusNotFound)
		return
	}
	writeJSON(w, AmbassadorResponse{
		MetroSlug:   slug,
		MetroName:   metro.Name,
		CountryCode: metro.CountryCode,
		Ambassador:  ToAmbassadorEntry(*amb),
	})
}

// HandleGetMetro: GET /city-rankings/{slug}
func (h *Handler) HandleGetMetro(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	if slug == "" {
		slug = strings.TrimSpace(r.URL.Query().Get("slug"))
	}
	slug = strings.TrimPrefix(slug, "/")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	category := parseCategory(r)

	if cached, history, hit, err := h.cache.GetMetro(r.Context(), slug, category); err == nil && hit && cached != nil {
		writeJSON(w, MetroDetailResponse{
			Current:   ptrEntry(ToEntry(*cached)),
			History:   ToEntries(history),
			PeriodEnd: cached.PeriodEnd,
			Category:  string(category),
			CacheHit:  true,
		})
		return
	}

	metro, err := h.repo.GetMetroBySlug(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get metro")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if metro == nil {
		http.Error(w, "metro not found", http.StatusNotFound)
		return
	}

	current, err := h.repo.GetRankingBySlug(r.Context(), slug, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to get ranking")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	history, err := h.repo.ListHistory(r.Context(), slug, 30, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list history")
	}

	notRanked := current == nil || current.ActiveUsers < cityranking.MinActiveUsersForPublic

	if current != nil {
		_ = h.cache.SetMetro(r.Context(), slug, category, current, history)
	}

	var currentEntry *CityRankingEntry
	if current != nil {
		ce := ToEntry(*current)
		currentEntry = &ce
	}
	writeJSON(w, MetroDetailResponse{
		Current:               currentEntry,
		History:               ToEntries(history),
		NotRanked:             notRanked,
		PrivacyMinActiveUsers: cityranking.MinActiveUsersForPublic,
		PeriodEnd:             currentPeriod(history, current),
		Category:              string(category),
	})
}

// HandleListMovers: GET /city-rankings/movers?direction=gainers|losers
func (h *Handler) HandleListMovers(w http.ResponseWriter, r *http.Request) {
	direction := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("direction")))
	if direction == "" {
		direction = "gainers"
	}
	if direction != "gainers" && direction != "losers" {
		http.Error(w, "direction must be gainers or losers", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r.URL.Query().Get("limit"), 25, 100)
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetMovers(r.Context(), direction, category); err == nil && hit {
		entries := ToEntries(cached)
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, MoversResponse{
			Direction: direction,
			PeriodEnd: periodOrZero(cached),
			Entries:   entries,
			Category:  string(category),
		})
		return
	}

	rows, err := h.repo.ListMovers(r.Context(), direction, 100, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list movers")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = h.cache.SetMovers(r.Context(), direction, category, rows)

	entries := ToEntries(rows)
	if len(entries) > limit {
		entries = entries[:limit]
	}
		writeJSON(w, MoversResponse{
			Direction: direction,
			PeriodEnd: periodOrZero(rows),
			Entries:   entries,
			Category:  string(category),
		})
}
// HandleGetMyCity: GET /city-rankings/me (auth required)
func (h *Handler) HandleGetMyCity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())

	optedOut, err := h.repo.IsOptedOut(r.Context(), userID.String())
	if err != nil {
		h.log.WithError(err).Error("Failed to read opt-out")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if optedOut {
		writeJSON(w, MyCityResponse{HasCity: false, OptedOut: true})
		return
	}

	if cached, hit, err := h.cache.GetMyCity(r.Context(), userID.String()); err == nil && hit && cached != nil {
		writeJSON(w, MyCityResponse{
			HasCity: true,
			Metro:   ptrEntry(ToEntry(*cached)),
		})
		return
	}

	metro, err := h.repo.GetUserMetro(r.Context(), userID.String())
	if err != nil || metro == nil {
		writeJSON(w, MyCityResponse{HasCity: false})
		return
	}
	rk, err := h.repo.GetRankingBySlug(r.Context(), metro.Slug, cityranking.CategoryComposite)
	if err != nil || rk == nil {
		writeJSON(w, MyCityResponse{HasCity: false})
		return
	}
	_ = h.cache.SetMyCity(r.Context(), userID.String(), rk)
	writeJSON(w, MyCityResponse{
		HasCity: true,
		Metro:   ptrEntry(ToEntry(*rk)),
	})
}

// HandleResolveCity: POST /city-rankings/resolve (auth, internal helper)
// Body: { "input": "Austin, TX" }
func (h *Handler) HandleResolveCity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())

	var body struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	norm := cityranking.NormalizeInput(body.Input)
	if norm == "" {
		http.Error(w, "input required", http.StatusBadRequest)
		return
	}
	cities, err := h.repo.LookupCityByAlias(r.Context(), norm)
	if err != nil {
		h.log.WithError(err).Error("Failed to lookup city by alias")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(cities) == 0 {
		// Try splitting on state abbreviation.
		city, state := cityranking.ExpandStateAbbreviations(norm)
		if city != "" {
			aliasKey := city + " " + strings.ToLower(state)
			cities, err = h.repo.LookupCityByAlias(r.Context(), aliasKey)
			if err != nil {
				h.log.WithError(err).Warn("alias lookup failed")
			}
			if len(cities) == 0 {
				cities, err = h.repo.SearchCitiesByName(r.Context(), city)
				if err != nil {
					h.log.WithError(err).Warn("city name search failed")
				}
			}
		}
	}
	if len(cities) == 0 {
		// Try geocoding as last resort before giving up.
		if h.geo == nil {
			h.log.WithFields(logrus.Fields{
				"input":   norm,
				"user_id": userID.String(),
			}).Warn("city geocoder unavailable — h.geo is nil; check if cityRankingPool initialized")
			http.Error(w, "no city match", http.StatusNotFound)
			return
		}
		result, geoErr := h.geo.Geocode(r.Context(), norm)
		if geoErr == nil && result != nil {
			h.log.WithFields(logrus.Fields{
				"input":         norm,
				"geocoded_city": result.Address.City,
				"state_code":    result.Address.StateCode,
				"country_code":  result.Address.CountryCode,
				"display_name":  result.DisplayName,
			}).Info("geocoding produced a result, auto-creating city")

			stateCode := result.Address.StateCode
			if stateCode == "" {
				stateCode = result.Address.CountryCode
			}
			citySlug := cityranking.SlugifyCity(result.Address.City) + "-" + strings.ToLower(stateCode)

			// Get or create metro for this state/country.
			metro, metroErr := h.repo.GetOrCreateMetroByState(r.Context(), stateCode, result.Address.CountryCode)
			if metroErr != nil {
				h.log.WithError(metroErr).Warn("failed to get/create metro for geocoded city")
			}

			var metroID *int64
			if metro != nil {
				metroID = &metro.ID
			}

			// Create the city with geocoded coords. All geocoded cities are marked
			// 'pending' review since we can't verify population from Nominatim.
			// Admin can approve/reject via the /admin/cities/{id}/review endpoint.
			input := cityranking.CityCreateInput{
				Name:        result.Address.City,
				Slug:        citySlug,
				StateCode:   result.Address.StateCode,
				StateName:   result.Address.State,
				CountryCode: result.Address.CountryCode,
				CountryName: result.Address.Country,
				Latitude:    result.Lat.Value,
				Longitude:   result.Lon.Value,
				Population:  0,
				MetroID:     metroID,
			}
			cityID, _, cityErr := h.repo.UpsertCityForReview(r.Context(), input, cityranking.CityReviewStatusPending, true)
			if cityErr != nil {
				h.log.WithError(cityErr).Warn("failed to upsert geocoded city")
				http.Error(w, "failed to resolve city", http.StatusInternalServerError)
				return
			}

			// Add the user's typed input as an alias so future lookups work.
			_ = h.repo.AddCityAlias(r.Context(), cityID, norm, "user_geocode")

			// Assign to user.
			if err := h.repo.AssignUserCity(r.Context(), userID.String(), &cityID, "user_geocode"); err != nil {
				h.log.WithError(err).Warn("failed to assign geocoded city")
			}

			writeJSON(w, CityResolution{
				CityID:        cityID,
				CitySlug:      citySlug,
				CityName:      result.Address.City,
				MetroSlug:     "",
				Source:        "geocode_pending_review",
				Ambiguous:     false,
				PendingReview: true,
			})
			return
		}
		if geoErr != nil {
			h.log.WithError(geoErr).WithField("input", norm).Warn("geocoding failed")
		}
		http.Error(w, "no city match", http.StatusNotFound)
		return
	}
	ambiguous := len(cities) > 1
	first := cities[0]
	metroSlug := ""
	if first.MetroSlug != nil {
		metroSlug = *first.MetroSlug
	}
	source := "fallback"
	if !ambiguous {
		source = "alias"
	}
	// Persist the assignment on the user.
	var cityPtr *int64
	if !ambiguous {
		id := first.ID
		cityPtr = &id
	}
	if err := h.repo.AssignUserCity(r.Context(), userID.String(), cityPtr, "user_input"); err != nil {
		h.log.WithError(err).Warn("failed to assign user city")
	}
	writeJSON(w, CityResolution{
		CityID:    first.ID,
		CitySlug:  first.Slug,
		CityName:  first.Name,
		MetroSlug: metroSlug,
		Source:    source,
		Ambiguous: ambiguous,
	})
}

// HandleResolveByIP: GET /city-rankings/resolve-by-ip (public, no auth)
//
// Best-effort IP→city lookup, used as a fallback when the user hasn't set
// their `Location` on the profile. The caller passes an `?ip=` query param
// (e.g. `?ip=8.8.8.8`). For privacy, we never log or persist the IP, and
// the response is cached for 1h.
//
// Returns:
//   200 + { city, country_code, source } on hit
//   200 + { not_found: true, country_code } on a country we know but no city
//   503 when the IP-geo resolver is not configured (e.g. dev environment
//       without MaxMind)
func (h *Handler) HandleResolveByIP(w http.ResponseWriter, r *http.Request) {
	if h.ipgeo == nil {
		http.Error(w, "ip geo resolver not configured", http.StatusServiceUnavailable)
		return
	}
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	if ip == "" {
		http.Error(w, "ip query param required", http.StatusBadRequest)
		return
	}
	res, err := h.ipgeo.Resolve(r.Context(), ip)
	if err != nil {
		h.log.WithError(err).Warn("IP geo resolve failed")
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}
	resp := IPGeoResponse{
		CountryCode: res.CountryCode,
		Source:      res.Source,
		NotFound:    res.NotFound,
	}
	if res.City != nil {
		metroSlug := ""
		metroName := ""
		if res.City.MetroSlug != nil {
			metroSlug = *res.City.MetroSlug
		}
		if res.City.MetroName != nil {
			metroName = *res.City.MetroName
		}
		resp.City = &IPGeoCity{
			CityID:      res.City.CityID,
			CitySlug:    res.City.CitySlug,
			CityName:    res.City.CityName,
			StateCode:   res.City.StateCode,
			CountryCode: res.City.CountryCode,
			MetroSlug:   metroSlug,
			MetroName:   metroName,
			Population:  res.City.Population,
		}
	}
	writeJSON(w, resp)
}

// HandleSetMyCity: POST /users/me/city (auth)
//
// Sets the caller's city. If the request body has a `slug` or `input`, that
// is used. Otherwise, the handler falls back to the IP geo resolver (if
// configured) so a fresh signup from a non-US IP can still land on the
// leaderboard without typing anything.
func (h *Handler) HandleSetMyCity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())

	var body struct {
		Slug  string `json:"slug"`
		Input string `json:"input"`
	}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
	}
	if body.Input != "" {
		// Reuse the alias-based resolver.
		norm := cityranking.NormalizeInput(body.Input)
		cities, err := h.repo.LookupCityByAlias(r.Context(), norm)
		if err == nil && len(cities) > 0 {
			id := cities[0].ID
			if err := h.repo.AssignUserCity(r.Context(), userID.String(), &id, "user_input"); err != nil {
				h.log.WithError(err).Warn("AssignUserCity (input) failed")
			}
			writeJSON(w, CityResolution{
				CityID:   cities[0].ID,
				CitySlug: cities[0].Slug,
				CityName: cities[0].Name,
				Source:   "alias",
			})
			return
		}
	}
	if body.Slug != "" {
		city, err := h.repo.GetCityBySlug(r.Context(), body.Slug)
		if err == nil && city != nil {
			if err := h.repo.AssignUserCity(r.Context(), userID.String(), &city.ID, "user_input"); err != nil {
				h.log.WithError(err).Warn("AssignUserCity (slug) failed")
			}
			writeJSON(w, CityResolution{
				CityID:   city.ID,
				CitySlug: city.Slug,
				CityName: city.Name,
				Source:   "slug",
			})
			return
		}
	}
	// IP fallback: only kicks in if the resolver is wired up. We deliberately
	// do NOT return 4xx here — a fresh signup from a fresh IP should be
	// silently ranked if possible, and a JSON response lets the front-end
	// distinguish "ip detected a city" from "no IP geo available".
	if h.ipgeo == nil {
		writeJSON(w, CityResolution{Source: "no_ip_geo"})
		return
	}
	ip := clientIP(r)
	if ip == "" {
		writeJSON(w, CityResolution{Source: "no_ip"})
		return
	}
	res, err := h.ipgeo.Resolve(r.Context(), ip)
	if err != nil || res == nil || res.City == nil {
		writeJSON(w, CityResolution{Source: "ip_unresolved"})
		return
	}
	id := res.City.CityID
	if err := h.repo.AssignUserCity(r.Context(), userID.String(), &id, "ip_geo"); err != nil {
		h.log.WithError(err).Warn("AssignUserCity (ip) failed")
	}
	metroSlug := ""
	if res.City.MetroSlug != nil {
		metroSlug = *res.City.MetroSlug
	}
	writeJSON(w, CityResolution{
		CityID:    res.City.CityID,
		CitySlug:  res.City.CitySlug,
		CityName:  res.City.CityName,
		MetroSlug: metroSlug,
		Source:    "ip_geo",
	})
}

// clientIP extracts the caller's IP from common proxy headers, falling back
// to RemoteAddr. Used only for the IP geo fallback — never logged.
func clientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		if i := strings.Index(h, ","); i >= 0 {
			return strings.TrimSpace(h[:i])
		}
		return strings.TrimSpace(h)
	}
	if h := r.Header.Get("X-Real-IP"); h != "" {
		return strings.TrimSpace(h)
	}
	if h := r.Header.Get("CF-Connecting-IP"); h != "" {
		return strings.TrimSpace(h)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// HandleSetOptOut: POST /users/me/city-ranking-opt-out
func (h *Handler) HandleSetOptOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())

	var body SetOptOutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OptedOut == nil {
		http.Error(w, "opted_out (bool) required", http.StatusBadRequest)
		return
	}
	if err := h.repo.SetOptOut(r.Context(), userID.String(), *body.OptedOut); err != nil {
		h.log.WithError(err).Error("Failed to set opt-out")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, OptOutResponse{OptedOut: *body.OptedOut})
}

// HandleListStates: GET /city-rankings/states
func (h *Handler) HandleListStates(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetStates(r.Context(), country, category); err == nil && hit {
		entries := ToStateEntries(cached)
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, StatesLeaderboardResponse{
			PeriodEnd:   statePeriod(cached),
			TotalStates: len(cached),
			Entries:     entries,
			Country:     country,
			Category:    string(category),
			CacheHit:    true,
		})
		return
	}

	rows, err := h.repo.ListStateRankings(r.Context(), country, 500, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list state rankings")
		http.Error(w, "failed to list state rankings", http.StatusInternalServerError)
		return
	}
	if len(rows) > 0 {
		_ = h.cache.SetStates(r.Context(), country, category, rows)
	}
	entries := ToStateEntries(rows)
	if len(entries) > limit {
		entries = entries[:limit]
	}
	periodEnd := time.Time{}
	if len(rows) > 0 {
		periodEnd = rows[0].PeriodEnd
	}
	writeJSON(w, StatesLeaderboardResponse{
		PeriodEnd:   periodEnd,
		TotalStates: len(rows),
		Entries:     entries,
		Country:     country,
		Category:    string(category),
	})
}

// HandleGetState: GET /city-rankings/states/{code}
func (h *Handler) HandleGetState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := strings.ToUpper(strings.TrimSpace(vars["code"]))
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	if country == "" {
		country = "US"
	}
	if code == "" {
		http.Error(w, "state code required", http.StatusBadRequest)
		return
	}
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetState(r.Context(), country, code, category); err == nil && hit && cached != nil {
		writeJSON(w, StateDetailResponse{
			Current:               ptrState(ToStateEntry(*cached)),
			PrivacyMinActiveUsers: cityranking.MinActiveUsersForPublic,
			PeriodEnd:             cached.PeriodEnd,
			Category:              string(category),
			CacheHit:              true,
		})
		return
	}

	s, err := h.repo.GetStateRankingByCode(r.Context(), country, code, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to get state")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, "state not found or not ranked", http.StatusNotFound)
		return
	}
	// Look up this state's rank position from the global leaderboard.
	all, err := h.repo.ListStateRankings(r.Context(), country, 500, category)
	if err == nil {
		for i, row := range all {
			if row.StateCode == s.StateCode && row.CountryCode == s.CountryCode {
				s.RankPosition = i + 1
				break
			}
		}
	}
	_ = h.cache.SetState(r.Context(), country, code, category, s)
	writeJSON(w, StateDetailResponse{
		Current:               ptrState(ToStateEntry(*s)),
		PrivacyMinActiveUsers: cityranking.MinActiveUsersForPublic,
		PeriodEnd:             s.PeriodEnd,
		Category:              string(category),
	})
}

// HandleListMapPoints: GET /city-rankings/map
// Returns all ranked metros with (lat, lon) and a tier label for the AI
// World Map visualization.
func (h *Handler) HandleListMapPoints(w http.ResponseWriter, r *http.Request) {
	category := parseCategory(r)
	if cached, hit, err := h.cache.GetMapPoints(r.Context(), category); err == nil && hit {
		period, _ := h.repo.LatestPeriod(r.Context())
		writeJSON(w, MapPointsResponse{
			PeriodEnd: period,
			Points:    toMapPointsResponse(cached),
			Category:  string(category),
			CacheHit:  true,
		})
		return
	}

	pts, err := h.repo.ListMapPoints(r.Context(), category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list map points")
		http.Error(w, "failed to list map points", http.StatusInternalServerError)
		return
	}
	period, _ := h.repo.LatestPeriod(r.Context())
	if len(pts) > 0 {
		_ = h.cache.SetMapPoints(r.Context(), category, pts)
	}
	writeJSON(w, MapPointsResponse{
		PeriodEnd: period,
		Points:    toMapPointsResponse(pts),
		Category:  string(category),
	})
}

// HandleGetOptOut: GET /users/me/city-ranking-opt-out
func (h *Handler) HandleGetOptOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())
	v, err := h.repo.IsOptedOut(r.Context(), userID.String())
	if err != nil {
		h.log.WithError(err).Error("Failed to read opt-out")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, OptOutResponse{OptedOut: v})
}

// HandleListCities: GET /city-rankings/cities
//
// City-proper leaderboard — every active city, ranked by its own score
// (instead of being rolled up to its metro). Useful for the front-end toggle
// "show city proper, not MSA". On-the-fly aggregation, so it is *not* cached
// in Redis in v1 (a single 100-row request runs ~100 sub-queries).
func (h *Handler) HandleListCities(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	category := parseCategory(r)

	rows, err := h.repo.ListCityRankings(r.Context(), country, limit, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list city rankings")
		http.Error(w, "failed to list city rankings", http.StatusInternalServerError)
		return
	}
	period, _ := h.repo.LatestPeriod(r.Context())
	writeJSON(w, CitiesLeaderboardResponse{
		PeriodEnd:   period,
		TotalRanked: len(rows),
		Entries:     ToCityEntries(rows),
		Country:     country,
		Category:    string(category),
		PrivacyMin:  cityranking.MinActiveUsersForPublic,
	})
}

// HandleListBuilders: GET /city-rankings/{slug}/builders
//
// Anonymized top contributors in a metro. Returns an empty list (not a 404)
// when the metro has fewer than k=5 active builders — that is the privacy
// contract. The front-end renders "builders hidden" when the list is empty
// and PrivacySuppressed is true.
func (h *Handler) HandleListBuilders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	slug = strings.TrimPrefix(slug, "/")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	category := parseCategory(r)
	limit := parseLimit(r.URL.Query().Get("limit"), 25, 100)

	// Verify the metro exists so we 404 cleanly on typos.
	metro, err := h.repo.GetMetroBySlug(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get metro for builders")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if metro == nil {
		http.Error(w, "metro not found", http.StatusNotFound)
		return
	}

	rows, err := h.repo.ListBuilders(r.Context(), slug, limit, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list builders")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	period, _ := h.repo.LatestPeriod(r.Context())
	suppressed := len(rows) == 0
	writeJSON(w, BuildersResponse{
		MetroSlug:         slug,
		PeriodEnd:         period,
		Entries:           ToBuilderEntries(rows),
		Category:          string(category),
		PrivacySuppressed: suppressed,
	})
}

// HandleListUniversities: GET /city-rankings/{slug}/universities
func (h *Handler) HandleListUniversities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	slug = strings.TrimPrefix(slug, "/")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	limit := parseLimit(r.URL.Query().Get("limit"), 20, 50)

	metro, err := h.repo.GetMetroBySlug(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get metro for universities")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if metro == nil {
		http.Error(w, "metro not found", http.StatusNotFound)
		return
	}

	rows, err := h.repo.ListUniversitiesByMetro(r.Context(), slug, limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to list universities for metro")
		http.Error(w, "failed to list universities for metro", http.StatusInternalServerError)
		return
	}

	writeJSON(w, MetroUniversitiesResponse{
		MetroSlug: slug,
		Entries:   rows,
		Total:     len(rows),
	})
}

// HandleListPendingCityReviews: GET /admin/cities/pending (auth + admin)
func (h *Handler) HandleListPendingCityReviews(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 50, 100)
	offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))

	cities, err := h.repo.ListCitiesPendingReview(r.Context(), limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list pending city reviews")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	total, err := h.repo.CountCitiesPendingReview(r.Context())
	if err != nil {
		h.log.WithError(err).Warn("Failed to count pending reviews")
	}
	writeJSON(w, CityReviewListResponse{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Cities: toCityReviewEntries(cities),
	})
}

// HandleListAllCitiesAdmin: GET /admin/cities (auth + admin)
func (h *Handler) HandleListAllCitiesAdmin(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))

	cities, err := h.repo.ListAllCitiesForAdmin(r.Context(), limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list all cities for admin")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, CityAdminListResponse{
		Total:  len(cities),
		Limit:  limit,
		Offset: offset,
		Cities: cities,
	})
}

// HandleGetCityReview: GET /admin/cities/{id}/review (auth + admin)
func (h *Handler) HandleGetCityReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := strings.TrimSpace(vars["id"])
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid city id", http.StatusBadRequest)
		return
	}

	city, err := h.repo.GetCityForReview(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get city for review")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if city == nil {
		http.Error(w, "city not found", http.StatusNotFound)
		return
	}

	aliases, err := h.repo.ListCityAliases(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Warn("Failed to list city aliases")
	}

	writeJSON(w, CityReviewDetail{
		CityID:      city.ID,
		Slug:        city.Slug,
		Name:        city.Name,
		StateCode:   city.StateCode,
		StateName:   city.StateName,
		CountryCode: city.CountryCode,
		CountryName: city.CountryName,
		Latitude:    city.Latitude,
		Longitude:   city.Longitude,
		Population:  city.Population,
		ReviewStatus: string(city.ReviewStatus),
		ReviewedAt:  city.ReviewedAt,
		ReviewedBy:  city.ReviewedBy,
		ReviewNotes: &city.ReviewNotes,
		CreatedAt:   city.CreatedAt,
		Aliases:     toCityAliasEntries(aliases),
	})
}

// HandleReviewCity: POST /admin/cities/{id}/review (auth + admin)
func (h *Handler) HandleReviewCity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	idStr := strings.TrimSpace(vars["id"])
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid city id", http.StatusBadRequest)
		return
	}

	var body ReviewCityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Status != "approved" && body.Status != "rejected" {
		http.Error(w, "status must be 'approved' or 'rejected'", http.StatusBadRequest)
		return
	}

	status := cityranking.CityReviewStatusApproved
	if body.Status == "rejected" {
		status = cityranking.CityReviewStatusRejected
	}

	if err := h.repo.ReviewCity(r.Context(), id, claims.UserID.String(), status, body.Notes); err != nil {
		h.log.WithError(err).Error("Failed to review city")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	city, _ := h.repo.GetCityForReview(r.Context(), id)
	slug := ""
	if city != nil {
		slug = city.Slug
	}

	writeJSON(w, CityReviewResponse{
		CityID:      id,
		Slug:        slug,
		ReviewStatus: body.Status,
		Message:     "City " + body.Status,
	})
}

func toCityReviewEntries(cities []cityranking.CityReviewSummary) []CityReviewEntry {
	out := make([]CityReviewEntry, 0, len(cities))
	for _, c := range cities {
		out = append(out, CityReviewEntry{
			CityID:      c.CityID,
			Slug:        c.Slug,
			Name:        c.Name,
			StateCode:   c.StateCode,
			StateName:   c.StateName,
			CountryCode: c.CountryCode,
			CountryName: c.CountryName,
			Latitude:    c.Latitude,
			Longitude:   c.Longitude,
			Population:  c.Population,
			MetroName:   c.MetroName,
			CreatedAt:   c.CreatedAt,
			AliasCount:  c.AliasCount,
		})
	}
	return out
}

func toCityAliasEntries(aliases []cityranking.CityAliasDetail) []CityAliasEntry {
	out := make([]CityAliasEntry, 0, len(aliases))
	for _, a := range aliases {
		out = append(out, CityAliasEntry{
			Alias:   a.Alias,
			Source:  a.Source,
			Created: a.Created,
		})
	}
	return out
}

// ── helpers ──────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func ptrEntry(e CityRankingEntry) *CityRankingEntry { return &e }

// parseCategory reads the ?category= query param and returns the canonical
// Category. Invalid categories fall back to "composite" rather than 400ing
// — the API is forgiving on the read path so a bad link still shows
// something useful.
func parseCategory(r *http.Request) cityranking.Category {
	raw := strings.TrimSpace(r.URL.Query().Get("category"))
	if raw == "" {
		return cityranking.CategoryComposite
	}
	if cityranking.IsValidCategory(raw) {
		return cityranking.Category(raw)
	}
	return cityranking.CategoryComposite
}

func parseLimit(s string, def, max int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func periodOrZero(rs []cityranking.Ranking) (t time.Time) {
	if len(rs) > 0 {
		return rs[0].PeriodEnd
	}
	return t
}

func currentPeriod(history []cityranking.Ranking, current *cityranking.Ranking) time.Time {
	if current != nil {
		return current.PeriodEnd
	}
	return periodOrZero(history)
}

func statePeriod(ss []cityranking.StateRanking) (t time.Time) {
	if len(ss) > 0 {
		return ss[0].PeriodEnd
	}
	return t
}

func ptrState(s StateRankingEntry) *StateRankingEntry { return &s }

// toMapPointsResponse is a thin adapter for the map cache hit path: the cache
// stores crstorage.MapPoint (the storage type), and we need to convert to the
// DTO type. We avoid importing the storage package in the dto file by doing it
// here.
func toMapPointsResponse(pts []cityranking.MapPoint) []MapPointEntry {
	out := make([]MapPointEntry, 0, len(pts))
	for _, p := range pts {
		out = append(out, MapPointEntry{
			MetroSlug:      p.MetroSlug,
			MetroName:      p.MetroName,
			CountryCode:    p.CountryCode,
			StateCode:      p.StateCode,
			Latitude:       p.Latitude,
			Longitude:      p.Longitude,
			Population:     p.Population,
			RankPosition:   p.RankPosition,
			ScorePerCapita: p.ScorePerCapita,
			ActiveUsers:    p.ActiveUsers,
			Tier:           p.Tier,
		})
	}
	return out
}
