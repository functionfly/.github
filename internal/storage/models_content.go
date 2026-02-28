package storage

import (
	"time"

	"github.com/google/uuid"
)

// ChangelogEntry represents a changelog entry
type ChangelogEntry struct {
	ID          uuid.UUID         `json:"id"`
	Version     string            `json:"version"`
	Date        time.Time         `json:"date"`
	Type        string            `json:"type"` // "major", "minor", "patch"
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Changes     []ChangelogChange `json:"changes"`
	ReleaseURL  *string           `json:"release_url,omitempty"`
	GitHubID    *string           `json:"github_id,omitempty"`
	IsPublished bool              `json:"is_published"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ChangelogChange represents a category of changes in a changelog entry
type ChangelogChange struct {
	ID        uuid.UUID `json:"id"`
	EntryID   uuid.UUID `json:"entry_id"`
	Category  string    `json:"category"`
	Icon      string    `json:"icon"` // Icon name from lucide-react
	Items     []string  `json:"items"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BlogPost represents a blog post
type BlogPost struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Content       string     `json:"content"`
	Excerpt       string     `json:"excerpt"`
	Author        string     `json:"author"`
	Tags          []string   `json:"tags"`
	FeaturedImage *string    `json:"featured_image,omitempty"`
	SanityID      *string    `json:"sanity_id,omitempty"`
	IsPublished   bool       `json:"is_published"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Feedback represents a user feedback submission
type Feedback struct {
	ID           uuid.UUID            `json:"id"`
	UserID       *uuid.UUID           `json:"user_id,omitempty"`    // Anonymous users can submit feedback
	UserEmail    *string              `json:"user_email,omitempty"` // For anonymous feedback
	FeedbackType string               `json:"feedback_type"`        // "bug", "feature", "improvement", "general"
	Subject      string               `json:"subject"`
	Message      string               `json:"message"`
	Priority     string               `json:"priority"`               // "low", "medium", "high", "critical"
	BrowserInfo  string               `json:"browser_info,omitempty"` // Browser/OS information
	Status       string               `json:"status"`                 // "submitted", "in-review", "resolved", "closed"
	IPAddress    string               `json:"ip_address,omitempty"`
	UserAgent    string               `json:"user_agent,omitempty"`
	Attachments  []FeedbackAttachment `json:"attachments,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// FeedbackAttachment represents a file attachment for feedback
type FeedbackAttachment struct {
	ID          uuid.UUID `json:"id"`
	FeedbackID  uuid.UUID `json:"feedback_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	S3Key       string    `json:"s3_key"` // S3 object key for the file
	S3Bucket    string    `json:"s3_bucket"`
	CreatedAt   time.Time `json:"created_at"`
}
