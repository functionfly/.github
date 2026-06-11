package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/community"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommunityRepository handles community forum persistence.
type CommunityRepository struct {
	db *gorm.DB
}

func escapeLikeWildcards(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// NewCommunityRepository creates a CommunityRepository.
func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

// ListCategories returns all forum categories ordered by sort_order.
func (r *CommunityRepository) ListCategories(ctx context.Context) ([]community.Category, error) {
	var cats []community.Category
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&cats).Error; err != nil {
		return nil, fmt.Errorf("list community categories: %w", err)
	}
	if len(cats) == 0 {
		if err := r.EnsureDefaultCategories(ctx); err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&cats).Error; err != nil {
			return nil, fmt.Errorf("list community categories: %w", err)
		}
	}
	return cats, nil
}

// EnsureDefaultCategories seeds default forum categories when the table is empty.
func (r *CommunityRepository) EnsureDefaultCategories(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&community.Category{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count community categories: %w", err)
	}
	if count > 0 {
		return nil
	}

	defaults := []community.Category{
		{Slug: "getting-started", Name: "Getting Started", Description: "Account setup, first deploy, CLI install, and local dev.", Icon: "rocket", SortOrder: 1},
		{Slug: "functions", Name: "Functions & Runtime", Description: "Deploy, debug, languages, cold starts, and execution issues.", Icon: "code", SortOrder: 2},
		{Slug: "studio", Name: "Studio & Workflows", Description: "FRG editor, AI Composer, Graph Editor, State, and visual workflows.", Icon: "layout", SortOrder: 3},
		{Slug: "agents", Name: "AI Agents", Description: "Agent deployment, wallet, memory, FlyMind, and autonomous workflows.", Icon: "bot", SortOrder: 4},
		{Slug: "integrations", Name: "API & Integrations", Description: "SDK, API keys, webhooks, GitHub import, and third-party providers.", Icon: "plug", SortOrder: 5},
		{Slug: "security", Name: "Security & Secrets", Description: "Vault, auth, MFA, SSO, credentials, and trust verification.", Icon: "shield", SortOrder: 6},
		{Slug: "billing", Name: "Billing & Account", Description: "Plans, usage, invoices, teams, wallet, and payouts.", Icon: "credit-card", SortOrder: 7},
		{Slug: "marketplace", Name: "Marketplace & Gallery", Description: "Publishing functions, buying from the gallery, and registry.", Icon: "store", SortOrder: 8},
		{Slug: "troubleshooting", Name: "Troubleshooting", Description: "Errors, outages, bug reports, and how to fix common problems.", Icon: "bug", SortOrder: 9},
		{Slug: "showcase", Name: "Show & Tell", Description: "Share what you built, tips, tutorials, and wins with the community.", Icon: "sparkles", SortOrder: 10},
		{Slug: "feedback", Name: "Ideas & Feedback", Description: "Feature requests, product suggestions, and platform improvements.", Icon: "lightbulb", SortOrder: 11},
		{Slug: "general", Name: "General", Description: "Announcements, platform news, and off-topic community chat.", Icon: "message-square", SortOrder: 12},
	}

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoNothing: true,
	}).Create(&defaults).Error; err != nil {
		return fmt.Errorf("seed community categories: %w", err)
	}
	return nil
}

// GetCategoryBySlug returns a category by slug.
func (r *CommunityRepository) GetCategoryBySlug(ctx context.Context, slug string) (*community.Category, error) {
	var cat community.Category
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&cat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &cat, nil
}

type postListRow struct {
	community.Post
	CategorySlug   string
	CategoryName   string
	AuthorUsername *string
	AuthorName     string
}

// ListPostsOptions filters and sorts community threads.
type ListPostsOptions struct {
	CategorySlug string
	Sort         string // hot, new, top
	Query        string
	Limit        int
	Offset       int
	ViewerID     *uuid.UUID
}

// ListPosts returns paginated threads with author and category info.
func (r *CommunityRepository) ListPosts(ctx context.Context, opts ListPostsOptions) ([]community.PostListItem, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	q := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name`).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id")

	if opts.CategorySlug != "" {
		q = q.Where("c.slug = ?", opts.CategorySlug)
	}
	if opts.Query != "" {
		like := "%" + escapeLikeWildcards(strings.ToLower(opts.Query)) + "%"
		q = q.Where("(LOWER(p.title) LIKE ? ESCAPE '\' OR LOWER(p.body) LIKE ? ESCAPE '\')", like, like)
	}

	switch opts.Sort {
	case "top":
		q = q.Order("p.is_pinned DESC, p.vote_score DESC, p.last_activity_at DESC")
	case "new":
		q = q.Order("p.is_pinned DESC, p.created_at DESC")
	default: // hot
		q = q.Order("p.is_pinned DESC, (p.vote_score + p.reply_count * 2)::float / GREATEST(EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600, 1) DESC, p.last_activity_at DESC")
	}

	var rows []postListRow
	if err := q.Limit(opts.Limit).Offset(opts.Offset).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list community posts: %w", err)
	}

	out := make([]community.PostListItem, len(rows))
	postIDs := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		postIDs[i] = row.ID
		out[i] = community.PostListItem{
			Post:         row.Post,
			CategorySlug: row.CategorySlug,
			CategoryName: row.CategoryName,
			Author:       authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName),
		}
	}

	if opts.ViewerID != nil && len(postIDs) > 0 {
		votes, err := r.votesForTargets(ctx, *opts.ViewerID, community.VoteTargetPost, postIDs)
		if err != nil {
			return nil, err
		}
		for i := range out {
			if v, ok := votes[out[i].ID]; ok {
				out[i].UserVote = &v
			}
		}
	}

	return out, nil
}

// GetPostByID returns a thread with category and author.
func (r *CommunityRepository) GetPostByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*community.PostDetail, error) {
	type row struct {
		community.Post
		CategorySlug        string
		CategoryName        string
		CategoryDescription string
		CategoryIcon        string
		CategorySortOrder   int
		AuthorUsername      *string
		AuthorName          string
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name, c.description AS category_description,
			c.icon AS category_icon, c.sort_order AS category_sort_order,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name`).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id").
		Where("p.id = ?", id).
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("get community post: %w", err)
	}
	if result.ID == uuid.Nil {
		return nil, nil
	}

	detail := &community.PostDetail{
		Post: result.Post,
		Category: community.Category{
			ID:          result.CategoryID,
			Slug:        result.CategorySlug,
			Name:        result.CategoryName,
			Description: result.CategoryDescription,
			Icon:        result.CategoryIcon,
			SortOrder:   result.CategorySortOrder,
		},
		Author: authorFromRow(result.AuthorID, result.AuthorUsername, result.AuthorName),
	}

	if viewerID != nil {
		votes, err := r.votesForTargets(ctx, *viewerID, community.VoteTargetPost, []uuid.UUID{id})
		if err != nil {
			return nil, err
		}
		if v, ok := votes[id]; ok {
			detail.UserVote = &v
		}
	}

	return detail, nil
}

// CreatePost inserts a new thread.
func (r *CommunityRepository) CreatePost(ctx context.Context, post *community.Post) error {
	now := time.Now().UTC()
	post.CreatedAt = now
	post.UpdatedAt = now
	post.LastActivityAt = now
	if post.Status == "" {
		post.Status = community.StatusOpen
	}
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return fmt.Errorf("create community post: %w", err)
	}
	return nil
}

// IncrementPostViews bumps view count.
func (r *CommunityRepository) IncrementPostViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&community.Post{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

type commentRow struct {
	community.Comment
	AuthorUsername *string
	AuthorName     string
}

// ListCommentsForPost returns all comments on a thread.
func (r *CommunityRepository) ListCommentsForPost(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID) ([]community.CommentWithAuthor, error) {
	var rows []commentRow
	err := r.db.WithContext(ctx).
		Table("community_comments AS cm").
		Select(`cm.*, u.username AS author_username,
			COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name`).
		Joins("JOIN users u ON u.id = cm.author_id").
		Where("cm.post_id = ?", postID).
		Order("cm.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list community comments: %w", err)
	}

	out := make([]community.CommentWithAuthor, len(rows))
	commentIDs := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		commentIDs[i] = row.ID
		out[i] = community.CommentWithAuthor{
			Comment: row.Comment,
			Author:  authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName),
		}
	}

	if viewerID != nil && len(commentIDs) > 0 {
		votes, err := r.votesForTargets(ctx, *viewerID, community.VoteTargetComment, commentIDs)
		if err != nil {
			return nil, err
		}
		for i := range out {
			if v, ok := votes[out[i].ID]; ok {
				out[i].UserVote = &v
			}
		}
	}

	return out, nil
}

// CreateComment adds a reply and updates post counters.
func (r *CommunityRepository) CreateComment(ctx context.Context, comment *community.Comment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		comment.CreatedAt = now
		comment.UpdatedAt = now
		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("create comment: %w", err)
		}
		if err := tx.Model(&community.Post{}).
			Where("id = ?", comment.PostID).
			Updates(map[string]interface{}{
				"reply_count":      gorm.Expr("reply_count + 1"),
				"last_activity_at": now,
				"updated_at":       now,
			}).Error; err != nil {
			return fmt.Errorf("update post reply count: %w", err)
		}
		return nil
	})
}

// GetCommentByID returns a comment by ID.
func (r *CommunityRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*community.Comment, error) {
	var c community.Comment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get comment: %w", err)
	}
	return &c, nil
}

// UpsertVote sets or changes a vote and adjusts target score.
func (r *CommunityRepository) UpsertVote(ctx context.Context, userID uuid.UUID, targetType community.VoteTargetType, targetID uuid.UUID, value int) (int, error) {
	if value != 1 && value != -1 {
		return 0, fmt.Errorf("invalid vote value")
	}

	var newScore int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing community.Vote
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
			First(&existing).Error

		delta := value
		if findErr == nil {
			if existing.Value == value {
				delta = 0
			} else {
				delta = value - existing.Value
				existing.Value = value
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}
		} else if findErr == gorm.ErrRecordNotFound {
			v := community.Vote{
				UserID:     userID,
				TargetType: targetType,
				TargetID:   targetID,
				Value:      value,
				CreatedAt:  time.Now().UTC(),
			}
			if err := tx.Create(&v).Error; err != nil {
				return err
			}
		} else {
			return findErr
		}

		if delta == 0 {
			return r.readVoteScore(tx, targetType, targetID, &newScore)
		}

		table := "community_posts"
		if targetType == community.VoteTargetComment {
			table = "community_comments"
		}
		if err := tx.Exec(
			fmt.Sprintf("UPDATE %s SET vote_score = vote_score + ? WHERE id = ?", table),
			delta, targetID,
		).Error; err != nil {
			return err
		}
		return r.readVoteScore(tx, targetType, targetID, &newScore)
	})
	return newScore, err
}

func (r *CommunityRepository) readVoteScore(tx *gorm.DB, targetType community.VoteTargetType, targetID uuid.UUID, score *int) error {
	if targetType == community.VoteTargetPost {
		var p community.Post
		if err := tx.Select("vote_score").Where("id = ?", targetID).First(&p).Error; err != nil {
			return err
		}
		*score = p.VoteScore
		return nil
	}
	var c community.Comment
	if err := tx.Select("vote_score").Where("id = ?", targetID).First(&c).Error; err != nil {
		return err
	}
	*score = c.VoteScore
	return nil
}

// AcceptComment marks a reply as the solution.
func (r *CommunityRepository) AcceptComment(ctx context.Context, postID, commentID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&community.Comment{}).
			Where("post_id = ?", postID).
			Update("is_accepted", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&community.Comment{}).
			Where("id = ? AND post_id = ?", commentID, postID).
			Updates(map[string]interface{}{"is_accepted": true, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&community.Post{}).
			Where("id = ?", postID).
			Updates(map[string]interface{}{
				"status":              community.StatusSolved,
				"accepted_comment_id": commentID,
				"updated_at":          now,
			}).Error
	})
}

func (r *CommunityRepository) votesForTargets(ctx context.Context, userID uuid.UUID, targetType community.VoteTargetType, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	var votes []community.Vote
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, targetType, ids).
		Find(&votes).Error; err != nil {
		return nil, fmt.Errorf("load user votes: %w", err)
	}
	out := make(map[uuid.UUID]int, len(votes))
	for _, v := range votes {
		out[v.TargetID] = v.Value
	}
	return out, nil
}

func authorFromRow(id uuid.UUID, username *string, name string) community.AuthorSummary {
	a := community.AuthorSummary{ID: id, Name: name}
	if username != nil {
		a.Username = *username
	}
	return a
}
