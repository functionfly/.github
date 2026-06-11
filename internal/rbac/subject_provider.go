package rbac

import (
	"context"
	"errors"
)

type SubjectProvider interface {
	GetCapabilities(ctx context.Context, subjectID string) ([]string, error)
	GetAllSubjects(ctx context.Context) ([]string, error)
}

var (
	ErrCapabilitiesUnavailable = errors.New("capabilities unavailable")
)
