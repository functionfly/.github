DROP TABLE IF EXISTS user_connectors;
DELETE FROM connectors WHERE slug IN ('github', 'notion', 'slack', 'gmail', 'linear');
DROP TABLE IF EXISTS connectors;
