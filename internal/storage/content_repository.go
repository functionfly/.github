package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ContentRepository handles content management operations
type ContentRepository struct {
	db *PostgresDB
}

// NewContentRepository creates a new content repository
func NewContentRepository(db *PostgresDB) *ContentRepository {
	return &ContentRepository{db: db}
}

// Changelog operations

// CreateChangelogEntry creates a new changelog entry
func (r *ContentRepository) CreateChangelogEntry(ctx context.Context, entry *ChangelogEntry) (*ChangelogEntry, error) {
	query := `
		INSERT INTO changelog_entries (version, date, type, title, description, release_url, github_id, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	result := &ChangelogEntry{}
	err := r.db.QueryRowContext(ctx, query,
		entry.Version, entry.Date, entry.Type, entry.Title, entry.Description,
		entry.ReleaseURL, entry.GitHubID, entry.IsPublished).Scan(
		&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create changelog entry: %w", err)
	}

	// Copy the input values
	result.Version = entry.Version
	result.Date = entry.Date
	result.Type = entry.Type
	result.Title = entry.Title
	result.Description = entry.Description
	result.ReleaseURL = entry.ReleaseURL
	result.GitHubID = entry.GitHubID
	result.IsPublished = entry.IsPublished

	// Load changes if provided
	if len(entry.Changes) > 0 {
		for _, change := range entry.Changes {
			change.EntryID = result.ID
			createdChange, err := r.CreateChangelogChange(ctx, &change)
			if err != nil {
				return nil, fmt.Errorf("failed to create changelog change: %w", err)
			}
			result.Changes = append(result.Changes, *createdChange)
		}
	}

	return result, nil
}

// GetChangelogEntryByID retrieves a changelog entry by ID with its changes
func (r *ContentRepository) GetChangelogEntryByID(id uuid.UUID) (*ChangelogEntry, error) {
	query := `
		SELECT id, version, date, type, title, description, release_url, github_id, is_published, created_at, updated_at
		FROM changelog_entries
		WHERE id = $1`

	entry := &ChangelogEntry{}
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&entry.ID, &entry.Version, &entry.Date, &entry.Type, &entry.Title, &entry.Description,
		&entry.ReleaseURL, &entry.GitHubID, &entry.IsPublished, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("changelog entry not found")
		}
		return nil, fmt.Errorf("failed to get changelog entry: %w", err)
	}

	// Load changes
	changes, err := r.getChangelogChangesByEntryID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load changelog changes: %w", err)
	}
	entry.Changes = changes

	return entry, nil
}

// GetChangelogEntryByVersion retrieves a changelog entry by version
func (r *ContentRepository) GetChangelogEntryByVersion(version string) (*ChangelogEntry, error) {
	query := `
		SELECT id, version, date, type, title, description, release_url, github_id, is_published, created_at, updated_at
		FROM changelog_entries
		WHERE version = $1`

	entry := &ChangelogEntry{}
	err := r.db.QueryRowContext(context.Background(), query, version).Scan(
		&entry.ID, &entry.Version, &entry.Date, &entry.Type, &entry.Title, &entry.Description,
		&entry.ReleaseURL, &entry.GitHubID, &entry.IsPublished, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("changelog entry not found")
		}
		return nil, fmt.Errorf("failed to get changelog entry: %w", err)
	}

	// Load changes
	changes, err := r.getChangelogChangesByEntryID(entry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load changelog changes: %w", err)
	}
	entry.Changes = changes

	return entry, nil
}

// ListChangelogEntries lists changelog entries with pagination
func (r *ContentRepository) ListChangelogEntries(limit, offset int, publishedOnly bool) ([]*ChangelogEntry, error) {
	query := `
		SELECT id, version, date, type, title, description, release_url, github_id, is_published, created_at, updated_at
		FROM changelog_entries`

	args := []interface{}{}
	argCount := 0

	if publishedOnly {
		argCount++
		query += fmt.Sprintf(" WHERE is_published = $%d", argCount)
		args = append(args, true)
	}

	query += " ORDER BY date DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list changelog entries: %w", err)
	}
	defer rows.Close()

	var entries []*ChangelogEntry
	for rows.Next() {
		entry := &ChangelogEntry{}
		err := rows.Scan(
			&entry.ID, &entry.Version, &entry.Date, &entry.Type, &entry.Title, &entry.Description,
			&entry.ReleaseURL, &entry.GitHubID, &entry.IsPublished, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan changelog entry: %w", err)
		}

		// Load changes for this entry
		changes, err := r.getChangelogChangesByEntryID(entry.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load changelog changes for entry %s: %w", entry.ID, err)
		}
		entry.Changes = changes

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// UpdateChangelogEntry updates a changelog entry
func (r *ContentRepository) UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogEntry, error) {
	if len(updates) == 0 {
		return r.GetChangelogEntryByID(id)
	}

	setParts := []string{}
	args := []interface{}{}
	argCount := 0

	for field, value := range updates {
		argCount++
		setParts = append(setParts, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(field), argCount))
		args = append(args, value)
	}

	// Always update updated_at
	argCount++
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE changelog_entries SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argCount+1)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update changelog entry: %w", err)
	}

	return r.GetChangelogEntryByID(id)
}

// DeleteChangelogEntry deletes a changelog entry
func (r *ContentRepository) DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM changelog_entries WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete changelog entry: %w", err)
	}
	return nil
}

// Changelog change operations

// CreateChangelogChange creates a new changelog change
func (r *ContentRepository) CreateChangelogChange(ctx context.Context, change *ChangelogChange) (*ChangelogChange, error) {
	query := `
		INSERT INTO changelog_changes (entry_id, category, icon, items)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	result := &ChangelogChange{}
	err := r.db.QueryRowContext(ctx, query,
		change.EntryID, change.Category, change.Icon, pq.Array(change.Items)).Scan(
		&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create changelog change: %w", err)
	}

	// Copy input values
	result.EntryID = change.EntryID
	result.Category = change.Category
	result.Icon = change.Icon
	result.Items = change.Items

	return result, nil
}

// UpdateChangelogChange updates a changelog change
func (r *ContentRepository) UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogChange, error) {
	if len(updates) == 0 {
		return r.getChangelogChangeByID(id)
	}

	setParts := []string{}
	args := []interface{}{}
	argCount := 0

	for field, value := range updates {
		argCount++
		setParts = append(setParts, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(field), argCount))
		args = append(args, value)
	}

	// Always update updated_at
	argCount++
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE changelog_changes SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argCount+1)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update changelog change: %w", err)
	}

	return r.getChangelogChangeByID(id)
}

// DeleteChangelogChange deletes a changelog change
func (r *ContentRepository) DeleteChangelogChange(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM changelog_changes WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete changelog change: %w", err)
	}
	return nil
}

// Helper methods

func (r *ContentRepository) getChangelogChangesByEntryID(entryID uuid.UUID) ([]ChangelogChange, error) {
	query := `
		SELECT id, entry_id, category, icon, items, created_at, updated_at
		FROM changelog_changes
		WHERE entry_id = $1
		ORDER BY created_at`

	rows, err := r.db.QueryContext(context.Background(), query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query changelog changes: %w", err)
	}
	defer rows.Close()

	var changes []ChangelogChange
	for rows.Next() {
		change := ChangelogChange{}
		err := rows.Scan(
			&change.ID, &change.EntryID, &change.Category, &change.Icon,
			pq.Array(&change.Items), &change.CreatedAt, &change.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan changelog change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, rows.Err()
}

func (r *ContentRepository) getChangelogChangeByID(id uuid.UUID) (*ChangelogChange, error) {
	query := `
		SELECT id, entry_id, category, icon, items, created_at, updated_at
		FROM changelog_changes
		WHERE id = $1`

	change := &ChangelogChange{}
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&change.ID, &change.EntryID, &change.Category, &change.Icon,
		pq.Array(&change.Items), &change.CreatedAt, &change.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("changelog change not found")
		}
		return nil, fmt.Errorf("failed to get changelog change: %w", err)
	}

	return change, nil
}

// Blog operations

// CreateBlogPost creates a new blog post
func (r *ContentRepository) CreateBlogPost(ctx context.Context, post *BlogPost) (*BlogPost, error) {
	query := `
		INSERT INTO blog_posts (title, slug, content, excerpt, author, tags, featured_image, sanity_id, is_published, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	result := &BlogPost{}
	err := r.db.QueryRowContext(ctx, query,
		post.Title, post.Slug, post.Content, post.Excerpt, post.Author,
		pq.Array(post.Tags), post.FeaturedImage, post.SanityID,
		post.IsPublished, post.PublishedAt).Scan(
		&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create blog post: %w", err)
	}

	// Copy input values
	result.Title = post.Title
	result.Slug = post.Slug
	result.Content = post.Content
	result.Excerpt = post.Excerpt
	result.Author = post.Author
	result.Tags = post.Tags
	result.FeaturedImage = post.FeaturedImage
	result.SanityID = post.SanityID
	result.IsPublished = post.IsPublished
	result.PublishedAt = post.PublishedAt

	return result, nil
}

// GetBlogPostByID retrieves a blog post by ID
func (r *ContentRepository) GetBlogPostByID(id uuid.UUID) (*BlogPost, error) {
	query := `
		SELECT id, title, slug, content, excerpt, author, tags, featured_image, sanity_id, is_published, published_at, created_at, updated_at
		FROM blog_posts
		WHERE id = $1`

	post := &BlogPost{}
	var tags pq.StringArray
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Content, &post.Excerpt, &post.Author,
		&tags, &post.FeaturedImage, &post.SanityID, &post.IsPublished, &post.PublishedAt,
		&post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found")
		}
		return nil, fmt.Errorf("failed to get blog post: %w", err)
	}

	post.Tags = []string(tags)
	return post, nil
}

// GetBlogPostBySlug retrieves a blog post by slug
func (r *ContentRepository) GetBlogPostBySlug(slug string) (*BlogPost, error) {
	query := `
		SELECT id, title, slug, content, excerpt, author, tags, featured_image, sanity_id, is_published, published_at, created_at, updated_at
		FROM blog_posts
		WHERE slug = $1`

	post := &BlogPost{}
	var tags pq.StringArray
	err := r.db.QueryRowContext(context.Background(), query, slug).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Content, &post.Excerpt, &post.Author,
		&tags, &post.FeaturedImage, &post.SanityID, &post.IsPublished, &post.PublishedAt,
		&post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found")
		}
		return nil, fmt.Errorf("failed to get blog post: %w", err)
	}

	post.Tags = []string(tags)
	return post, nil
}

// ListBlogPosts lists blog posts with filtering and pagination
func (r *ContentRepository) ListBlogPosts(limit, offset int, publishedOnly bool, tagFilter []string) ([]*BlogPost, error) {
	query := `
		SELECT id, title, slug, content, excerpt, author, tags, featured_image, sanity_id, is_published, published_at, created_at, updated_at
		FROM blog_posts`

	args := []interface{}{}
	argCount := 0
	conditions := []string{}

	if publishedOnly {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_published = $%d", argCount))
		args = append(args, true)
	}

	if len(tagFilter) > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argCount))
		args = append(args, pq.Array(tagFilter))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY published_at DESC NULLS LAST, created_at DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list blog posts: %w", err)
	}
	defer rows.Close()

	var posts []*BlogPost
	for rows.Next() {
		post := &BlogPost{}
		var tags pq.StringArray
		err := rows.Scan(
			&post.ID, &post.Title, &post.Slug, &post.Content, &post.Excerpt, &post.Author,
			&tags, &post.FeaturedImage, &post.SanityID, &post.IsPublished, &post.PublishedAt,
			&post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blog post: %w", err)
		}
		post.Tags = []string(tags)
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// UpdateBlogPost updates a blog post
func (r *ContentRepository) UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogPost, error) {
	if len(updates) == 0 {
		return r.GetBlogPostByID(id)
	}

	setParts := []string{}
	args := []interface{}{}
	argCount := 0

	for field, value := range updates {
		argCount++
		setParts = append(setParts, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(field), argCount))
		args = append(args, value)
	}

	// Always update updated_at
	argCount++
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE blog_posts SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argCount+1)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog post: %w", err)
	}

	return r.GetBlogPostByID(id)
}

// DeleteBlogPost deletes a blog post
func (r *ContentRepository) DeleteBlogPost(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM blog_posts WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete blog post: %w", err)
	}
	return nil
}

// ListBlogCategories returns all blog categories ordered by order, title
func (r *ContentRepository) ListBlogCategories(ctx context.Context) ([]*BlogCategory, error) {
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at FROM blog_categories ORDER BY "order" ASC, title ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list blog categories: %w", err)
	}
	defer rows.Close()
	var list []*BlogCategory
	for rows.Next() {
		c := &BlogCategory{}
		var desc, color, icon sql.NullString
		if err := rows.Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan blog category: %w", err)
		}
		if desc.Valid {
			c.Description = desc.String
		}
		if color.Valid {
			c.Color = color.String
		}
		if icon.Valid {
			c.Icon = icon.String
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// CreateBlogCategory creates a new blog category
func (r *ContentRepository) CreateBlogCategory(ctx context.Context, c *BlogCategory) (*BlogCategory, error) {
	query := `INSERT INTO blog_categories (title, slug, description, color, icon, "order")
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6)
		RETURNING id, created_at, updated_at`
	out := &BlogCategory{Title: c.Title, Slug: c.Slug, Description: c.Description, Color: c.Color, Icon: c.Icon, Order: c.Order}
	err := r.db.QueryRowContext(ctx, query, c.Title, c.Slug, c.Description, c.Color, c.Icon, c.Order).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog category: %w", err)
	}
	return out, nil
}

// GetBlogCategoryByID returns a blog category by ID
func (r *ContentRepository) GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*BlogCategory, error) {
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at FROM blog_categories WHERE id = $1`
	c := &BlogCategory{}
	var desc, color, icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog category not found")
		}
		return nil, fmt.Errorf("failed to get blog category: %w", err)
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if color.Valid {
		c.Color = color.String
	}
	if icon.Valid {
		c.Icon = icon.String
	}
	return c, nil
}

// GetBlogCategoryBySlug returns a blog category by slug
func (r *ContentRepository) GetBlogCategoryBySlug(ctx context.Context, slug string) (*BlogCategory, error) {
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at FROM blog_categories WHERE slug = $1`
	c := &BlogCategory{}
	var desc, color, icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog category not found")
		}
		return nil, fmt.Errorf("failed to get blog category: %w", err)
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if color.Valid {
		c.Color = color.String
	}
	if icon.Valid {
		c.Icon = icon.String
	}
	return c, nil
}

// UpdateBlogCategory updates a blog category
func (r *ContentRepository) UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogCategory, error) {
	if len(updates) == 0 {
		return r.GetBlogCategoryByID(ctx, id)
	}
	set := []string{}
	args := []interface{}{}
	pos := 1
	for k, v := range updates {
		switch k {
		case "title", "slug", "description", "color", "icon":
			set = append(set, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case "order":
			set = append(set, fmt.Sprintf(`"order" = $%d`, pos))
			args = append(args, v)
			pos++
		}
	}
	if len(set) == 0 {
		return r.GetBlogCategoryByID(ctx, id)
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE blog_categories SET %s, updated_at = NOW() WHERE id = $%d RETURNING id, title, slug, description, color, icon, \"order\", created_at, updated_at", strings.Join(set, ", "), pos)
	c := &BlogCategory{}
	var desc, color, icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog category not found")
		}
		return nil, fmt.Errorf("failed to update blog category: %w", err)
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if color.Valid {
		c.Color = color.String
	}
	if icon.Valid {
		c.Icon = icon.String
	}
	return c, nil
}

// DeleteBlogCategory deletes a blog category
func (r *ContentRepository) DeleteBlogCategory(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM blog_categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete blog category: %w", err)
	}
	return nil
}

// ListBlogAuthors returns all blog authors
func (r *ContentRepository) ListBlogAuthors(ctx context.Context) ([]*BlogAuthor, error) {
	query := `SELECT id, name, slug, bio, photo, email, website, social_links, role, active, created_at, updated_at FROM blog_authors ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list blog authors: %w", err)
	}
	defer rows.Close()
	var list []*BlogAuthor
	for rows.Next() {
		a, err := r.scanBlogAuthor(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *ContentRepository) scanBlogAuthor(row interface {
	Scan(dest ...interface{}) error
}) (*BlogAuthor, error) {
	a := &BlogAuthor{}
	var bio, email, website, role sql.NullString
	var photoBytes, socialBytes []byte
	err := row.Scan(&a.ID, &a.Name, &a.Slug, &bio, &photoBytes, &email, &website, &socialBytes, &role, &a.Active, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if bio.Valid {
		a.Bio = bio.String
	}
	if email.Valid {
		a.Email = email.String
	}
	if website.Valid {
		a.Website = website.String
	}
	if role.Valid {
		a.Role = role.String
	}
	if len(photoBytes) > 0 {
		_ = json.Unmarshal(photoBytes, &a.Photo)
	}
	if len(socialBytes) > 0 {
		_ = json.Unmarshal(socialBytes, &a.SocialLinks)
	}
	return a, nil
}

// CreateBlogAuthor creates a new blog author
func (r *ContentRepository) CreateBlogAuthor(ctx context.Context, a *BlogAuthor) (*BlogAuthor, error) {
	photoJSON := nullJSON(a.Photo)
	socialJSON := nullJSON(a.SocialLinks)
	query := `INSERT INTO blog_authors (name, slug, bio, photo, email, website, social_links, role, active)
		VALUES ($1, $2, NULLIF($3,''), $4, NULLIF($5,''), NULLIF($6,''), $7, NULLIF($8,''), $9)
		RETURNING id, created_at, updated_at`
	out := &BlogAuthor{Name: a.Name, Slug: a.Slug, Bio: a.Bio, Photo: a.Photo, Email: a.Email, Website: a.Website, SocialLinks: a.SocialLinks, Role: a.Role, Active: a.Active}
	err := r.db.QueryRowContext(ctx, query, a.Name, a.Slug, a.Bio, photoJSON, a.Email, a.Website, socialJSON, a.Role, a.Active).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog author: %w", err)
	}
	return out, nil
}

func nullJSON(m map[string]interface{}) interface{} {
	if m == nil || len(m) == 0 {
		return nil
	}
	b, _ := json.Marshal(m)
	return string(b) // PostgreSQL jsonb accepts JSON string
}

// GetBlogAuthorByID returns a blog author by ID
func (r *ContentRepository) GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*BlogAuthor, error) {
	query := `SELECT id, name, slug, bio, photo, email, website, social_links, role, active, created_at, updated_at FROM blog_authors WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	a, err := r.scanBlogAuthor(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog author not found")
		}
		return nil, err
	}
	return a, nil
}

// GetBlogAuthorBySlug returns a blog author by slug
func (r *ContentRepository) GetBlogAuthorBySlug(ctx context.Context, slug string) (*BlogAuthor, error) {
	query := `SELECT id, name, slug, bio, photo, email, website, social_links, role, active, created_at, updated_at FROM blog_authors WHERE slug = $1`
	row := r.db.QueryRowContext(ctx, query, slug)
	a, err := r.scanBlogAuthor(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog author not found")
		}
		return nil, err
	}
	return a, nil
}

// UpdateBlogAuthor updates a blog author
func (r *ContentRepository) UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogAuthor, error) {
	if len(updates) == 0 {
		return r.GetBlogAuthorByID(ctx, id)
	}
	set := []string{}
	args := []interface{}{}
	pos := 1
	for k, v := range updates {
		switch k {
		case "name", "slug", "bio", "email", "website", "role":
			set = append(set, fmt.Sprintf("%s = $%d", k, pos))
			args = append(args, v)
			pos++
		case "active":
			set = append(set, fmt.Sprintf("active = $%d", pos))
			args = append(args, v)
			pos++
		case "photo":
			set = append(set, fmt.Sprintf("photo = $%d", pos))
			if m, ok := v.(map[string]interface{}); ok {
				args = append(args, nullJSON(m))
			} else {
				args = append(args, v)
			}
			pos++
		case "social_links", "socialLinks":
			set = append(set, fmt.Sprintf("social_links = $%d", pos))
			if m, ok := v.(map[string]interface{}); ok {
				args = append(args, nullJSON(m))
			} else {
				args = append(args, v)
			}
			pos++
		}
	}
	if len(set) == 0 {
		return r.GetBlogAuthorByID(ctx, id)
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE blog_authors SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(set, ", "), pos)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog author: %w", err)
	}
	return r.GetBlogAuthorByID(ctx, id)
}

// DeleteBlogAuthor deletes a blog author
func (r *ContentRepository) DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM blog_authors WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete blog author: %w", err)
	}
	return nil
}

// GetBlogSettings retrieves the single blog settings row
func (r *ContentRepository) GetBlogSettings(ctx context.Context) (*BlogSettings, error) {
	query := `SELECT id, blog_title, posts_per_page, meta_description, created_at, updated_at FROM blog_settings LIMIT 1`
	s := &BlogSettings{}
	var title, desc string
	var postsPerPage int
	err := r.db.QueryRowContext(ctx, query).Scan(&s.ID, &title, &postsPerPage, &desc, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get blog settings: %w", err)
	}
	s.BlogTitle = title
	s.PostsPerPage = postsPerPage
	s.MetaDescription = desc
	return s, nil
}

// UpdateBlogSettings updates the blog settings row
func (r *ContentRepository) UpdateBlogSettings(ctx context.Context, updates map[string]interface{}) (*BlogSettings, error) {
	if len(updates) == 0 {
		return r.GetBlogSettings(ctx)
	}
	set := []string{}
	args := []interface{}{}
	pos := 1
	for k, v := range updates {
		switch k {
		case "blog_title", "blogTitle":
			set = append(set, fmt.Sprintf("blog_title = $%d", pos))
			args = append(args, v)
			pos++
		case "posts_per_page", "postsPerPage":
			set = append(set, fmt.Sprintf("posts_per_page = $%d", pos))
			args = append(args, v)
			pos++
		case "meta_description", "metaDescription":
			set = append(set, fmt.Sprintf("meta_description = $%d", pos))
			args = append(args, v)
			pos++
		}
	}
	if len(set) == 0 {
		return r.GetBlogSettings(ctx)
	}
	args = append(args, time.Now())
	query := fmt.Sprintf("UPDATE blog_settings SET %s, updated_at = $%d", strings.Join(set, ", "), pos)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog settings: %w", err)
	}
	return r.GetBlogSettings(ctx)
}
