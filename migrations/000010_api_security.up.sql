CREATE TABLE api_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    allowed_ips JSONB NULL,
    rate_limit_per_minute INTEGER NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES api_clients(id) ON DELETE CASCADE,
    key_id TEXT NOT NULL UNIQUE,
    secret_hash TEXT NOT NULL,
    scopes JSONB NOT NULL,
    last_used_at TIMESTAMPTZ NULL,
    last_used_ip TEXT NULL,
    last_used_user_agent TEXT NULL,
    expires_at TIMESTAMPTZ NULL,
    status TEXT NOT NULL DEFAULT 'active',
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_client_public_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES api_clients(id) ON DELETE CASCADE,
    algorithm TEXT NOT NULL,
    public_key_raw BYTEA NOT NULL,
    fingerprint_sha256 TEXT NOT NULL UNIQUE,
    source_format TEXT NOT NULL DEFAULT 'openssh',
    status TEXT NOT NULL DEFAULT 'pending',
    activated_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_key_nonces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    nonce TEXT NOT NULL,
    timestamp_unix BIGINT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (api_key_id, nonce)
);

CREATE TABLE api_key_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NULL REFERENCES companies(id) ON DELETE SET NULL,
    client_id UUID NULL REFERENCES api_clients(id) ON DELETE SET NULL,
    api_key_id UUID NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address TEXT NULL,
    user_agent TEXT NULL,
    details JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
