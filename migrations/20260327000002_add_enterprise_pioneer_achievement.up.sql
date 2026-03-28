-- Add Enterprise Pioneer achievement for users who upgrade to Enterprise tier
-- This is a rare/legendary achievement that <1% of users will earn

INSERT INTO achievements (
    id,
    slug,
    name,
    description,
    icon,
    color,
    category,
    requirement_type,
    requirement_value,
    points,
    is_hidden,
    created_at
) VALUES (
    gen_random_uuid(),
    'enterprise_pioneer',
    'Enterprise Pioneer',
    'Upgraded to Enterprise tier - elite member with unlimited access, dedicated support, and premium features. Less than 1% of users achieve this status.',
    'Crown',
    '#8b5cf6',
    'milestone',
    'plan_upgrade',
    1,
    500,
    false,
    NOW()
)
ON CONFLICT (slug) DO NOTHING;

-- Add index on slug if it doesn't exist for faster achievement lookups
CREATE INDEX IF NOT EXISTS idx_achievements_slug ON achievements(slug);
