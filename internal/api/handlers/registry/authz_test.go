package registry

import (
	"testing"

	"github.com/functionfly/functionfly/internal/auth"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCanViewRegistryFunction(t *testing.T) {
	ownerID := uuid.New()
	tenantID := uuid.New()
	fn := &storageregistry.RegistryFunction{
		Visibility:  "private",
		OwnerUserID: &ownerID,
		TenantID:    &tenantID,
	}

	assert.False(t, CanViewRegistryFunction(fn, nil))

	other := &auth.Claims{UserID: uuid.New(), TenantID: tenantID}
	assert.True(t, CanViewRegistryFunction(fn, other))

	stranger := &auth.Claims{UserID: uuid.New(), TenantID: uuid.New()}
	assert.False(t, CanViewRegistryFunction(fn, stranger))

	publicFn := &storageregistry.RegistryFunction{Visibility: "public"}
	assert.True(t, CanViewRegistryFunction(publicFn, nil))
}

func TestIsRegistryFunctionOwner(t *testing.T) {
	ownerID := uuid.New()
	fn := &storageregistry.RegistryFunction{OwnerUserID: &ownerID}

	owner := &auth.Claims{UserID: ownerID}
	assert.True(t, IsRegistryFunctionOwner(fn, owner))

	other := &auth.Claims{UserID: uuid.New()}
	assert.False(t, IsRegistryFunctionOwner(fn, other))

	admin := &auth.Claims{UserID: uuid.New(), Role: "admin"}
	assert.True(t, IsRegistryFunctionOwner(fn, admin))
}

func TestAuthorMatchesPublisher(t *testing.T) {
	user := &auth.Claims{UserID: uuid.New(), Username: "alice", Email: "alice@example.com"}
	assert.True(t, AuthorMatchesPublisher(user, "alice"))
	assert.False(t, AuthorMatchesPublisher(user, "functionfly"))
	assert.False(t, AuthorMatchesPublisher(user, "bob"))

	admin := &auth.Claims{UserID: uuid.New(), Role: "admin", Username: "admin"}
	assert.True(t, AuthorMatchesPublisher(admin, "functionfly"))
}
