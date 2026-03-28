package storage

import (
	"context"

	"github.com/google/uuid"
)

// PostgresDB methods: changelog and blog content.

// Content operations
func (db *PostgresDB) CreateChangelogEntry(ctx context.Context, entry *ChangelogEntry) (*ChangelogEntry, error) {
	return db.contentRepository.CreateChangelogEntry(ctx, entry)
}

func (db *PostgresDB) GetChangelogEntryByID(id uuid.UUID) (*ChangelogEntry, error) {
	return db.contentRepository.GetChangelogEntryByID(id)
}

func (db *PostgresDB) GetChangelogEntryByVersion(version string) (*ChangelogEntry, error) {
	return db.contentRepository.GetChangelogEntryByVersion(version)
}

func (db *PostgresDB) ListChangelogEntries(limit, offset int, publishedOnly bool) ([]*ChangelogEntry, error) {
	return db.contentRepository.ListChangelogEntries(limit, offset, publishedOnly)
}

func (db *PostgresDB) UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogEntry, error) {
	return db.contentRepository.UpdateChangelogEntry(ctx, id, updates)
}

func (db *PostgresDB) DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteChangelogEntry(ctx, id)
}

func (db *PostgresDB) CreateChangelogChange(ctx context.Context, change *ChangelogChange) (*ChangelogChange, error) {
	return db.contentRepository.CreateChangelogChange(ctx, change)
}

func (db *PostgresDB) UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogChange, error) {
	return db.contentRepository.UpdateChangelogChange(ctx, id, updates)
}

func (db *PostgresDB) DeleteChangelogChange(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteChangelogChange(ctx, id)
}

func (db *PostgresDB) CreateBlogPost(ctx context.Context, post *BlogPost) (*BlogPost, error) {
	return db.contentRepository.CreateBlogPost(ctx, post)
}

func (db *PostgresDB) GetBlogPostByID(id uuid.UUID) (*BlogPost, error) {
	return db.contentRepository.GetBlogPostByID(id)
}

func (db *PostgresDB) GetBlogPostBySlug(slug string) (*BlogPost, error) {
	return db.contentRepository.GetBlogPostBySlug(slug)
}

func (db *PostgresDB) ListBlogPosts(limit, offset int, publishedOnly bool, tagFilter []string) ([]*BlogPost, error) {
	return db.contentRepository.ListBlogPosts(limit, offset, publishedOnly, tagFilter)
}

func (db *PostgresDB) UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogPost, error) {
	return db.contentRepository.UpdateBlogPost(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogPost(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogPost(ctx, id)
}

func (db *PostgresDB) ListBlogCategories(ctx context.Context) ([]*BlogCategory, error) {
	return db.contentRepository.ListBlogCategories(ctx)
}

func (db *PostgresDB) CreateBlogCategory(ctx context.Context, c *BlogCategory) (*BlogCategory, error) {
	return db.contentRepository.CreateBlogCategory(ctx, c)
}

func (db *PostgresDB) GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*BlogCategory, error) {
	return db.contentRepository.GetBlogCategoryByID(ctx, id)
}

func (db *PostgresDB) GetBlogCategoryBySlug(ctx context.Context, slug string) (*BlogCategory, error) {
	return db.contentRepository.GetBlogCategoryBySlug(ctx, slug)
}

func (db *PostgresDB) UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogCategory, error) {
	return db.contentRepository.UpdateBlogCategory(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogCategory(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogCategory(ctx, id)
}

func (db *PostgresDB) ListBlogAuthors(ctx context.Context) ([]*BlogAuthor, error) {
	return db.contentRepository.ListBlogAuthors(ctx)
}

func (db *PostgresDB) CreateBlogAuthor(ctx context.Context, a *BlogAuthor) (*BlogAuthor, error) {
	return db.contentRepository.CreateBlogAuthor(ctx, a)
}

func (db *PostgresDB) GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*BlogAuthor, error) {
	return db.contentRepository.GetBlogAuthorByID(ctx, id)
}

func (db *PostgresDB) GetBlogAuthorBySlug(ctx context.Context, slug string) (*BlogAuthor, error) {
	return db.contentRepository.GetBlogAuthorBySlug(ctx, slug)
}

func (db *PostgresDB) UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogAuthor, error) {
	return db.contentRepository.UpdateBlogAuthor(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogAuthor(ctx, id)
}
