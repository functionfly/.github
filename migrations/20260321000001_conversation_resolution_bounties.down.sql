DROP TABLE IF EXISTS conversation_bounties;
ALTER TABLE conversations
  DROP COLUMN IF EXISTS resolved_at,
  DROP COLUMN IF EXISTS resolved_by_user_id,
  DROP COLUMN IF EXISTS resolved_by_message_id;
