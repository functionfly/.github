package wasm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseFabricIDFromInput(t *testing.T) {
	id := uuid.New()
	input := []byte(`{"state_fabric_id":"` + id.String() + `"}`)
	assert.Equal(t, id, ParseFabricIDFromInput(input))
	assert.Equal(t, uuid.Nil, ParseFabricIDFromInput([]byte(`{}`)))
	assert.Equal(t, uuid.Nil, ParseFabricIDFromInput(nil))
}

func TestDelegatingHostHandler_SwapDelegate(t *testing.T) {
	d := NewDelegatingHostHandler(NewDefaultHostHandler(nil))
	d.SetDelegate(NewDefaultHostHandler(nil))
	assert.NotPanics(t, func() {
		d.Log("ok")
	})
}

func TestHostHandlerForExecution_NoRepo(t *testing.T) {
	h := HostHandlerForExecution(StateFabricHostConfig{})
	_, ok := h.(*DefaultHostHandler)
	assert.True(t, ok)
}
