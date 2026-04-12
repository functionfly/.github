-- Add additional blog categories
INSERT INTO blog_categories (title, slug, description, color, icon, "order") VALUES
('AI / ML', 'ai-ml', 'AI agents, FlyMind, LLMs, and machine learning features', 'indigo', 'brain', 7),
('DevOps', 'devops', 'Deployment, infrastructure, CI/CD, and operations', 'cyan', 'server', 8),
('API & SDKs', 'api-sdks', 'API documentation, SDK updates, and integration guides', 'amber', 'terminal', 9),
('Releases', 'releases', 'Changelog, version announcements, and release notes', 'emerald', 'tag', 10),
('Case Studies', 'case-studies', 'Customer stories, use cases, and success stories', 'teal', 'file-text', 11)
ON CONFLICT (slug) DO NOTHING;
