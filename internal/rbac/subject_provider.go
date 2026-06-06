package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/postgres"
	"gorm.io/gorm"
)

type SubjectProvider interface {
	GetCapabilities(ctx context.Context, subjectID string) ([]string, error)
	GetAllSubjects(ctx context.Context) ([]string, error)
}

var (
	ErrCapabilitiesUnavailable = errors.New("capabilities unavailable")
)
