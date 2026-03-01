package storage

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// UserFollow represents a user-to-user follow relationship
type UserFollow struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FollowerID           uuid.UUID  `json:"follower_id" gorm:"type:uuid;not null;index"`
	FollowedUserID       uuid.UUID  `json:"followed_user_id" gorm:"type:uuid;not null;index"`
	FollowReason         *string    `json:"follow_reason,omitempty" gorm:"size:255"`
	NotifyOnNewFunction  bool       `json:"notify_on_new_function" gorm:"default:true"`
	NotifyOnFunctionUpdate bool      `json:"notify_on_function_update" gorm:"default:true"`
	NotifyOnNewVersion   bool       `json:"notify_on_new_version" gorm:"default:true"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Populated relationships
	Follower     *User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	FollowedUser *User `json:"followed_user,omitempty" gorm:"foreignKey:FollowedUserID"`
}

// FunctionFollow represents a user-to-function follow relationship
type FunctionFollow struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID              uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	FunctionID          uuid.UUID  `json:"function_id" gorm:"type:uuid;not null;index"`
	FollowReason        *string    `json:"follow_reason,omitempty" gorm:"size:255"`
	NotifyOnNewVersion  bool       `json:"notify_on_new_version" gorm:"default:true"`
	NotifyOnRatingChange bool      `json:"notify_on_rating_change" gorm:"default:false"`
	NotifyOnTrustChange bool       `json:"notify_on_trust_change" gorm:"default:true"`
	NotifyOnVerification bool      `json:"notify_on_verification" gorm:"default:true"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Populated relationships
	User     *User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Function *registry.RegistryFunction  `json:"function,omitempty" gorm:"foreignKey:FunctionID"`
}

// UserFollowStats represents cached follower/following counts for a user
type UserFollowStats struct {
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	FollowersCount int       `json:"followers_count" gorm:"default:0"`
	FollowingCount int       `json:"following_count" gorm:"default:0"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// FunctionFollowStats represents cached follower count for a function
type FunctionFollowStats struct {
	FunctionID      uuid.UUID `json:"function_id" gorm:"type:uuid;primaryKey"`
	FollowersCount  int       `json:"followers_count" gorm:"default:0"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
