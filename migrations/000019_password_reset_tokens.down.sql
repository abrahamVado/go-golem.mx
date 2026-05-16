DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_company_id;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX IF EXISTS uq_password_reset_tokens_token_hash;
DROP TABLE IF EXISTS password_reset_tokens;
