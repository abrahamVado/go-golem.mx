CREATE TABLE api_clients (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(255) NULL,
    status ENUM('active','disabled','revoked') NOT NULL DEFAULT 'active',
    allowed_ips JSON NULL,
    rate_limit_per_minute INT UNSIGNED NULL,
    created_by_user_id BIGINT UNSIGNED NULL,
    revoked_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    PRIMARY KEY (id),
    KEY idx_api_clients_org_status (organization_id, status),
    CONSTRAINT fk_api_clients_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_api_clients_created_by FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE api_keys (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    client_id BIGINT UNSIGNED NOT NULL,
    key_id VARCHAR(40) NOT NULL,
    secret_hash CHAR(64) NOT NULL,
    scopes JSON NOT NULL,
    last_used_at DATETIME NULL,
    last_used_ip VARCHAR(64) NULL,
    last_used_user_agent VARCHAR(255) NULL,
    expires_at DATETIME NULL,
    status ENUM('active','disabled','revoked') NOT NULL DEFAULT 'active',
    revoked_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_api_keys_key_id (key_id),
    KEY idx_api_keys_client_status (client_id, status),
    KEY idx_api_keys_org_status (organization_id, status),
    CONSTRAINT fk_api_keys_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_api_keys_client FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE api_client_public_keys (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    client_id BIGINT UNSIGNED NOT NULL,
    algorithm ENUM('ed25519') NOT NULL,
    public_key_raw VARBINARY(32) NOT NULL,
    fingerprint_sha256 CHAR(64) NOT NULL,
    source_format ENUM('openssh','raw') NOT NULL DEFAULT 'openssh',
    status ENUM('pending','active','revoked') NOT NULL DEFAULT 'pending',
    activated_at DATETIME NULL,
    revoked_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_api_pubkeys_fingerprint (fingerprint_sha256),
    KEY idx_api_pubkeys_client_status (client_id, status),
    CONSTRAINT fk_api_pubkeys_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_api_pubkeys_client FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE api_key_nonces (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    api_key_id BIGINT UNSIGNED NOT NULL,
    nonce VARCHAR(96) NOT NULL,
    timestamp_unix BIGINT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_api_key_nonce (api_key_id, nonce),
    KEY idx_api_key_nonces_expires (expires_at),
    CONSTRAINT fk_api_key_nonces_key FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE api_key_audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NULL,
    client_id BIGINT UNSIGNED NULL,
    api_key_id BIGINT UNSIGNED NULL,
    event_type VARCHAR(80) NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address VARCHAR(64) NULL,
    user_agent VARCHAR(255) NULL,
    details JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_api_audit_org_created (organization_id, created_at),
    KEY idx_api_audit_key_created (api_key_id, created_at),
    KEY idx_api_audit_event (event_type, success),
    CONSTRAINT fk_api_audit_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_api_audit_client FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_api_audit_key FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
