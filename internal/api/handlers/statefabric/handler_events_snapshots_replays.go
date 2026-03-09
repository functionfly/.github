package statefabric

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	statefabricstorage "github.com/functionfly/functionfly/internal/storage/statefabric"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListEvents GET /v1/state-fabrics/{id}/events
func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	var storeID *uuid.UUID
	if s := q.Get("storeId"); s != "" {
		parsed, _ := uuid.Parse(s)
		storeID = &parsed
	}
	eventType := q.Get("eventType")
	var startTime, endTime *time.Time
	if t := q.Get("startTime"); t != "" {
		parsed, _ := time.Parse(time.RFC3339, t)
		startTime = &parsed
	}
	if t := q.Get("endTime"); t != "" {
		parsed, _ := time.Parse(time.RFC3339, t)
		endTime = &parsed
	}
	list, total, err := h.repo.ListEvents(r.Context(), id, storeID, eventType, startTime, endTime, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("list events")
		writeErr(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	events := make([]map[string]interface{}, 0, len(list))
	for _, e := range list {
		storeIDStr := ""
		if e.StoreID != nil {
			storeIDStr = e.StoreID.String()
		}
		events = append(events, map[string]interface{}{
			"id":             e.ID.String(),
			"fabricId":       e.FabricID.String(),
			"storeId":        storeIDStr,
			"eventType":      e.EventType,
			"payload":        e.Payload,
			"timestamp":      e.Timestamp.UTC().Format(time.RFC3339),
			"sequenceNumber": e.SequenceNumber,
			"correlationId":  e.CorrelationID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"events": events, "total": total})
}

// HandleListSnapshots GET /v1/state-fabrics/{id}/snapshots
func (h *Handler) HandleListSnapshots(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	var storeID *uuid.UUID
	if s := r.URL.Query().Get("storeId"); s != "" {
		parsed, _ := uuid.Parse(s)
		storeID = &parsed
	}
	list, err := h.repo.ListSnapshotsByFabric(r.Context(), id, storeID)
	if err != nil {
		logrus.WithError(err).Error("list snapshots")
		writeErr(w, http.StatusInternalServerError, "failed to list snapshots")
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		storeIDStr := ""
		if s.StoreID != nil {
			storeIDStr = s.StoreID.String()
		}
		expiresAt := ""
		if s.ExpiresAt != nil {
			expiresAt = s.ExpiresAt.UTC().Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"id":          s.ID.String(),
			"fabricId":    s.FabricID.String(),
			"storeId":     storeIDStr,
			"name":        s.Name,
			"description": s.Description,
			"state":       s.StateData,
			"eventCount":  s.EventCount,
			"sizeBytes":   s.SizeBytes,
			"createdAt":   s.CreatedAt.UTC().Format(time.RFC3339),
			"expiresAt":   expiresAt,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// HandleCreateSnapshot POST /v1/state-fabrics/{id}/snapshots
func (h *Handler) HandleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		StoreID     string `json:"storeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	var storeID *uuid.UUID
	if body.StoreID != "" {
		parsed, _ := uuid.Parse(body.StoreID)
		storeID = &parsed
	}
	snap := &statefabricstorage.StateFabricSnapshot{
		FabricID:    id,
		StoreID:     storeID,
		Name:        body.Name,
		Description: body.Description,
		StateData:   statefabricstorage.JSONMap{},
		EventCount:  0,
		SizeBytes:   0,
	}
	if err := h.repo.CreateSnapshot(r.Context(), snap); err != nil {
		logrus.WithError(err).Error("create snapshot")
		writeErr(w, http.StatusInternalServerError, "failed to create snapshot")
		return
	}
	storeIDStr := ""
	if snap.StoreID != nil {
		storeIDStr = snap.StoreID.String()
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          snap.ID.String(),
		"fabricId":    snap.FabricID.String(),
		"storeId":     storeIDStr,
		"name":        snap.Name,
		"description": snap.Description,
		"state":       snap.StateData,
		"eventCount":  snap.EventCount,
		"sizeBytes":   snap.SizeBytes,
		"createdAt":   snap.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// HandleDeleteSnapshot DELETE /v1/state-fabrics/{fabricId}/snapshots/{snapshotId}
func (h *Handler) HandleDeleteSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	snapshotID, err := uuid.Parse(vars["snapshotId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid snapshotId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	snap, err := h.repo.GetSnapshotByID(r.Context(), snapshotID)
	if err != nil || snap == nil || snap.FabricID != fabricID {
		writeErr(w, http.StatusNotFound, "snapshot not found")
		return
	}
	if err := h.repo.DeleteSnapshot(r.Context(), snapshotID); err != nil {
		logrus.WithError(err).Error("delete snapshot")
		writeErr(w, http.StatusInternalServerError, "failed to delete snapshot")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleListReplays GET /v1/state-fabrics/{id}/replays
func (h *Handler) HandleListReplays(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	list, err := h.repo.ListReplaysByFabric(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("list replays")
		writeErr(w, http.StatusInternalServerError, "failed to list replays")
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		result = append(result, replayToAPI(s))
	}
	writeJSON(w, http.StatusOK, result)
}

// HandleCreateReplay POST /v1/state-fabrics/{id}/replays
func (h *Handler) HandleCreateReplay(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	var body struct {
		SnapshotID   string `json:"snapshotId"`
		StartEventID string `json:"startEventId"`
		EndEventID   string `json:"endEventId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	var snapshotID, startEventID, endEventID *uuid.UUID
	if body.SnapshotID != "" {
		parsed, _ := uuid.Parse(body.SnapshotID)
		snapshotID = &parsed
	}
	if body.StartEventID != "" {
		parsed, _ := uuid.Parse(body.StartEventID)
		startEventID = &parsed
	}
	if body.EndEventID != "" {
		parsed, _ := uuid.Parse(body.EndEventID)
		endEventID = &parsed
	}
	replay := &statefabricstorage.StateFabricReplay{
		FabricID:     id,
		SnapshotID:   snapshotID,
		StartEventID: startEventID,
		EndEventID:   endEventID,
		Status:       "pending",
		Progress:     0,
	}
	if err := h.repo.CreateReplay(r.Context(), replay); err != nil {
		logrus.WithError(err).Error("create replay")
		writeErr(w, http.StatusInternalServerError, "failed to create replay")
		return
	}
	writeJSON(w, http.StatusCreated, replayToAPI(replay))
}

// HandleGetReplay GET /v1/state-fabrics/{fabricId}/replays/{replayId}
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	replayID, err := uuid.Parse(vars["replayId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid replayId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	replay, err := h.repo.GetReplayByID(r.Context(), replayID)
	if err != nil || replay == nil || replay.FabricID != fabricID {
		writeErr(w, http.StatusNotFound, "replay not found")
		return
	}
	writeJSON(w, http.StatusOK, replayToAPI(replay))
}

func replayToAPI(s *statefabricstorage.StateFabricReplay) map[string]interface{} {
	snapshotID := ""
	if s.SnapshotID != nil {
		snapshotID = s.SnapshotID.String()
	}
	startEventID := ""
	if s.StartEventID != nil {
		startEventID = s.StartEventID.String()
	}
	endEventID := ""
	if s.EndEventID != nil {
		endEventID = s.EndEventID.String()
	}
	completedAt := ""
	if s.CompletedAt != nil {
		completedAt = s.CompletedAt.UTC().Format(time.RFC3339)
	}
	errMsg := ""
	if s.ErrorMessage != nil {
		errMsg = *s.ErrorMessage
	}
	return map[string]interface{}{
		"id":             s.ID.String(),
		"fabricId":       s.FabricID.String(),
		"snapshotId":     snapshotID,
		"startEventId":   startEventID,
		"endEventId":     endEventID,
		"status":         s.Status,
		"progress":       s.Progress,
		"eventsReplayed": s.EventsReplayed,
		"startedAt":      s.StartedAt.UTC().Format(time.RFC3339),
		"completedAt":    completedAt,
		"error":          errMsg,
	}
}
