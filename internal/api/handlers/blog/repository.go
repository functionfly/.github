package blog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type BlogRepository struct {
	db *sql.DB
}

func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

type BlogPostFilter struct {
	Status   string
	Category string
	Author   string
	Tag      string
	Search   string
	Limit    int
	Offset   int
	Admin    bool
}

func (r *BlogRepository) ListPosts(filter BlogPostFilter) ([]BlogPost, int, error) {
	ctx := context.Background()

	// Build WHERE conditions
	conditions := []string{}
	args := []interface{}{}
	argNum := 1

	// Default to published/scheduled for public API
	if !filter.Admin {
		if filter.Status != "" {
			conditions = append(conditions, fmt.Sprintf("bp.status = $%d", argNum))
			args = append(args, filter.Status)
			argNum++
		} else {
			conditions = append(conditions, fmt.Sprintf("(bp.status = $%d OR bp.status = 'scheduled')", argNum))
			args = append(args, "published")
			argNum++
		}
	} else if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("bp.status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}

	if filter.Category != "" {
		// Join with categories to filter by slug
		conditions = append(conditions, fmt.Sprintf("bc.slug = $%d", argNum))
		args = append(args, filter.Category)
		argNum++
	}

	if filter.Author != "" {
		// Join with authors to filter by slug
		conditions = append(conditions, fmt.Sprintf("ba.slug = $%d", argNum))
		args = append(args, filter.Author)
		argNum++
	}

	if filter.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("bp.tags @> $%d", argNum))
		args = append(args, pq.Array([]string{filter.Tag}))
		argNum++
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		conditions = append(conditions,
			fmt.Sprintf("(bp.title ILIKE $%d OR bp.description ILIKE $%d OR CAST(bp.body AS TEXT) ILIKE $%d OR bp.tags::text ILIKE $%d)", argNum, argNum+1, argNum+2, argNum+3),
		)
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		argNum += 4
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM blog_posts bp
		LEFT JOIN blog_categories bc ON bp.category_id = bc.id
		LEFT JOIN blog_authors ba ON bp.author_id = ba.id
		%s`, whereClause)

	var total int
	countArgs := args[:len(args)]
	if filter.Category != "" {
		countArgs = args[:len(args)-1]
	}
	if filter.Author != "" {
		countArgs = args[:len(args)-1]
	}

	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Get posts
	query := fmt.Sprintf(`
		SELECT bp.id, bp.title, bp.slug, bp.description, bp.body, bp.author_id, bp.category_id,
			   bp.tags, bp.hero_image, bp.status, bp.published_at, bp.scheduled_at,
			   bp.updated_at, bp.created_at, bp.seo_title, bp.seo_description,
			   bp.keywords, bp.canonical_url, bp.og_image, bp.campaign, bp.owner_id
		FROM blog_posts bp
		LEFT JOIN blog_categories bc ON bp.category_id = bc.id
		LEFT JOIN blog_authors ba ON bp.author_id = ba.id
		%s
		ORDER BY bp.published_at DESC NULLS LAST
		LIMIT $%d OFFSET $%d`, whereClause, argNum, argNum+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []BlogPost
	for rows.Next() {
		post, err := r.scanBlogPost(rows)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, *post)
	}

	return posts, total, rows.Err()
}

func (r *BlogRepository) GetPostByID(id uuid.UUID) (*BlogPost, error) {
	ctx := context.Background()

	query := `
		SELECT bp.id, bp.title, bp.slug, bp.description, bp.body, bp.author_id, bp.category_id,
			   bp.tags, bp.hero_image, bp.status, bp.published_at, bp.scheduled_at,
			   bp.updated_at, bp.created_at, bp.seo_title, bp.seo_description,
			   bp.keywords, bp.canonical_url, bp.og_image, bp.campaign, bp.owner_id
		FROM blog_posts bp
		WHERE bp.id = $1
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, id)
	post, err := r.scanBlogPost(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found")
		}
		return nil, err
	}

	if post.AuthorID != nil {
		author, _ := r.GetAuthorByID(*post.AuthorID)
		if author != nil {
			post.Author = &AuthorSummary{Name: author.Name, Slug: author.Slug}
		}
	}
	if post.CategoryID != nil {
		category, _ := r.GetCategoryByID(*post.CategoryID)
		if category != nil {
			post.Category = &CategorySummary{Title: category.Title, Slug: category.Slug}
		}
	}

	return post, nil
}

func (r *BlogRepository) GetPostBySlug(slug string) (*BlogPost, error) {
	ctx := context.Background()

	query := `
		SELECT bp.id, bp.title, bp.slug, bp.description, bp.body, bp.author_id, bp.category_id,
			   bp.tags, bp.hero_image, bp.status, bp.published_at, bp.scheduled_at,
			   bp.updated_at, bp.created_at, bp.seo_title, bp.seo_description,
			   bp.keywords, bp.canonical_url, bp.og_image, bp.campaign, bp.owner_id
		FROM blog_posts bp
		WHERE bp.slug = $1 AND (bp.status = 'published' OR bp.status = 'scheduled')
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, slug)
	post, err := r.scanBlogPost(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found")
		}
		return nil, err
	}

	if post.AuthorID != nil {
		author, _ := r.GetAuthorByID(*post.AuthorID)
		if author != nil {
			post.Author = &AuthorSummary{Name: author.Name, Slug: author.Slug}
		}
	}
	if post.CategoryID != nil {
		category, _ := r.GetCategoryByID(*post.CategoryID)
		if category != nil {
			post.Category = &CategorySummary{Title: category.Title, Slug: category.Slug}
		}
	}

	return post, nil
}

func (r *BlogRepository) CreatePost(post *BlogPost) (*BlogPost, error) {
	ctx := context.Background()

	bodyJSON, _ := json.Marshal(post.Body)
	tagsJSON, _ := json.Marshal(post.Tags)
	keywordsJSON, _ := json.Marshal(post.Keywords)

	var heroImageJSON []byte
	if post.HeroImage != nil {
		heroImageJSON, _ = json.Marshal(post.HeroImage)
	}

	query := `
		INSERT INTO blog_posts (title, slug, description, body, author_id, category_id, tags,
								hero_image, status, published_at, scheduled_at, seo_title,
								seo_description, keywords, canonical_url, campaign, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query,
		post.Title, post.Slug, post.Description, bodyJSON,
		post.AuthorID, post.CategoryID, tagsJSON,
		heroImageJSON, post.Status, post.PublishedAt, post.ScheduledAt,
		post.SEOTitle, post.SEODescription, keywordsJSON,
		post.CanonicalURL, post.Campaign, post.OwnerID,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create blog post: %w", err)
	}

	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = &updatedAt

	return post, nil
}

func (r *BlogRepository) UpdatePost(id uuid.UUID, updates map[string]interface{}) (*BlogPost, error) {
	ctx := context.Background()

	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	setParts := []string{}
	args := []interface{}{}
	argNum := 1

	for key, value := range updates {
		switch key {
		case "title", "slug", "description", "status":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argNum))
			args = append(args, value)
			argNum++
		case "body":
			setParts = append(setParts, fmt.Sprintf("body = $%d", argNum))
			bodyJSON, _ := json.Marshal(value)
			args = append(args, bodyJSON)
			argNum++
		case "tags":
			setParts = append(setParts, fmt.Sprintf("tags = $%d", argNum))
			tagsJSON, _ := json.Marshal(value)
			args = append(args, tagsJSON)
			argNum++
		case "keywords":
			setParts = append(setParts, fmt.Sprintf("keywords = $%d", argNum))
			keywordsJSON, _ := json.Marshal(value)
			args = append(args, keywordsJSON)
			argNum++
		case "heroImage", "hero_image":
			setParts = append(setParts, fmt.Sprintf("hero_image = $%d", argNum))
			heroJSON, _ := json.Marshal(value)
			args = append(args, heroJSON)
			argNum++
		case "seoTitle", "seo_title":
			setParts = append(setParts, fmt.Sprintf("seo_title = $%d", argNum))
			args = append(args, value)
			argNum++
		case "seoDescription", "seo_description":
			setParts = append(setParts, fmt.Sprintf("seo_description = $%d", argNum))
			args = append(args, value)
			argNum++
		case "canonicalUrl", "canonical_url":
			setParts = append(setParts, fmt.Sprintf("canonical_url = $%d", argNum))
			args = append(args, value)
			argNum++
		case "campaign":
			setParts = append(setParts, fmt.Sprintf("campaign = $%d", argNum))
			args = append(args, value)
			argNum++
		case "isPublished", "is_published":
			setParts = append(setParts, fmt.Sprintf("status = $%d", argNum))
			if v, ok := value.(bool); ok {
				if v {
					args = append(args, "published")
				} else {
					args = append(args, "draft")
				}
			} else {
				args = append(args, value)
			}
			argNum++
		case "featuredImage", "featured_image":
			setParts = append(setParts, fmt.Sprintf("hero_image = $%d", argNum))
			heroJSON, _ := json.Marshal(value)
			args = append(args, heroJSON)
			argNum++
		case "excerpt":
			setParts = append(setParts, fmt.Sprintf("description = $%d", argNum))
			args = append(args, value)
			argNum++
		case "author":
			if authorStr, ok := value.(string); ok && authorStr != "" {
				author, err := r.FindOrCreateAuthorByName(authorStr)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve author: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("author_id = $%d", argNum))
				args = append(args, author.ID)
				argNum++
			}
		case "authorId", "author_id":
			setParts = append(setParts, fmt.Sprintf("author_id = $%d", argNum))
			args = append(args, value)
			argNum++
		case "categoryId", "category_id":
			setParts = append(setParts, fmt.Sprintf("category_id = $%d", argNum))
			args = append(args, value)
			argNum++
		case "publishedAt", "published_at":
			setParts = append(setParts, fmt.Sprintf("published_at = $%d", argNum))
			args = append(args, value)
			argNum++
		case "scheduledAt", "scheduled_at":
			setParts = append(setParts, fmt.Sprintf("scheduled_at = $%d", argNum))
			args = append(args, value)
			argNum++
		}
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no valid updates provided")
	}

	// Always update updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	query := fmt.Sprintf("UPDATE blog_posts SET %s WHERE id = $%d", strings.Join(setParts, ", "), argNum)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog post: %w", err)
	}

	// Return the updated post - fetch it
	return r.GetPostByID(id)
}

func (r *BlogRepository) DeletePost(id uuid.UUID) error {
	ctx := context.Background()
	query := "DELETE FROM blog_posts WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *BlogRepository) ListCategories() ([]Category, error) {
	ctx := context.Background()
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at
			  FROM blog_categories ORDER BY "order" ASC, title ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		var desc, color, icon sql.NullString
		err := rows.Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
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
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *BlogRepository) CreateCategory(c *Category) (*Category, error) {
	ctx := context.Background()
	query := `INSERT INTO blog_categories (title, slug, description, color, icon, "order")
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, c.Title, c.Slug, c.Description, c.Color, c.Icon, c.Order).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *BlogRepository) GetCategoryByID(id uuid.UUID) (*Category, error) {
	ctx := context.Background()
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at
			  FROM blog_categories WHERE id = $1`

	var c Category
	var desc, color, icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
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
	return &c, nil
}

func (r *BlogRepository) GetCategoryBySlug(slug string) (*Category, error) {
	ctx := context.Background()
	query := `SELECT id, title, slug, description, color, icon, "order", created_at, updated_at
			  FROM blog_categories WHERE slug = $1`

	var c Category
	var desc, color, icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).
		Scan(&c.ID, &c.Title, &c.Slug, &desc, &color, &icon, &c.Order, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
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
	return &c, nil
}

func (r *BlogRepository) UpdateCategory(id uuid.UUID, updates map[string]interface{}) (*Category, error) {
	ctx := context.Background()

	setParts := []string{}
	args := []interface{}{}
	argNum := 1

	for key, value := range updates {
		switch key {
		case "title", "slug", "description", "color", "icon":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argNum))
			args = append(args, value)
			argNum++
		case "order":
			setParts = append(setParts, fmt.Sprintf("\"order\" = $%d", argNum))
			args = append(args, value)
			argNum++
		}
	}

	if len(setParts) == 0 {
		return r.GetCategoryByID(id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	query := fmt.Sprintf("UPDATE blog_categories SET %s WHERE id = $%d", strings.Join(setParts, ", "), argNum)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetCategoryByID(id)
}

func (r *BlogRepository) DeleteCategory(id uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, "DELETE FROM blog_categories WHERE id = $1", id)
	return err
}

func (r *BlogRepository) ListAuthors() ([]Author, error) {
	ctx := context.Background()
	query := `SELECT id, name, slug, bio, email, website, photo, social_links, role, active, created_at, updated_at
			  FROM blog_authors ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var a Author
		var bio, email, website, role sql.NullString
		var photoBytes, socialBytes []byte
		err := rows.Scan(&a.ID, &a.Name, &a.Slug, &bio, &email, &website, &photoBytes, &socialBytes, &role, &a.Active, &a.CreatedAt, &a.UpdatedAt)
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
			json.Unmarshal(photoBytes, &a.Photo)
		}
		if len(socialBytes) > 0 {
			json.Unmarshal(socialBytes, &a.SocialLinks)
		}
		authors = append(authors, a)
	}
	return authors, rows.Err()
}

func (r *BlogRepository) CreateAuthor(a *Author) (*Author, error) {
	ctx := context.Background()

	photoJSON, _ := json.Marshal(a.Photo)
	socialJSON, _ := json.Marshal(a.SocialLinks)

	query := `INSERT INTO blog_authors (name, slug, bio, email, website, photo, social_links, role, active)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, a.Name, a.Slug, a.Bio, a.Email, a.Website, photoJSON, socialJSON, a.Role, a.Active).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *BlogRepository) GetAuthorByID(id uuid.UUID) (*Author, error) {
	ctx := context.Background()
	query := `SELECT id, name, slug, bio, email, website, photo, social_links, role, active, created_at, updated_at
			  FROM blog_authors WHERE id = $1`

	var a Author
	var bio, email, website, role sql.NullString
	var photoBytes, socialBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&a.ID, &a.Name, &a.Slug, &bio, &email, &website, &photoBytes, &socialBytes, &role, &a.Active, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("author not found")
		}
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
		json.Unmarshal(photoBytes, &a.Photo)
	}
	if len(socialBytes) > 0 {
		json.Unmarshal(socialBytes, &a.SocialLinks)
	}
	return &a, nil
}

func (r *BlogRepository) FindOrCreateAuthorByName(name string) (*Author, error) {
	ctx := context.Background()

	// Try to find existing author by name
	query := `SELECT id, name, slug, bio, email, website, photo, social_links, role, active, created_at, updated_at
			  FROM blog_authors WHERE name = $1`
	var a Author
	var bio, email, website, role sql.NullString
	var photoBytes, socialBytes []byte
	err := r.db.QueryRowContext(ctx, query, name).
		Scan(&a.ID, &a.Name, &a.Slug, &bio, &email, &website, &photoBytes, &socialBytes, &role, &a.Active, &a.CreatedAt, &a.UpdatedAt)
	if err == nil {
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
			json.Unmarshal(photoBytes, &a.Photo)
		}
		if len(socialBytes) > 0 {
			json.Unmarshal(socialBytes, &a.SocialLinks)
		}
		return &a, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new author with slugified name
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	now := time.Now()
	newAuthor := &Author{
		Name:     name,
		Slug:     slug,
		Active:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	insertQuery := `INSERT INTO blog_authors (name, slug, active, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING id`
	err = r.db.QueryRowContext(ctx, insertQuery, newAuthor.Name, newAuthor.Slug, newAuthor.Active, newAuthor.CreatedAt, newAuthor.UpdatedAt).
		Scan(&newAuthor.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}
	return newAuthor, nil
}

func (r *BlogRepository) GetAuthorBySlug(slug string) (*Author, error) {
	ctx := context.Background()
	query := `SELECT id, name, slug, bio, email, website, photo, social_links, role, active, created_at, updated_at
			  FROM blog_authors WHERE slug = $1`

	var a Author
	var bio, email, website, role sql.NullString
	var photoBytes, socialBytes []byte
	err := r.db.QueryRowContext(ctx, query, slug).
		Scan(&a.ID, &a.Name, &a.Slug, &bio, &email, &website, &photoBytes, &socialBytes, &role, &a.Active, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("author not found")
		}
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
		json.Unmarshal(photoBytes, &a.Photo)
	}
	if len(socialBytes) > 0 {
		json.Unmarshal(socialBytes, &a.SocialLinks)
	}
	return &a, nil
}

func (r *BlogRepository) UpdateAuthor(id uuid.UUID, updates map[string]interface{}) (*Author, error) {
	ctx := context.Background()

	setParts := []string{}
	args := []interface{}{}
	argNum := 1

	for key, value := range updates {
		switch key {
		case "name", "slug", "bio", "email", "website", "role":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argNum))
			args = append(args, value)
			argNum++
		case "active":
			setParts = append(setParts, fmt.Sprintf("active = $%d", argNum))
			args = append(args, value)
			argNum++
		case "photo":
			setParts = append(setParts, fmt.Sprintf("photo = $%d", argNum))
			photoJSON, _ := json.Marshal(value)
			args = append(args, photoJSON)
			argNum++
		case "socialLinks", "social_links":
			setParts = append(setParts, fmt.Sprintf("social_links = $%d", argNum))
			socialJSON, _ := json.Marshal(value)
			args = append(args, socialJSON)
			argNum++
		}
	}

	if len(setParts) == 0 {
		return r.GetAuthorByID(id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	query := fmt.Sprintf("UPDATE blog_authors SET %s WHERE id = $%d", strings.Join(setParts, ", "), argNum)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetAuthorByID(id)
}

func (r *BlogRepository) DeleteAuthor(id uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, "DELETE FROM blog_authors WHERE id = $1", id)
	return err
}

type scanableBlogPost interface {
	Scan(dest ...interface{}) error
}

func (r *BlogRepository) scanBlogPost(row scanableBlogPost) (*BlogPost, error) {
	var post BlogPost
	var authorID, categoryID, ownerID sql.NullString
	var tagsBytes, heroImageBytes, keywordsBytes, ogImageBytes []byte
	var bodyBytes []byte
	var seoTitle, seoDescription, canonicalURL, campaign sql.NullString
	var publishedAt, scheduledAt, updatedAt sql.NullString

	err := row.Scan(
		&post.ID, &post.Title, &post.Slug, &post.Description, &bodyBytes,
		&authorID, &categoryID, &tagsBytes, &heroImageBytes,
		&post.Status, &publishedAt, &scheduledAt, &updatedAt, &post.CreatedAt,
		&seoTitle, &seoDescription, &keywordsBytes, &canonicalURL, &ogImageBytes, &campaign, &ownerID,
	)
	if err != nil {
		return nil, err
	}

	if bodyBytes != nil {
		if err := json.Unmarshal(bodyBytes, &post.Body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal body: %w", err)
		}
	}

	if authorID.Valid {
		id, _ := uuid.Parse(authorID.String)
		post.AuthorID = &id
	}
	if categoryID.Valid {
		id, _ := uuid.Parse(categoryID.String)
		post.CategoryID = &id
	}
	if ownerID.Valid {
		id, _ := uuid.Parse(ownerID.String)
		post.OwnerID = &id
	}
	if tagsBytes != nil {
		json.Unmarshal(tagsBytes, &post.Tags)
	}
	if heroImageBytes != nil {
		json.Unmarshal(heroImageBytes, &post.HeroImage)
	}
	if keywordsBytes != nil {
		json.Unmarshal(keywordsBytes, &post.Keywords)
	}
	if ogImageBytes != nil {
		json.Unmarshal(ogImageBytes, &post.OGImage)
	}
	if seoTitle.Valid {
		post.SEOTitle = &seoTitle.String
	}
	if seoDescription.Valid {
		post.SEODescription = &seoDescription.String
	}
	if canonicalURL.Valid {
		post.CanonicalURL = &canonicalURL.String
	}
	if campaign.Valid {
		post.Campaign = &campaign.String
	}
	if publishedAt.Valid {
		t, _ := time.Parse(time.RFC3339, publishedAt.String)
		post.PublishedAt = &t
	}
	if scheduledAt.Valid {
		t, _ := time.Parse(time.RFC3339, scheduledAt.String)
		post.ScheduledAt = &t
	}
	if updatedAt.Valid {
		t, _ := time.Parse(time.RFC3339, updatedAt.String)
		post.UpdatedAt = &t
	}

	return &post, nil
}

func (r *BlogRepository) GetSettings() (*BlogSettings, error) {
	ctx := context.Background()
	query := `SELECT id, blog_title, posts_per_page, meta_description, created_at, updated_at
			  FROM blog_settings LIMIT 1`

	var s BlogSettings
	var metaDesc sql.NullString
	err := r.db.QueryRowContext(ctx, query).
		Scan(&s.ID, &s.BlogTitle, &s.PostsPerPage, &metaDesc, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog settings not found")
		}
		return nil, err
	}
	if metaDesc.Valid {
		s.MetaDescription = metaDesc.String
	}
	return &s, nil
}

func (r *BlogRepository) UpdateSettings(updates map[string]interface{}) (*BlogSettings, error) {
	ctx := context.Background()

	setParts := []string{}
	args := []interface{}{}
	argNum := 1

	for key, value := range updates {
		switch key {
		case "blogTitle":
			setParts = append(setParts, fmt.Sprintf("blog_title = $%d", argNum))
			args = append(args, value)
			argNum++
		case "postsPerPage":
			setParts = append(setParts, fmt.Sprintf("posts_per_page = $%d", argNum))
			args = append(args, value)
			argNum++
		case "metaDescription":
			setParts = append(setParts, fmt.Sprintf("meta_description = $%d", argNum))
			args = append(args, value)
			argNum++
		}
	}

	if len(setParts) == 0 {
		return r.GetSettings()
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	query := fmt.Sprintf("UPDATE blog_settings SET %s", strings.Join(setParts, ", "))
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetSettings()
}

func (r *BlogRepository) GetRelatedPosts(postID uuid.UUID) ([]uuid.UUID, error) {
	ctx := context.Background()
	query := `SELECT related_post_id FROM blog_related_posts WHERE post_id = $1 ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *BlogRepository) SetRelatedPosts(postID uuid.UUID, relatedIDs []uuid.UUID) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM blog_related_posts WHERE post_id = $1", postID)
	if err != nil {
		return err
	}

	for _, relatedID := range relatedIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO blog_related_posts (post_id, related_post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			postID, relatedID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BlogRepository) GetCTABlocks(postID uuid.UUID) ([]BlogCTABlock, error) {
	ctx := context.Background()
	query := `SELECT id, post_id, title, description, button_text, button_url, style, "order"
			  FROM blog_cta_blocks WHERE post_id = $1 ORDER BY "order" ASC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []BlogCTABlock
	for rows.Next() {
		var b BlogCTABlock
		var desc sql.NullString
		err := rows.Scan(&b.ID, &b.PostID, &b.Title, &desc, &b.ButtonText, &b.ButtonURL, &b.Style, &b.Order)
		if err != nil {
			return nil, err
		}
		if desc.Valid {
			b.Description = desc.String
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

func (r *BlogRepository) SetCTABlocks(postID uuid.UUID, blocks []BlogCTABlock) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM blog_cta_blocks WHERE post_id = $1", postID)
	if err != nil {
		return err
	}

	for i, b := range blocks {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO blog_cta_blocks (post_id, title, description, button_text, button_url, style, "order")
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			postID, b.Title, b.Description, b.ButtonText, b.ButtonURL, b.Style, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
