package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetFunctionByID_InvalidUUID(t *testing.T) {
	h := &Handler{}
	router := mux.NewRouter()
	router.HandleFunc("/registry/functions/by-id/{functionId}", h.HandleGetFunctionByID).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/registry/functions/by-id/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
