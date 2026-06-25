package citywarhandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo *cityranking.Repository
	log  *logrus.Logger
}

func NewHandler(repo *cityranking.Repository, log *logrus.Logger) *Handler {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Handler{repo: repo, log: log}
}

// HandleListWars: GET /city-wars
func (h *Handler) HandleListWars(w http.ResponseWriter, r *http.Request) {
	wars, err := h.repo.ListWars(r.Context(), 20)
	if err != nil {
		h.log.WithError(err).Error("Failed to list wars")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, WarsResponse{Wars: wars})
}

// HandleGetWar: GET /city-wars/{slug}
func (h *Handler) HandleGetWar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	war, err := h.repo.GetWar(r.Context(), slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get war")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if war == nil {
		http.Error(w, "war not found", http.StatusNotFound)
		return
	}
	writeJSON(w, WarResponse{War: war})
}

// HandleGetLatestWar: GET /city-wars/latest
func (h *Handler) HandleGetLatestWar(w http.ResponseWriter, r *http.Request) {
	war, err := h.repo.LatestWar(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get latest war")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if war == nil {
		writeJSON(w, WarResponse{War: nil})
		return
	}
	writeJSON(w, WarResponse{War: war})
}

// HandleGetWarChampions: GET /city-wars/champions
func (h *Handler) HandleGetWarChampions(w http.ResponseWriter, r *http.Request) {
	wars, err := h.repo.ListWars(r.Context(), 10)
	if err != nil {
		h.log.WithError(err).Error("Failed to list wars for champions")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var champions []WarChampion
	for _, w := range wars {
		if w.ChampionMetroID != nil {
			champions = append(champions, WarChampion{
				WarSlug:     w.Slug,
				WarSeason:   w.Season,
				WarEndsAt:   w.EndsAt,
				MetroID:     *w.ChampionMetroID,
				MetroSlug:   w.ChampionSlug,
				MetroName:   w.ChampionName,
				CountryCode: w.ChampionCountry,
			})
		}
	}
	writeJSON(w, ChampionsResponse{Champions: champions})
}

// ── helpers ───────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

// WarChampion is a past war champion for display on the homepage.
type WarChampion struct {
	WarSlug     string    `json:"war_slug"`
	WarSeason   string    `json:"war_season"`
	WarEndsAt   time.Time `json:"war_ends_at"`
	MetroID     int64     `json:"metro_id"`
	MetroSlug   string    `json:"metro_slug"`
	MetroName   string    `json:"metro_name"`
	CountryCode string    `json:"country_code"`
}

// ChampionsResponse is the payload for GET /city-wars/champions.
type ChampionsResponse struct {
	Champions []WarChampion `json:"champions"`
}

// WarsResponse is the payload for GET /city-wars.
type WarsResponse struct {
	Wars []cityranking.War `json:"wars"`
}

// WarResponse is the payload for GET /city-wars/{slug} and /city-wars/latest.
type WarResponse struct {
	War *cityranking.War `json:"war,omitempty"`
}
