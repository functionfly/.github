-- Migration: Collaborative Sessions (down)
-- Description: Removes collaborative session tables

DROP TABLE IF EXISTS playground_session_participants;
DROP TABLE IF EXISTS playground_sessions;