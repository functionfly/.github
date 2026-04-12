-- Seed default blog categories
INSERT INTO blog_categories (title, slug, description, color, icon, "order") VALUES
('Engineering', 'engineering', 'Technical deep dives, architecture decisions, and engineering culture', 'blue', 'code', 1),
('Product', 'product', 'Product updates, feature announcements, and roadmap insights', 'green', 'layers', 2),
('Company News', 'company-news', 'Announcements, milestones, and company updates', 'purple', 'building', 3),
('Tutorials', 'tutorials', 'How-to guides, step-by-step tutorials, and educational content', 'orange', 'book-open', 4),
('Community', 'community', 'Community highlights, user stories, and ecosystem updates', 'pink', 'users', 5),
('Security', 'security', 'Security best practices, updates, and compliance news', 'red', 'shield', 6)
ON CONFLICT (slug) DO NOTHING;
