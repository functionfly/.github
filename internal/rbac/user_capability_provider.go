package rbac

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserCapabilityProvider struct {
	env              storage.Environment
	cacheEnabled     bool
	cacheTTL         time.Duration
	fallbackToRoleCap bool
}

func NewUserCapabilityProvider(env storage.Environment, fallbackToRoleCap bool) *UserCapabilityProvider {
	return &UserCapabilityProvider{
		env:              env,
		cacheEnabled:     true,
		cacheTTL:         5 * time.Minute,
		fallbackToRoleCap: fallbackToRoleCap,
	}
}

func (p *UserCapabilityProvider) GetCapabilities(ctx context.Context, subjectID string) ([]string, error) {
	caps, err := p.env.GetCapabilities(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("GetCapabilities(%q): %w", subjectID, err)
	}
	return caps, nil
}

func (p *UserCapabilityProvider) GetAllSubjects(ctx context.Context) ([]string, error) {
	users, err := p.env.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.ID)
	}
	return out, nil
}
