-- Backend-in-a-Box pricing bundles
CREATE TABLE IF NOT EXISTS pricing_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL, -- 'saas-starter', 'marketplace', 'ai-app'
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL, -- e.g., "SaaS Starter Pack"
    description TEXT,
    short_description VARCHAR(200), -- Tagline like "One click → full backend"
    display_price_cents INTEGER NOT NULL, -- 2900 for $29
    billing_interval VARCHAR(20) NOT NULL DEFAULT 'month', -- 'month', 'year'
    stripe_price_id VARCHAR(100), -- Stripe price ID for checkout
    icon VARCHAR(50) NOT NULL DEFAULT 'rocket', -- 'rocket', 'shopping-cart', 'brain', 'zap'
    color VARCHAR(20) DEFAULT 'blue', -- 'blue', 'purple', 'orange' for UI theming
    
    -- Features included in this bundle
    features_included JSONB NOT NULL DEFAULT '[]', -- ['auth', 'payments', 'user_db', 'email', 'analytics']
    
    -- Resource limits for the bundle
    feature_limits JSONB NOT NULL DEFAULT '{}', -- {functions: 10, requests: 1000000, ...}
    
    -- Sort order for display
    sort_order INTEGER DEFAULT 0,
    
    -- Auto-provisioning templates to apply
    provisioning_templates JSONB DEFAULT '[]', -- ['auth-setup', 'stripe-payments', 'user-db-schema']
    
    is_active BOOLEAN DEFAULT true,
    is_popular BOOLEAN DEFAULT false, -- Show "Most Popular" badge
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Deferred billing configuration (Build Now, Pay Later)
CREATE TABLE IF NOT EXISTS deferred_billing_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id UUID NOT NULL REFERENCES pricing_bundles(id) ON DELETE CASCADE,
    
    -- Which bundle does this apply to
    is_default BOOLEAN DEFAULT false,
    
    -- Trigger conditions (OR logic - any trigger can start billing)
    trigger_user_count INTEGER, -- Number of users (NULL = disabled)
    trigger_revenue_cents INTEGER, -- MRR in cents (NULL = disabled)
    trigger_api_calls INTEGER, -- API call volume per month (NULL = disabled)
    trigger_days_elapsed INTEGER, -- Days since signup (NULL = disabled)
    
    -- Default: 100 users OR $1000 MRR OR 90 days
    
    -- Grace period after trigger hit
    grace_period_days INTEGER DEFAULT 7,
    
    -- What happens when triggered
    convert_to_bundle_id UUID REFERENCES pricing_bundles(id), -- NULL = same bundle
    
    -- Email notification templates
    warning_email_template VARCHAR(50), -- 80% threshold
    trigger_email_template VARCHAR(50), -- 100% threshold
    conversion_email_template VARCHAR(50), -- Grace period ending
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(bundle_id, is_default)
);

-- Founder mode registrations
CREATE TABLE IF NOT EXISTS founder_mode_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    bundle_id UUID NOT NULL REFERENCES pricing_bundles(id),
    
    -- Founder mode type
    mode_type VARCHAR(20) NOT NULL DEFAULT 'time_based', -- 'time_based', 'revenue_based', 'hybrid'
    
    -- Time-based founder mode
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ends_at TIMESTAMP WITH TIME ZONE, -- NULL if revenue-based or hybrid
    free_days INTEGER DEFAULT 90, -- 3 months default
    
    -- Revenue-based founder mode
    mrr_threshold_cents INTEGER DEFAULT 100000, -- $1000 MRR default
    mrr_reached_at TIMESTAMP WITH TIME ZONE,
    
    -- Current status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'grace_period', 'converted', 'expired', 'canceled'
    
    -- Conversion tracking
    converted_to_bundle_id UUID REFERENCES pricing_bundles(id),
    converted_at TIMESTAMP WITH TIME ZONE,
    stripe_subscription_id VARCHAR(100),
    
    -- Grace period (if applicable)
    grace_period_started_at TIMESTAMP WITH TIME ZONE,
    grace_period_ends_at TIMESTAMP WITH TIME ZONE,
    
    -- Usage tracking for deferred billing
    max_users_seen INTEGER DEFAULT 0,
    max_mrr_seen_cents INTEGER DEFAULT 0,
    max_api_calls_monthly INTEGER DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id, bundle_id, status) -- Only one active founder mode per tenant per bundle
);

-- Bundle subscriptions (separate from standard subscriptions)
CREATE TABLE IF NOT EXISTS bundle_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    bundle_id UUID NOT NULL REFERENCES pricing_bundles(id),
    founder_mode_id UUID REFERENCES founder_mode_registrations(id),
    
    -- Was converted from founder mode
    converted_from_founder_mode BOOLEAN DEFAULT false,
    
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'canceled', 'past_due'
    stripe_subscription_id VARCHAR(100),
    
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    
    cancel_at_period_end BOOLEAN DEFAULT false,
    canceled_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id, bundle_id, status)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_pricing_bundles_slug ON pricing_bundles(slug);
CREATE INDEX IF NOT EXISTS idx_pricing_bundles_active ON pricing_bundles(is_active);
CREATE INDEX IF NOT EXISTS idx_pricing_bundles_sort ON pricing_bundles(sort_order);

CREATE INDEX IF NOT EXISTS idx_deferred_billing_bundle ON deferred_billing_configs(bundle_id);
CREATE INDEX IF NOT EXISTS idx_deferred_billing_default ON deferred_billing_configs(is_default);

CREATE INDEX IF NOT EXISTS idx_founder_mode_tenant ON founder_mode_registrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_founder_mode_status ON founder_mode_registrations(status);
CREATE INDEX IF NOT EXISTS idx_founder_mode_bundle ON founder_mode_registrations(bundle_id);

CREATE INDEX IF NOT EXISTS idx_bundle_subs_tenant ON bundle_subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bundle_subs_status ON bundle_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_bundle_subs_bundle ON bundle_subscriptions(bundle_id);
