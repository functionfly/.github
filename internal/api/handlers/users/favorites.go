package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/apierror"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type FavoritesHandler struct {
	repo storage.Repository
}

func NewFavoritesHandler(repo storage.Repository) *FavoritesHandler {
	return &FavoritesHandler{repo: repo}
}

func (h *FavoritesHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *FavoritesHandler) getCurrentUserID(r *http.Request) (uuid.UUID, error) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return uuid.Nil, http.ErrNoCookie
	}
	return claims.UserID, nil
}

type AddFavoriteRequest struct {
	FunctionID string `json:"function_id"`
	Position   int    `json:"position"`
}

type ToggleFavoriteResponse struct {
	Favorited bool `json:"favorited"`
}

type FavoriteItem struct {
	ID         string `json:"id"`
	FunctionID string `json:"function_id"`
	Position   int    `json:"position"`
	CreatedAt  string `json:"created_at"`
}

type ListFavoritesResponse struct {
	Favorites []FavoriteItem `json:"favorites"`
	Total     int            `json:"total"`
}

func (h *FavoritesHandler) HandleListFavorites(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	favorites, total, err := h.repo.GetUserFavorites(r.Context(), userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Error("Failed to get user favorites")
		apierror.WriteError(w, apierror.NewInternal("failed to get favorites"))
		return
	}

	items := make([]FavoriteItem, 0, len(favorites))
	for _, fav := range favorites {
		items = append(items, FavoriteItem{
			ID:         fav.ID.String(),
			FunctionID: fav.FunctionID.String(),
			Position:   fav.Position,
			CreatedAt:  fav.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	h.writeJSON(w, http.StatusOK, ListFavoritesResponse{
		Favorites: items,
		Total:     total,
	})
}

func (h *FavoritesHandler) HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	var req AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid function_id"))
		return
	}

	position := req.Position
	if position == 0 {
		existing, _, _ := h.repo.GetUserFavorites(r.Context(), userID, 1000, 0)
		position = len(existing)
	}

	_, err = h.repo.AddFavorite(r.Context(), userID, functionID, position)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to add favorite")
		apierror.WriteError(w, apierror.NewInternal("failed to add favorite"))
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":     "favorite added",
		"function_id": functionID.String(),
		"favorited":   true,
	})
}

func (h *FavoritesHandler) HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid function_id"))
		return
	}

	err = h.repo.RemoveFavorite(r.Context(), userID, functionID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to remove favorite")
		apierror.WriteError(w, apierror.NewInternal("failed to remove favorite"))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "favorite removed",
		"function_id": functionID.String(),
		"favorited":   false,
	})
}

func (h *FavoritesHandler) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid function_id"))
		return
	}

	isFav, err := h.repo.IsFavorite(r.Context(), userID, functionID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to check favorite status")
		apierror.WriteError(w, apierror.NewInternal("failed to check favorite status"))
		return
	}

	if isFav {
		err = h.repo.RemoveFavorite(r.Context(), userID, functionID)
		if err != nil {
			logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to remove favorite")
			apierror.WriteError(w, apierror.NewInternal("failed to remove favorite"))
			return
		}
		h.writeJSON(w, http.StatusOK, ToggleFavoriteResponse{Favorited: false})
	} else {
		existing, _, _ := h.repo.GetUserFavorites(r.Context(), userID, 1000, 0)
		position := len(existing)
		_, err = h.repo.AddFavorite(r.Context(), userID, functionID, position)
		if err != nil {
			logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to add favorite")
			apierror.WriteError(w, apierror.NewInternal("failed to add favorite"))
			return
		}
		h.writeJSON(w, http.StatusOK, ToggleFavoriteResponse{Favorited: true})
	}
}

func (h *FavoritesHandler) HandleCheckFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid function_id"))
		return
	}

	isFav, err := h.repo.IsFavorite(r.Context(), userID, functionID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to check favorite status")
		apierror.WriteError(w, apierror.NewInternal("failed to check favorite status"))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"function_id": functionID.String(),
		"favorited":   isFav,
	})
}

func (h *FavoritesHandler) HandleUpdateFavoritePosition(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("authentication required"))
		return
	}

	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid function_id"))
		return
	}

	var req struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	err = h.repo.UpdateFavoritePosition(r.Context(), userID, functionID, req.Position)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).WithField("functionID", functionID).Error("Failed to update favorite position")
		apierror.WriteError(w, apierror.NewInternal("failed to update favorite position"))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "position updated"})
}
