CREATE TABLE IF NOT EXISTS community_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'conduct',
    enforcement VARCHAR(16) NOT NULL DEFAULT 'warning',
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_community_rules_active ON community_rules(is_active, sort_order);

INSERT INTO community_rules (title, description, category, enforcement, sort_order) VALUES
    ('Be respectful and constructive', 'Treat all community members with respect. No personal attacks, harassment, hate speech, or discriminatory language. Disagree with ideas, not people.', 'conduct', 'warning', 1),
    ('No spam or self-promotion', 'Do not post repetitive content, unsolicited advertisements, or affiliate links. Share your work in Show & Tell with context and value.', 'content', 'deletion', 2),
    ('Use appropriate categories and tags', 'Post in the correct category and use relevant tags. This helps others find your content and keeps the forum organized.', 'content', 'warning', 3),
    ('Search before posting', 'Check if your question has already been answered. Use the search bar and browse existing threads before creating a new one.', 'conduct', 'warning', 4),
    ('Share knowledge freely', 'Help others when you can. Share code snippets, solutions, and resources. The community grows when everyone contributes.', 'conduct', 'info', 5),
    ('No malicious content', 'Do not post malware, phishing links, credential stealers, or any content designed to harm others. Violations result in immediate suspension.', 'safety', 'suspension', 6),
    ('Respect intellectual property', 'Do not share copyrighted material without permission. Attribute sources and respect licenses when sharing code or content.', 'legal', 'deletion', 7),
    ('Report violations', 'If you see content that violates these rules, use the report function. Do not engage in public arguments with rule-breakers.', 'moderation', 'info', 8),
    ('Keep discussions on-topic', 'Stay focused on the thread topic. Off-topic comments may be moved or removed. Use General category for casual conversation.', 'content', 'warning', 9),
    ('No vote manipulation', 'Do not use multiple accounts to upvote your own content or downvote others. Artificial engagement is grounds for suspension.', 'safety', 'suspension', 10);
