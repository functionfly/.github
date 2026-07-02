package health

import (
	"context"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/circuitbreaker"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// dbPersistence implements circuitbreaker.Persistence using the storage repository.
// Keys are in the format "backend:<uuid>" — the UUID is parsed and used to
// load/save CircuitState records.
type dbPersistence struct {
	repo storage.Repository
}

// NewDBPersistence creates a new DB-backed persistence for circuit breaker state.
func NewDBPersistence(repo storage.Repository) circuitbreaker.Persistence {
	return &dbPersistence{repo: repo}
}

func (p *dbPersistence) Load(ctx context.Context, key string) (*circuitbreaker.StoredState, error) {
	backendID, ok := parseBackendKey(key)
	if !ok {
		return nil, nil // not a backend key, skip
	}

	state, err := p.repo.GetCircuitState(ctx, backendID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
	}

	return &circuitbreaker.StoredState{
		State:       stateToInt(state.State),
		FailCount:   state.FailCount,
		ReopenCount: 0, // not tracked in DB yet; future migration can add this
		Since:       state.SinceTs,
		LastFailure: derefTime(state.LastFailureTs),
	}, nil
}

func (p *dbPersistence) Save(ctx context.Context, key string, state *circuitbreaker.StoredState) error {
	backendID, ok := parseBackendKey(key)
	if !ok {
		return nil // not a backend key, skip
	}

	cs := &storage.CircuitState{
		BackendID: backendID,
		State:     intToState(state.State),
		SinceTs:   state.Since,
		FailCount: state.FailCount,
	}

	if !state.LastFailure.IsZero() {
		cs.LastFailureTs = &state.LastFailure
	}

	return p.repo.UpsertCircuitState(ctx, cs)
}

func parseBackendKey(key string) (uuid.UUID, bool) {
	if !strings.HasPrefix(key, "backend:") {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(strings.TrimPrefix(key, "backend:"))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func stateToInt(s string) int {
	switch s {
	case "open":
		return 1
	case "half-open":
		return 2
	default:
		return 0
	}
}

func intToState(i int) string {
	switch i {
	case 1:
		return "open"
	case 2:
		return "half-open"
	default:
		return "closed"
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
