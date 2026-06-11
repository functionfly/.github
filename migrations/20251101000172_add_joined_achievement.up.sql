-- Migration: Add "Joined FunctionFly" achievement (special badge for every member)
-- Created: 2026-03-02

INSERT INTO achievements (slug, name, description, icon, color, category, requirement_type, requirement_value, points) VALUES
    ('joined_functionfly', 'Member', 'Joined FunctionFly', 'UserPlus', 'blue', 'milestone', 'joined', 1, 0)
ON CONFLICT (slug) DO NOTHING;
