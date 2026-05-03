INSERT INTO plans (code, name, description, billing_interval, amount_cents, currency, features, limits, is_active) VALUES
('free', 'Free', 'Starter plan for evaluation', 'month', 0, 'USD', JSON_OBJECT('projects', true, 'tasks', true), JSON_OBJECT('projects', 3, 'members', 3, 'storage_mb', 500), TRUE),
('pro', 'Pro', 'Growing teams', 'month', 2900, 'USD', JSON_OBJECT('projects', true, 'tasks', true, 'webhooks', true), JSON_OBJECT('projects', 100, 'members', 25, 'storage_mb', 10240), TRUE),
('business', 'Business', 'Advanced SaaS controls', 'month', 9900, 'USD', JSON_OBJECT('audit', true, 'api_keys', true, 'webhooks', true), JSON_OBJECT('projects', 1000, 'members', 250, 'storage_mb', 102400), TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description), features = VALUES(features), limits = VALUES(limits), is_active = VALUES(is_active);

INSERT INTO modules (id, name, description, is_system) VALUES
('projects', 'Projects', 'Project and board management', TRUE),
('files', 'Files', 'File storage and attachments', TRUE),
('webhooks', 'Webhooks', 'Outbound integration webhooks', TRUE),
('billing', 'Billing', 'Subscriptions and usage tracking', TRUE),
('api_keys', 'API Keys', 'API clients and key management', TRUE),
('chat', 'Chat', 'Conversation and message storage', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description), is_system = VALUES(is_system);
