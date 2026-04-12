-- Seed data for Backend-in-a-Box pricing bundles
-- These are the default bundles with "viral pricing" tactics

-- SaaS Starter Pack ($29/month) - "One click → full backend"
INSERT INTO pricing_bundles (
    id, slug, name, display_name, description, short_description,
    display_price_cents, billing_interval, stripe_price_id, icon, color,
    features_included, feature_limits, provisioning_templates,
    sort_order, is_active, is_popular, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'saas-starter',
    'SaaS Starter Pack',
    '🚀 SaaS Starter Pack',
    'Everything you need to launch a SaaS product. One click → full backend.',
    'One click → full backend',
    2900,
    'month',
    NULL, -- Set STRIPE_PRICE_SAAS_STARTER env var for actual Stripe ID
    'rocket',
    'blue',
    '["auth", "payments", "user_db", "email", "analytics"]'::jsonb,
    '{"functions": 20, "providers": 5, "requests": 2000000, "ai_calls": 5000, "storage_gb": 10, "workflows": 100}'::jsonb,
    '["auth-setup", "stripe-payments", "user-db-schema", "email-templates", "analytics-dashboard"]'::jsonb,
    1,
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO NOTHING;

-- Marketplace Pack ($49/month)
INSERT INTO pricing_bundles (
    id, slug, name, display_name, description, short_description,
    display_price_cents, billing_interval, stripe_price_id, icon, color,
    features_included, feature_limits, provisioning_templates,
    sort_order, is_active, is_popular, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'marketplace',
    'Marketplace Pack',
    '🛒 Marketplace Pack',
    'Complete marketplace backend with listings, payments, messaging, and notifications.',
    'Complete marketplace backend',
    4900,
    'month',
    NULL, -- Set STRIPE_PRICE_MARKETPLACE env var for actual Stripe ID
    'shopping-cart',
    'purple',
    '["listings", "payments", "messaging", "notifications"]'::jsonb,
    '{"functions": 30, "providers": 8, "requests": 5000000, "ai_calls": 10000, "storage_gb": 25, "workflows": 250}'::jsonb,
    '["listings-schema", "payment-escrow", "messaging-workflow", "notification-templates"]'::jsonb,
    2,
    true,
    false,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO NOTHING;

-- AI App Pack ($39/month)
INSERT INTO pricing_bundles (
    id, slug, name, display_name, description, short_description,
    display_price_cents, billing_interval, stripe_price_id, icon, color,
    features_included, feature_limits, provisioning_templates,
    sort_order, is_active, is_popular, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'ai-app',
    'AI App Pack',
    '🤖 AI App Pack',
    'AI-powered app infrastructure with vector DB, embeddings, chat workflows, and memory system.',
    'AI infrastructure in one bundle',
    3900,
    'month',
    NULL, -- Set STRIPE_PRICE_AI_APP env var for actual Stripe ID
    'brain',
    'orange',
    '["vector_db", "embeddings", "chat", "memory"]'::jsonb,
    '{"functions": 25, "providers": 5, "requests": 3000000, "ai_calls": 50000, "storage_gb": 15, "workflows": 150, "vector_searches": 100000}'::jsonb,
    '["vector-collection", "embeddings-pipeline", "chat-templates", "memory-config", "openrouter-preset"]'::jsonb,
    3,
    true,
    false,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO NOTHING;

-- Insert default deferred billing configuration
-- "Build Now, Pay Later" - 100 users OR $1000 MRR OR 90 days
INSERT INTO deferred_billing_configs (
    id,
    bundle_id,
    is_default,
    trigger_user_count,
    trigger_revenue_cents,
    trigger_days_elapsed,
    grace_period_days,
    warning_email_template,
    trigger_email_template,
    conversion_email_template,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    id,
    true,
    100,
    100000,  -- $1000 MRR
    90,      -- 3 months
    7,       -- 7-day grace period
    'founder_mode_threshold_warning',
    'founder_mode_threshold_reached',
    'founder_mode_grace_period_ending',
    NOW(),
    NOW()
FROM pricing_bundles
WHERE is_active = true
ON CONFLICT (bundle_id, is_default) DO NOTHING;
