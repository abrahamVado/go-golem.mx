CREATE TABLE whitelist_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    company TEXT NULL,
    subject TEXT NULL,
    message TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'landing_page',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_whitelist_requests_email ON whitelist_requests(email);
CREATE INDEX idx_whitelist_requests_status ON whitelist_requests(status);
CREATE INDEX idx_whitelist_requests_created_at ON whitelist_requests(created_at DESC);
