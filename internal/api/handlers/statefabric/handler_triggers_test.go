package statefabric

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandleListTriggers_Unauthorized(t *testing.T) {
	handler := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/state-fabrics/"+uuid.New().String()+"/triggers", nil)
	w := httptest.NewRecorder()

	handler.HandleListTriggers(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleCreateTrigger_InvalidBody(t *testing.T) {
	handler := &Handler{}
	fabricID := uuid.New()
	req := createAuthRequest(t, http.MethodPost, "/v1/state-fabrics/"+fabricID.String()+"/triggers", []byte("not-json"))
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})
	w := httptest.NewRecorder()

	handler.HandleCreateTrigger(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateTrigger_MissingTarget(t *testing.T) {
	handler := &Handler{}
	fabricID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"triggerType":             "on_update",
		"includePrevious":         true,
		"includeNew":              true,
		"maxInvocationsPerMinute": 60,
		"isActive":                true,
	})
	req := createAuthRequest(t, http.MethodPost, "/v1/state-fabrics/"+fabricID.String()+"/triggers", body)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})
	w := httptest.NewRecorder()

	handler.HandleCreateTrigger(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "targetFunction")
}

func TestHandleDeleteTrigger_InvalidTriggerID(t *testing.T) {
	handler := &Handler{}
	fabricID := uuid.New()
	req := createAuthRequest(t, http.MethodDelete, "/v1/state-fabrics/"+fabricID.String()+"/triggers/bad-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String(), "triggerId": "bad-id"})
	w := httptest.NewRecorder()

	handler.HandleDeleteTrigger(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateFabricTriggerRequest_JSON(t *testing.T) {
	payload := createFabricTriggerRequest{
		TriggerType:             "on_update",
		TargetFunction:          "my-fn",
		IncludePrevious:         true,
		IncludeNew:              true,
		MaxInvocationsPerMinute: 60,
		IsActive:                true,
	}
	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded createFabricTriggerRequest
	assert.NoError(t, json.NewDecoder(bytes.NewReader(data)).Decode(&decoded))
	assert.Equal(t, "my-fn", decoded.TargetFunction)
}
