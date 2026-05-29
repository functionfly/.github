// Package community provides models for the public help forum.
package community

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus represents thread state.
type PostStatus string

const (
	StatusOpen   PostStatus = "open"
	StatusSolved PostStatus = "solved"
	StatusLocked PostStatus = "locked"
)

// VoteTargetType is post or comment.
type VoteTargetType string

const (
	VoteTargetPost    VoteTargetType = "post"
	VoteTargetComment VoteTargetType = "comment"
)

// Category is a forum section.
type Category struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug        string    `json:"slug" gorm:"size:64;not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	Description string    `json:"description" gorm:"type:text;not null;default:''"`
	Icon        string    `json:"icon" gorm:"size:32;not null;default:'help-circle'"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:now()"`
}

func (Category) TableName() string { return "community_categories" }

// Post is a community thread.
type Post struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CategoryID         uuid.UUID  `json:"category_id" gorm:"type:uuid;not null"`
	AuthorID           uuid.UUID  `json:"author_id" gorm:"type:uuid;not null"`
	Title              string     `json:"title" gorm:"size:300;not null"`
	Body               string     `json:"body" gorm:"type:text;not null"`
	Status             PostStatus `json:"status" gorm:"size:20;not null;default:'open'"`
	VoteScore          int        `json:"vote_score" gorm:"not null;default:0"`
	ReplyCount         int        `json:"reply_count" gorm:"not null;default:0"`
	ViewCount          int        `json:"view_count" gorm:"not null;default:0"`
	Tags               []string   `json:"tags" gorm:"type:text[];not null;default:'{}'"`
	IsPinned           bool       `json:"is_pinned" gorm:"not null;default:false"`
	AcceptedCommentID  *uuid.UUID `json:"accepted_comment_id,omitempty" gorm:"type:uuid"`
	CreatedAt          time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"not null;default:now()"`
	LastActivityAt     time.Time  `json:"last_activity_at" gorm:"not null;default:now()"`
}

func (Post) TableName() string { return "community_posts" }

// Comment is a reply on a thread.
type Comment struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID     uuid.UUID  `json:"post_id" gorm:"type:uuid;not null"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	AuthorID   uuid.UUID  `json:"author_id" gorm:"type:uuid;not null"`
	Body       string     `json:"body" gorm:"type:text;not null"`
	VoteScore  int        `json:"vote_score" gorm:"not null;default:0"`
	IsAccepted bool       `json:"is_accepted" gorm:"not null;default:false"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt  time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"not null;default:now()"`
}

func (Comment) TableName() string { return "community_comments" }

// Vote records a user's vote on a post or comment.
type Vote struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	TargetType VoteTargetType `json:"target_type" gorm:"size:16;not null"`
	TargetID   uuid.UUID      `json:"target_id" gorm:"type:uuid;not null"`
	Value      int            `json:"value" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at" gorm:"not null;default:now()"`
}

func (Vote) TableName() string { return "community_votes" }

// AuthorSummary is denormalized author info for API responses.
type AuthorSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username,omitempty"`
	Name     string    `json:"name,omitempty"`
}

// PostListItem extends Post with category and author for list views.
type PostListItem struct {
	Post
	CategorySlug string        `json:"category_slug"`
	CategoryName string        `json:"category_name"`
	Author       AuthorSummary `json:"author"`
	UserVote     *int          `json:"user_vote,omitempty"`
}

// PostDetail extends Post with full relations for thread view.
type PostDetail struct {
	Post
	Category Category      `json:"category"`
	Author   AuthorSummary `json:"author"`
	UserVote *int          `json:"user_vote,omitempty"`
}

// CommentWithAuthor includes author info for display.
type CommentWithAuthor struct {
	Comment
	Author   AuthorSummary `json:"author"`
	UserVote *int          `json:"user_vote,omitempty"`
}
