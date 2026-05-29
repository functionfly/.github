package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRequireFunctionOwner_Forbidden(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	ownerID := uuid.New()
	otherID := uuid.New()
	fnID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "author", "name", "latest_version", "title", "description", "category", "tags",
		"visibility", "price_per_call", "popularity_score", "reliability_score", "deterministic_score",
		"capabilities", "embed_config", "settings", "tenant_id", "owner_user_id",
		"platform_fee_paid", "platform_fee_amount_usd", "last_fee_charged_at", "created_at", "updated_at",
		"trust_score", "trust_tier", "trust_updated_at", "trust_calculation_version",
		"providers", "region", "code", "env_vars", "schedule", "playground_enabled", "playground_config",
		"status", "app_id",
	}).AddRow(
		fnID, "alice", "demo", nil, nil, nil, nil, []byte(`[]`),
		"public", 0, 0, 0, 0,
		nil, nil, nil, nil, ownerID,
		false, 0, nil, nil, nil,
		0, "untrusted", nil, 0,
		nil, "", "", nil, nil, false, nil,
		"draft", nil,
	)

	mock.ExpectQuery(`SELECT .* FROM "registry_functions"`).
		WithArgs(fnID, 1).
		WillReturnRows(rows)

	repo := versioning.NewRepository(db)
	regRepo := storageregistry.NewRegistryRepository(gormDB, nil)
	h := NewHandler(repo, regRepo)

	router := mux.NewRouter()
	router.HandleFunc("/functions/{functionId}/versions/{version}/publish", h.HandlePublishVersion).Methods("POST")

	claims := &auth.Claims{UserID: otherID, Email: "other@example.com", Username: "other"}
	req := httptest.NewRequest("POST", "/functions/"+fnID.String()+"/versions/1.0.0/publish", nil)
	req = middleware.SetUserInContext(req, claims)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
