-- Migration: Remove seeded blog authors and posts
-- Reverses: 20260429130000_create_blog_tables (which includes seeding)

-- Delete blog posts by their seeded slugs
DELETE FROM blog_posts WHERE slug IN (
    'welcome-to-functionfly',
    'durable-execution-for-ai-agents',
    'state-fabric',
    'trust-layer-for-ai-agents',
    'zero-knowledge-secrets-vault',
    'building-first-ai-agent'
);

-- Delete blog categories by their seeded slugs
DELETE FROM blog_categories WHERE slug IN (
    'announcements', 'trust', 'security', 'architecture', 'ai-agents', 'tutorials', 'stories'
);

-- Delete blog author
DELETE FROM blog_authors WHERE slug = 'functionfly-team';