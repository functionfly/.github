package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/community"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	CategorySlug    string
	CategoryName    string
	AuthorUsername  *string
	AuthorName      string
	AuthorAvatarURL *string
}

// ListPostsOptions filters and sorts community threads.
type ListPostsOptions struct {
	CategorySlug string
	Sort         string // hot, new, top
	Query        string
	TagFilter    string
	AuthorID     *uuid.UUID
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
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id")

	if opts.CategorySlug != "" {
		q = q.Where("c.slug = ?", opts.CategorySlug)
	}
	if opts.Query != "" {
		like := "%" + escapeLikeWildcards(strings.ToLower(opts.Query)) + "%"
		q = q.Where("(LOWER(p.title) LIKE ? ESCAPE '\\' OR LOWER(p.body) LIKE ? ESCAPE '\\')", like, like)
	}
	if opts.AuthorID != nil {
		q = q.Where("p.author_id = ?", *opts.AuthorID)
	}
	if opts.TagFilter != "" {
		q = q.Where("? = ANY(p.tags)", opts.TagFilter)
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
			Author:       authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName, row.AuthorAvatarURL),
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
		AuthorAvatarURL     *string
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name, c.description AS category_description,
			c.icon AS category_icon, c.sort_order AS category_sort_order,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
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
		Author: authorFromRow(result.AuthorID, result.AuthorUsername, result.AuthorName, result.AuthorAvatarURL),
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
	if post.Tags == nil {
		post.Tags = pq.StringArray{}
	}
	if post.Slug == "" {
		post.Slug = community.GenerateSlug(post.Title, false)
	}
	// Retry on slug collision with random suffix
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Create(post).Error
		if err == nil {
			return nil
		}
		if !isUniqueViolation(err) {
			return fmt.Errorf("create community post: %w", err)
		}
		post.Slug = community.GenerateSlug(post.Title, true)
	}
	return fmt.Errorf("create community post: slug collision after retries")
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
	AuthorUsername  *string
	AuthorName      string
	AuthorAvatarURL *string
}

// ListCommentsForPost returns all comments on a thread.
func (r *CommunityRepository) ListCommentsForPost(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID) ([]community.CommentWithAuthor, error) {
	var rows []commentRow
	err := r.db.WithContext(ctx).
		Table("community_comments AS cm").
		Select(`cm.*, u.username AS author_username,
			COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
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
			Author:  authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName, row.AuthorAvatarURL),
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

func authorFromRow(id uuid.UUID, username *string, name string, avatarURL *string) community.AuthorSummary {
	a := community.AuthorSummary{ID: id, Name: name}
	if username != nil {
		a.Username = *username
	}
	if avatarURL != nil {
		a.AvatarURL = *avatarURL
	}
	return a
}

// GetPostBySlug returns a thread by its URL slug.
func (r *CommunityRepository) GetPostBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*community.PostDetail, error) {
	type row struct {
		community.Post
		CategorySlug        string
		CategoryName        string
		CategoryDescription string
		CategoryIcon        string
		CategorySortOrder   int
		AuthorUsername      *string
		AuthorName          string
		AuthorAvatarURL     *string
	}
	var result row
	err := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name, c.description AS category_description,
			c.icon AS category_icon, c.sort_order AS category_sort_order,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id").
		Where("p.slug = ?", slug).
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("get community post by slug: %w", err)
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
		Author: authorFromRow(result.AuthorID, result.AuthorUsername, result.AuthorName, result.AuthorAvatarURL),
	}

	if viewerID != nil {
		votes, err := r.votesForTargets(ctx, *viewerID, community.VoteTargetPost, []uuid.UUID{result.ID})
		if err != nil {
			return nil, err
		}
		if v, ok := votes[result.ID]; ok {
			detail.UserVote = &v
		}
	}

	return detail, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}

// CountPosts returns the total number of posts matching the given filters.
func (r *CommunityRepository) CountPosts(ctx context.Context, opts ListPostsOptions) (int, error) {
	q := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Joins("JOIN community_categories c ON c.id = p.category_id")

	if opts.CategorySlug != "" {
		q = q.Where("c.slug = ?", opts.CategorySlug)
	}
	if opts.Query != "" {
		like := "%" + escapeLikeWildcards(strings.ToLower(opts.Query)) + "%"
		q = q.Where("(LOWER(p.title) LIKE ? ESCAPE '\\' OR LOWER(p.body) LIKE ? ESCAPE '\\')", like, like)
	}
	if opts.AuthorID != nil {
		q = q.Where("p.author_id = ?", *opts.AuthorID)
	}
	if opts.TagFilter != "" {
		q = q.Where("? = ANY(p.tags)", opts.TagFilter)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count community posts: %w", err)
	}
	return int(count), nil
}

// ListCategoriesWithCounts returns all categories with their post counts.
func (r *CommunityRepository) ListCategoriesWithCounts(ctx context.Context) ([]community.CategoryWithCount, error) {
	type row struct {
		community.Category
		PostCount int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("community_categories AS c").
		Select("c.*, COALESCE(cnt.cnt, 0) AS post_count").
		Joins("LEFT JOIN (SELECT category_id, COUNT(*) AS cnt FROM community_posts GROUP BY category_id) cnt ON cnt.category_id = c.id").
		Order("c.sort_order ASC, c.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list categories with counts: %w", err)
	}
	if len(rows) == 0 {
		if err := r.EnsureDefaultCategories(ctx); err != nil {
			return nil, err
		}
		return r.ListCategoriesWithCounts(ctx)
	}
	out := make([]community.CategoryWithCount, len(rows))
	for i, rw := range rows {
		out[i] = community.CategoryWithCount{Category: rw.Category, PostCount: rw.PostCount}
	}
	return out, nil
}

// ListPostsByAuthor returns posts by a specific author for profile pages.
func (r *CommunityRepository) ListPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]community.PostListItem, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	r.db.WithContext(ctx).Model(&community.Post{}).Where("author_id = ?", authorID).Count(&total)

	var rows []postListRow
	err := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id").
		Where("p.author_id = ?", authorID).
		Order("p.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list posts by author: %w", err)
	}

	out := make([]community.PostListItem, len(rows))
	for i, row := range rows {
		out[i] = community.PostListItem{
			Post:         row.Post,
			CategorySlug: row.CategorySlug,
			CategoryName: row.CategoryName,
			Author:       authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName, row.AuthorAvatarURL),
		}
	}
	return out, int(total), nil
}

// UpdatePost updates a post's title, body, and tags. Only the author may do this.
func (r *CommunityRepository) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, title, body string, tags []string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&community.Post{}).
		Where("id = ? AND author_id = ?", postID, authorID).
		Updates(map[string]interface{}{
			"title":      title,
			"body":       body,
			"tags":       tags,
			"updated_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("update post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found or not owned by user")
	}
	return nil
}

// DeletePost soft-deletes a post. Only the author may do this.
func (r *CommunityRepository) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND author_id = ?", postID, authorID).
		Delete(&community.Post{})
	if result.Error != nil {
		return fmt.Errorf("delete post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found or not owned by user")
	}
	return nil
}

// UpdateComment updates a comment body. Only the author may do this.
func (r *CommunityRepository) UpdateComment(ctx context.Context, commentID, authorID uuid.UUID, body string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&community.Comment{}).
		Where("id = ? AND author_id = ?", commentID, authorID).
		Updates(map[string]interface{}{
			"body":       body,
			"updated_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("update comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("comment not found or not owned by user")
	}
	return nil
}

// DeleteComment soft-deletes a comment. Only the author may do this.
func (r *CommunityRepository) DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&community.Comment{}).
		Where("id = ? AND author_id = ?", commentID, authorID).
		Updates(map[string]interface{}{
			"deleted_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("delete comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("comment not found or not owned by user")
	}
	return nil
}

// BookmarkPost adds a bookmark. No-op if already bookmarked.
func (r *CommunityRepository) BookmarkPost(ctx context.Context, userID, postID uuid.UUID) error {
	bm := community.Bookmark{
		UserID:    userID,
		PostID:    postID,
		CreatedAt: time.Now().UTC(),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&bm).Error
}

// UnbookmarkPost removes a bookmark.
func (r *CommunityRepository) UnbookmarkPost(ctx context.Context, userID, postID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&community.Bookmark{}).Error
}

// IsBookmarked checks if a user has bookmarked a post.
func (r *CommunityRepository) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&community.Bookmark{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

// ListBookmarks returns bookmarked posts for a user.
func (r *CommunityRepository) ListBookmarks(ctx context.Context, userID uuid.UUID, limit, offset int) ([]community.PostListItem, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	r.db.WithContext(ctx).Model(&community.Bookmark{}).Where("user_id = ?", userID).Count(&total)

	var rows []postListRow
	err := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`p.*, c.slug AS category_slug, c.name AS category_name,
			u.username AS author_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS author_name,
			u.provider_data->>'avatar_url' AS author_avatar_url`).
		Joins("JOIN community_bookmarks b ON b.post_id = p.id AND b.user_id = ?", userID).
		Joins("JOIN community_categories c ON c.id = p.category_id").
		Joins("JOIN users u ON u.id = p.author_id").
		Order("b.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list bookmarks: %w", err)
	}

	out := make([]community.PostListItem, len(rows))
	for i, row := range rows {
		out[i] = community.PostListItem{
			Post:         row.Post,
			CategorySlug: row.CategorySlug,
			CategoryName: row.CategoryName,
			Author:       authorFromRow(row.AuthorID, row.AuthorUsername, row.AuthorName, row.AuthorAvatarURL),
		}
	}
	return out, int(total), nil
}

// CreateNotification inserts a community notification.
func (r *CommunityRepository) CreateNotification(ctx context.Context, n *community.Notification) error {
	n.CreatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Create(n).Error
}

// ListNotifications returns notifications for a user with actor info.
func (r *CommunityRepository) ListNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]community.NotificationWithActor, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	var total int64
	r.db.WithContext(ctx).Model(&community.Notification{}).Where("user_id = ?", userID).Count(&total)

	type notifRow struct {
		community.Notification
		ActorUsername  *string
		ActorName      string
		ActorAvatarURL *string
		PostSlug       string
		PostTitle      string
	}
	var rows []notifRow
	err := r.db.WithContext(ctx).
		Table("community_notifications AS n").
		Select(`n.*,
			u.username AS actor_username, COALESCE(NULLIF(u.name, ''), split_part(u.email, '@', 1)) AS actor_name,
			u.provider_data->>'avatar_url' AS actor_avatar_url,
			COALESCE(p.slug, '') AS post_slug, COALESCE(p.title, '') AS post_title`).
		Joins("JOIN users u ON u.id = n.actor_id").
		Joins("LEFT JOIN community_posts p ON p.id = n.post_id").
		Where("n.user_id = ?", userID).
		Order("n.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	out := make([]community.NotificationWithActor, len(rows))
	for i, rw := range rows {
		out[i] = community.NotificationWithActor{
			Notification: rw.Notification,
			Actor:        authorFromRow(rw.ActorID, rw.ActorUsername, rw.ActorName, rw.ActorAvatarURL),
			PostSlug:     rw.PostSlug,
			PostTitle:    rw.PostTitle,
		}
	}
	return out, int(total), nil
}

// CountUnreadNotifications returns the count of unread notifications.
func (r *CommunityRepository) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&community.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return int(count), err
}

// MarkNotificationsRead marks all notifications as read for a user.
func (r *CommunityRepository) MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&community.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

// ListRules returns all active community rules ordered by sort_order.
func (r *CommunityRepository) ListRules(ctx context.Context) ([]community.Rule, error) {
	var rules []community.Rule
	if err := r.db.WithContext(ctx).Where("is_active = true").Order("sort_order ASC, created_at ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list community rules: %w", err)
	}
	return rules, nil
}

// ListAllRules returns all rules (including inactive) for admin management.
func (r *CommunityRepository) ListAllRules(ctx context.Context) ([]community.Rule, error) {
	var rules []community.Rule
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list all community rules: %w", err)
	}
	return rules, nil
}

// GetRule returns a single rule by ID.
func (r *CommunityRepository) GetRule(ctx context.Context, id uuid.UUID) (*community.Rule, error) {
	var rule community.Rule
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get rule: %w", err)
	}
	return &rule, nil
}

// CreateRule inserts a new community rule.
func (r *CommunityRepository) CreateRule(ctx context.Context, rule *community.Rule) error {
	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	return r.db.WithContext(ctx).Create(rule).Error
}

// UpdateRule updates an existing community rule.
func (r *CommunityRepository) UpdateRule(ctx context.Context, rule *community.Rule) error {
	rule.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Save(rule)
	if result.Error != nil {
		return fmt.Errorf("update rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// DeleteRule removes a community rule by ID.
func (r *CommunityRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&community.Rule{})
	if result.Error != nil {
		return fmt.Errorf("delete rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}
