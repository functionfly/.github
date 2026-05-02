-- Migration: Create blog tables for NestJS blog-api migration
-- This creates the schema used by the Go blog handler

-- Content status enum
CREATE TYPE content_status AS ENUM ('draft', 'in_review', 'approved', 'scheduled', 'published');

-- Authors table
CREATE TABLE IF NOT EXISTS blog_authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(96) NOT NULL UNIQUE,
    bio TEXT,
    email VARCHAR(255),
    website VARCHAR(255),
    photo JSONB,
    social_links JSONB,
    role VARCHAR(100),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_blog_authors_slug ON blog_authors(slug);
CREATE INDEX IF NOT EXISTS idx_blog_authors_active ON blog_authors(active);

-- Categories table
CREATE TABLE IF NOT EXISTS blog_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(96) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7),
    icon VARCHAR(50),
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_blog_categories_slug ON blog_categories(slug);
CREATE INDEX IF NOT EXISTS idx_blog_categories_order ON blog_categories("order");

-- Blog posts table
CREATE TABLE IF NOT EXISTS blog_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(96) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    body JSONB NOT NULL DEFAULT '{}',
    author_id UUID REFERENCES blog_authors(id) ON DELETE SET NULL,
    category_id UUID REFERENCES blog_categories(id) ON DELETE SET NULL,
    tags JSONB DEFAULT '[]',
    hero_image JSONB,
    status content_status NOT NULL DEFAULT 'draft',
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    seo_title VARCHAR(70),
    seo_description TEXT,
    keywords JSONB DEFAULT '[]',
    canonical_url VARCHAR(500),
    og_image JSONB,
    campaign VARCHAR(100),
    owner_id UUID REFERENCES blog_authors(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_blog_posts_slug ON blog_posts(slug);
CREATE INDEX IF NOT EXISTS idx_blog_posts_status ON blog_posts(status);
CREATE INDEX IF NOT EXISTS idx_blog_posts_published_at ON blog_posts(published_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_blog_posts_author_id ON blog_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_blog_posts_category_id ON blog_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_blog_posts_tags ON blog_posts USING GIN(tags);

-- Related posts junction table
CREATE TABLE IF NOT EXISTS blog_related_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    related_post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    UNIQUE(post_id, related_post_id)
);

CREATE INDEX IF NOT EXISTS idx_blog_related_posts_post_id ON blog_related_posts(post_id);

-- CTA blocks junction table
CREATE TABLE IF NOT EXISTS blog_cta_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    button_text VARCHAR(100) NOT NULL,
    button_url VARCHAR(500) NOT NULL,
    style VARCHAR(20) DEFAULT 'primary',
    "order" INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_blog_cta_blocks_post_id ON blog_cta_blocks(post_id);
