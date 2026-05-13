ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_type TEXT NOT NULL DEFAULT 'free_client',
    ADD COLUMN IF NOT EXISTS premium_expires_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS free_expires_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS blocked_at TIMESTAMPTZ NULL;

UPDATE users
SET account_type = 'owner',
    premium_expires_at = NULL,
    free_expires_at = NULL,
    blocked_at = NULL
WHERE lower(email) = lower('admin@example.com');

ALTER TABLE users
    ADD CONSTRAINT chk_users_account_type
    CHECK (account_type IN ('owner', 'founder', 'premium_client', 'free_client', 'invalid_client'));

CREATE INDEX IF NOT EXISTS idx_users_account_type ON users(account_type);
CREATE INDEX IF NOT EXISTS idx_users_premium_expires_at ON users(premium_expires_at);
CREATE INDEX IF NOT EXISTS idx_users_free_expires_at ON users(free_expires_at);
