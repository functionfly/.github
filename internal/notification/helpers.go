package notification

import (
	"fmt"

	"github.com/google/uuid"
)

func parseFlexibleUserID(userID interface{}) (uuid.UUID, error) {
	if uid, ok := userID.(uuid.UUID); ok {
		return uid, nil
	}
	if idStr, ok := userID.(string); ok {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
		}
		return parsed, nil
	}
	return uuid.Nil, fmt.Errorf("userID must be uuid.UUID or string")
}
