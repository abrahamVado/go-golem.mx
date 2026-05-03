CREATE TABLE organizations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(180) NOT NULL,
    slug VARCHAR(90) NOT NULL,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    status ENUM('active','suspended','trialing','closed') NOT NULL DEFAULT 'active',
    logo_url VARCHAR(512) NULL,
    settings JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_organizations_slug (slug),
    KEY idx_orgs_owner (owner_user_id),
    KEY idx_orgs_status_deleted (status, deleted_at),
    CONSTRAINT fk_orgs_owner FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE organization_members (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    membership_type ENUM('owner','admin','member','guest') NOT NULL DEFAULT 'member',
    status ENUM('invited','active','suspended','removed') NOT NULL DEFAULT 'active',
    invited_by_user_id BIGINT UNSIGNED NULL,
    invited_at DATETIME NULL,
    joined_at DATETIME NULL,
    removed_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_org_members_org_user (organization_id, user_id),
    UNIQUE KEY uq_org_members_id_org (id, organization_id),
    KEY idx_org_members_user_status (user_id, status),
    KEY idx_org_members_org_status (organization_id, status),
    CONSTRAINT fk_org_members_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_org_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_org_members_invited_by FOREIGN KEY (invited_by_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
