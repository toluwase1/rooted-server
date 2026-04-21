-- Add Telegram group tracking to conversations
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS telegram_group_id BIGINT;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS telegram_invite_link TEXT;

-- Add message source tracking for two-way sync
ALTER TABLE messages ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'miniapp';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS synced_to_telegram BOOLEAN DEFAULT FALSE;

-- Index for sync worker (find unsynced messages)
CREATE INDEX IF NOT EXISTS idx_messages_unsynced
  ON messages(synced_to_telegram, created_at)
  WHERE synced_to_telegram = FALSE;

-- Index for conversation lookup by telegram group
CREATE INDEX IF NOT EXISTS idx_conversations_telegram_group
  ON conversations(telegram_group_id)
  WHERE telegram_group_id IS NOT NULL;
