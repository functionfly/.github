INSERT INTO community_categories (slug, name, description, icon, sort_order) VALUES
    ('support', 'Support', 'Get help with your account, deployments, configuration, and platform usage.', 'help-circle', 13)
ON CONFLICT (slug) DO NOTHING;
