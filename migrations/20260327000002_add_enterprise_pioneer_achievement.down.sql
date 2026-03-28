-- Remove Enterprise Pioneer achievement

DELETE FROM user_achievements
WHERE achievement_id IN (
    SELECT id FROM achievements WHERE slug = 'enterprise_pioneer'
);

DELETE FROM achievements
WHERE slug = 'enterprise_pioneer';
