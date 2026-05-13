DROP INDEX IF EXISTS idx_users_free_expires_at;
DROP INDEX IF EXISTS idx_users_premium_expires_at;
DROP INDEX IF EXISTS idx_users_account_type;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_users_account_type;

ALTER TABLE users
    DROP COLUMN IF EXISTS blocked_at,
    DROP COLUMN IF EXISTS free_expires_at,
    DROP COLUMN IF EXISTS premium_expires_at,
    DROP COLUMN IF EXISTS account_type;
