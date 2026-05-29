-- Community forum: Reddit-style help threads for platform users

CREATE TABLE IF NOT EXISTS community_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon VARCHAR(32) NOT NULL DEFAULT 'help-circle',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES community_categories(id) ON DELETE RESTRICT,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(300) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'solved', 'locked')),
    vote_score INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,
    view_count INT NOT NULL DEFAULT 0,
    tags TEXT[] NOT NULL DEFAULT '{}',
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    accepted_comment_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES community_comments(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    vote_score INT NOT NULL DEFAULT 0,
    is_accepted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(16) NOT NULL CHECK (target_type IN ('post', 'comment')),
    target_id UUID NOT NULL,
    value SMALLINT NOT NULL CHECK (value IN (-1, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, target_type, target_id)
);

ALTER TABLE community_posts
    ADD CONSTRAINT fk_community_posts_accepted_comment
    FOREIGN KEY (accepted_comment_id) REFERENCES community_comments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_community_posts_category ON community_posts(category_id, last_activity_at DESC);
CREATE INDEX IF NOT EXISTS idx_community_posts_hot ON community_posts(is_pinned DESC, vote_score DESC, last_activity_at DESC);
CREATE INDEX IF NOT EXISTS idx_community_posts_author ON community_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_community_comments_post ON community_comments(post_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_community_votes_target ON community_votes(target_type, target_id);

INSERT INTO community_categories (slug, name, description, icon, sort_order) VALUES
    ('getting-started', 'Getting Started', 'Account setup, first deploy, CLI install, and local dev.', 'rocket', 1),
    ('functions', 'Functions & Runtime', 'Deploy, debug, languages, cold starts, and execution issues.', 'code', 2),
    ('studio', 'Studio & Workflows', 'FRG editor, AI Composer, Graph Editor, State, and visual workflows.', 'layout', 3),
    ('agents', 'AI Agents', 'Agent deployment, wallet, memory, FlyMind, and autonomous workflows.', 'bot', 4),
    ('integrations', 'API & Integrations', 'SDK, API keys, webhooks, GitHub import, and third-party providers.', 'plug', 5),
    ('security', 'Security & Secrets', 'Vault, auth, MFA, SSO, credentials, and trust verification.', 'shield', 6),
    ('billing', 'Billing & Account', 'Plans, usage, invoices, teams, wallet, and payouts.', 'credit-card', 7),
    ('marketplace', 'Marketplace & Gallery', 'Publishing functions, buying from the gallery, and registry.', 'store', 8),
    ('troubleshooting', 'Troubleshooting', 'Errors, outages, bug reports, and how to fix common problems.', 'bug', 9),
    ('showcase', 'Show & Tell', 'Share what you built, tips, tutorials, and wins with the community.', 'sparkles', 10),
    ('feedback', 'Ideas & Feedback', 'Feature requests, product suggestions, and platform improvements.', 'lightbulb', 11),
    ('general', 'General', 'Announcements, platform news, and off-topic community chat.', 'message-square', 12)
ON CONFLICT (slug) DO NOTHING;

COMMENT ON TABLE community_posts IS 'Public community help threads (Reddit-style)';
COMMENT ON TABLE community_comments IS 'Replies on community threads, optionally nested';
COMMENT ON TABLE community_votes IS 'Upvotes and downvotes on posts and comments';
