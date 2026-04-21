DROP INDEX IF EXISTS idx_messages_unsynced;
DROP INDEX IF EXISTS idx_conversations_telegram_group;
ALTER TABLE messages DROP COLUMN IF EXISTS source;
ALTER TABLE messages DROP COLUMN IF EXISTS synced_to_telegram;
ALTER TABLE conversations DROP COLUMN IF EXISTS telegram_group_id;
ALTER TABLE conversations DROP COLUMN IF EXISTS telegram_invite_link;
