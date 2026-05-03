CREATE TABLE plans (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(80) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(255) NULL,
    billing_interval ENUM('month','year') NOT NULL DEFAULT 'month',
    amount_cents INT UNSIGNED NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    features JSON NULL,
    limits JSON NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_plans_code_interval_currency (code, billing_interval, currency),
    KEY idx_plans_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE subscriptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    plan_id BIGINT UNSIGNED NOT NULL,
    provider ENUM('stripe','mercadopago','manual') NOT NULL DEFAULT 'manual',
    provider_subscription_id VARCHAR(255) NULL,
    provider_customer_id VARCHAR(255) NULL,
    status ENUM('trialing','active','past_due','canceled','incomplete','paused') NOT NULL DEFAULT 'trialing',
    current_period_start DATETIME NULL,
    current_period_end DATETIME NULL,
    trial_ends_at DATETIME NULL,
    canceled_at DATETIME NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_subscriptions_provider_id (provider, provider_subscription_id),
    KEY idx_subscriptions_org_status (organization_id, status),
    KEY idx_subscriptions_period_end (current_period_end),
    CONSTRAINT fk_subscriptions_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_subscriptions_plan FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE invoices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    subscription_id BIGINT UNSIGNED NULL,
    provider ENUM('stripe','mercadopago','manual') NOT NULL DEFAULT 'manual',
    provider_invoice_id VARCHAR(255) NULL,
    invoice_number VARCHAR(120) NULL,
    status ENUM('draft','open','paid','void','uncollectible') NOT NULL DEFAULT 'draft',
    amount_due_cents INT UNSIGNED NOT NULL DEFAULT 0,
    amount_paid_cents INT UNSIGNED NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    issued_at DATETIME NULL,
    due_at DATETIME NULL,
    paid_at DATETIME NULL,
    metadata JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_invoices_provider_id (provider, provider_invoice_id),
    KEY idx_invoices_org_status (organization_id, status, created_at),
    KEY idx_invoices_subscription (subscription_id),
    CONSTRAINT fk_invoices_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_invoices_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE usage_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    subscription_id BIGINT UNSIGNED NULL,
    metric VARCHAR(120) NOT NULL,
    quantity DECIMAL(18,6) NOT NULL,
    recorded_at DATETIME NOT NULL,
    metadata JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_usage_records_org_metric_time (organization_id, metric, recorded_at),
    KEY idx_usage_records_subscription (subscription_id),
    CONSTRAINT fk_usage_records_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_usage_records_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CHECK (quantity >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE payment_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NULL,
    provider ENUM('stripe','mercadopago','manual') NOT NULL,
    provider_event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(120) NOT NULL,
    payload JSON NOT NULL,
    processed_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_events_provider_event (provider, provider_event_id),
    KEY idx_payment_events_org_created (organization_id, created_at),
    CONSTRAINT fk_payment_events_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
