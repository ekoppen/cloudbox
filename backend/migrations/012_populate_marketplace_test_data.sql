-- Migration: Populate marketplace with test data
-- Created: 2024-08-17
-- Purpose: Add CloudBox Script Runner and sample plugins for marketplace demo

-- Add CloudBox Script Runner repository to approved repositories
INSERT INTO approved_repositories (repository_url, repository_owner, repository_name, approved_by, approval_reason, is_official, trust_level)
VALUES 
    ('https://github.com/ekoppen/cloudbox-script-runner', 'ekoppen', 'cloudbox-script-runner', 1, 'CloudBox Script Runner - official script execution plugin', true, 3)
ON CONFLICT (repository_url) DO NOTHING;

-- Add sample community repositories
INSERT INTO approved_repositories (repository_url, repository_owner, repository_name, approved_by, approval_reason, is_official, trust_level)
VALUES 
    ('https://github.com/cloudbox-community/analytics-dashboard', 'cloudbox-community', 'analytics-dashboard', 1, 'Community analytics dashboard plugin', false, 2),
    ('https://github.com/cloudbox-community/notification-center', 'cloudbox-community', 'notification-center', 1, 'Community notification management plugin', false, 2),
    ('https://github.com/trusted-dev/cloudbox-tools', 'trusted-dev', 'cloudbox-tools', 1, 'Trusted developer utility plugins', false, 2)
ON CONFLICT (repository_url) DO NOTHING;

-- Add CloudBox Script Runner to plugin registry
INSERT INTO plugin_registry (
    name, version, description, author, repository, license, type, main_file,
    checksum, signature, is_verified, is_approved,
    permissions, dependencies, ui_config,
    status, download_count, install_count,
    published_at, registry_source, source_metadata
) VALUES (
    'cloudbox-script-runner',
    '1.0.0',
    'Execute custom scripts and commands directly from your CloudBox dashboard. Supports Node.js, Python, and shell scripts with secure execution environment.',
    'Eelko Koppen',
    'https://github.com/ekoppen/cloudbox-script-runner',
    'MIT',
    'dashboard-plugin',
    'index.js',
    'a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456',
    'verified-signature-placeholder',
    true,
    true,
    ARRAY['execute', 'filesystem:read', 'filesystem:write', 'network:outbound'],
    '{"node": ">=16.0.0", "cloudbox-sdk": "^1.0.0"}',
    '{"dashboard": {"category": "Development Tools", "icon": "terminal", "color": "#22c55e"}}',
    'available',
    142,
    37,
    NOW(),
    'github',
    '{"stars": 89, "forks": 12, "issues": 3, "last_commit": "2024-08-15T10:30:00Z"}'
) ON CONFLICT (name) DO NOTHING;

-- Add Analytics Dashboard plugin
INSERT INTO plugin_registry (
    name, version, description, author, repository, license, type, main_file,
    checksum, signature, is_verified, is_approved,
    permissions, dependencies, ui_config,
    status, download_count, install_count,
    published_at, registry_source, source_metadata
) VALUES (
    'analytics-dashboard',
    '2.1.3',
    'Beautiful analytics dashboard with real-time metrics, custom charts, and data visualization. Track user engagement, performance metrics, and business KPIs.',
    'CloudBox Community',
    'https://github.com/cloudbox-community/analytics-dashboard',
    'Apache-2.0',
    'dashboard-plugin',
    'dashboard.js',
    'b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123457',
    'community-verified-signature',
    true,
    true,
    ARRAY['data:read', 'api:read', 'storage:read'],
    '{"react": "^18.0.0", "chart.js": "^4.0.0", "cloudbox-sdk": "^1.0.0"}',
    '{"dashboard": {"category": "Analytics", "icon": "chart-bar", "color": "#3b82f6"}}',
    'available',
    285,
    73,
    NOW() - INTERVAL '2 months',
    'github',
    '{"stars": 156, "forks": 34, "issues": 7, "last_commit": "2024-08-10T14:22:00Z"}'
) ON CONFLICT (name) DO NOTHING;

-- Add Notification Center plugin
INSERT INTO plugin_registry (
    name, version, description, author, repository, license, type, main_file,
    checksum, signature, is_verified, is_approved,
    permissions, dependencies, ui_config,
    status, download_count, install_count,
    published_at, registry_source, source_metadata
) VALUES (
    'notification-center',
    '1.5.2',
    'Centralized notification management system with email, SMS, and webhook support. Configure alerts, reminders, and automated notifications for your applications.',
    'NotifyTech Solutions',
    'https://github.com/cloudbox-community/notification-center',
    'MIT',
    'service-plugin',
    'notification-service.js',
    'c3d4e5f6789012345678901234567890abcdef1234567890abcdef123458',
    'verified-community-signature',
    true,
    true,
    ARRAY['email:send', 'sms:send', 'webhook:send', 'api:write', 'storage:write'],
    '{"nodemailer": "^6.9.0", "twilio": "^4.19.0", "cloudbox-sdk": "^1.0.0"}',
    '{"dashboard": {"category": "Communication", "icon": "bell", "color": "#f59e0b"}}',
    'available',
    198,
    51,
    NOW() - INTERVAL '1 month',
    'github',
    '{"stars": 203, "forks": 28, "issues": 5, "last_commit": "2024-08-12T09:15:00Z"}'
) ON CONFLICT (name) DO NOTHING;

-- Add Database Manager plugin
INSERT INTO plugin_registry (
    name, version, description, author, repository, license, type, main_file,
    checksum, signature, is_verified, is_approved,
    permissions, dependencies, ui_config,
    status, download_count, install_count,
    published_at, registry_source, source_metadata
) VALUES (
    'database-manager',
    '3.0.1',
    'Advanced database management interface with query builder, schema visualization, and data export capabilities. Supports PostgreSQL, MySQL, and MongoDB.',
    'TrustedDev',
    'https://github.com/trusted-dev/cloudbox-tools',
    'GPL-3.0',
    'dashboard-plugin',
    'db-manager.js',
    'd4e5f6789012345678901234567890abcdef1234567890abcdef123459',
    'enterprise-verified-signature',
    true,
    true,
    ARRAY['database:read', 'database:write', 'database:admin', 'export:data'],
    '{"pg": "^8.11.0", "mysql2": "^3.6.0", "mongodb": "^5.7.0", "cloudbox-sdk": "^1.0.0"}',
    '{"dashboard": {"category": "Database", "icon": "database", "color": "#8b5cf6"}}',
    'available',
    89,
    23,
    NOW() - INTERVAL '3 weeks',
    'github',
    '{"stars": 67, "forks": 15, "issues": 2, "last_commit": "2024-08-08T16:45:00Z"}'
) ON CONFLICT (name) DO NOTHING;

-- Add File Manager plugin
INSERT INTO plugin_registry (
    name, version, description, author, repository, license, type, main_file,
    checksum, signature, is_verified, is_approved,
    permissions, dependencies, ui_config,
    status, download_count, install_count,
    published_at, registry_source, source_metadata
) VALUES (
    'file-manager',
    '2.3.0',
    'Comprehensive file management system with drag-and-drop upload, bulk operations, and preview capabilities. Organize your project files with ease.',
    'CloudBox Core Team',
    'https://github.com/cloudbox/plugins',
    'MIT',
    'dashboard-plugin',
    'file-manager.js',
    'e5f6789012345678901234567890abcdef1234567890abcdef12345a',
    'official-cloudbox-signature',
    true,
    true,
    ARRAY['filesystem:read', 'filesystem:write', 'filesystem:delete', 'upload:files'],
    '{"react-dropzone": "^14.2.0", "cloudbox-sdk": "^1.0.0"}',
    '{"dashboard": {"category": "File Management", "icon": "folder", "color": "#059669"}}',
    'available',
    412,
    127,
    NOW() - INTERVAL '6 months',
    'github',
    '{"stars": 289, "forks": 45, "issues": 8, "last_commit": "2024-08-14T11:20:00Z"}'
) ON CONFLICT (name) DO NOTHING;

-- Add CloudBox Script Runner to marketplace
INSERT INTO plugin_marketplace (
    plugin_name, repository, version, description, author, category, tags, license,
    website, support_email, screenshots, demo_url, permissions, dependencies,
    pricing_model, price, currency, installation_count, rating, review_count,
    featured, status, metadata
) VALUES (
    'cloudbox-script-runner',
    'https://github.com/ekoppen/cloudbox-script-runner',
    '1.0.0',
    'Execute custom scripts and commands directly from your CloudBox dashboard. Supports Node.js, Python, and shell scripts with secure execution environment. Perfect for automation, data processing, and custom workflows.',
    'Eelko Koppen',
    'Development Tools',
    ARRAY['scripting', 'automation', 'nodejs', 'python', 'shell', 'execution', 'development'],
    'MIT',
    'https://github.com/ekoppen/cloudbox-script-runner',
    'support@cloudbox.dev',
    ARRAY[
        'https://raw.githubusercontent.com/ekoppen/cloudbox-script-runner/main/screenshots/dashboard.png',
        'https://raw.githubusercontent.com/ekoppen/cloudbox-script-runner/main/screenshots/editor.png',
        'https://raw.githubusercontent.com/ekoppen/cloudbox-script-runner/main/screenshots/output.png'
    ],
    'https://demo.cloudbox.dev/plugins/script-runner',
    ARRAY['execute', 'filesystem:read', 'filesystem:write', 'network:outbound'],
    '{"node": ">=16.0.0", "cloudbox-sdk": "^1.0.0"}',
    'free',
    0.00,
    'USD',
    37,
    4.8,
    15,
    true,
    'published',
    '{"github_stars": 89, "github_forks": 12, "last_updated": "2024-08-15T10:30:00Z", "compatibility": ["cloudbox >= 1.0.0"], "size": "2.3 MB"}'
) ON CONFLICT (plugin_name) DO NOTHING;

-- Add Analytics Dashboard to marketplace
INSERT INTO plugin_marketplace (
    plugin_name, repository, version, description, author, category, tags, license,
    website, support_email, screenshots, demo_url, permissions, dependencies,
    pricing_model, price, currency, installation_count, rating, review_count,
    featured, status, metadata
) VALUES (
    'analytics-dashboard',
    'https://github.com/cloudbox-community/analytics-dashboard',
    '2.1.3',
    'Beautiful analytics dashboard with real-time metrics, custom charts, and data visualization. Track user engagement, performance metrics, and business KPIs with interactive charts and customizable widgets.',
    'CloudBox Community',
    'Analytics',
    ARRAY['analytics', 'dashboard', 'metrics', 'charts', 'visualization', 'business-intelligence', 'kpi'],
    'Apache-2.0',
    'https://analytics-dashboard.cloudbox.community',
    'community@cloudbox.dev',
    ARRAY[
        'https://raw.githubusercontent.com/cloudbox-community/analytics-dashboard/main/screenshots/overview.png',
        'https://raw.githubusercontent.com/cloudbox-community/analytics-dashboard/main/screenshots/charts.png',
        'https://raw.githubusercontent.com/cloudbox-community/analytics-dashboard/main/screenshots/settings.png'
    ],
    'https://demo.cloudbox.dev/plugins/analytics-dashboard',
    ARRAY['data:read', 'api:read', 'storage:read'],
    '{"react": "^18.0.0", "chart.js": "^4.0.0", "cloudbox-sdk": "^1.0.0"}',
    'freemium',
    19.99,
    'USD',
    73,
    4.6,
    28,
    true,
    'published',
    '{"github_stars": 156, "github_forks": 34, "last_updated": "2024-08-10T14:22:00Z", "compatibility": ["cloudbox >= 1.0.0"], "size": "5.7 MB", "pro_features": ["advanced_charts", "custom_metrics", "data_export"]}'
) ON CONFLICT (plugin_name) DO NOTHING;

-- Add Notification Center to marketplace
INSERT INTO plugin_marketplace (
    plugin_name, repository, version, description, author, category, tags, license,
    website, support_email, screenshots, demo_url, permissions, dependencies,
    pricing_model, price, currency, installation_count, rating, review_count,
    featured, status, metadata
) VALUES (
    'notification-center',
    'https://github.com/cloudbox-community/notification-center',
    '1.5.2',
    'Centralized notification management system with email, SMS, and webhook support. Configure alerts, reminders, and automated notifications for your applications with advanced scheduling and templating.',
    'NotifyTech Solutions',
    'Communication',
    ARRAY['notifications', 'email', 'sms', 'webhooks', 'alerts', 'automation', 'messaging'],
    'MIT',
    'https://notify-tech.com/cloudbox-plugin',
    'support@notify-tech.com',
    ARRAY[
        'https://raw.githubusercontent.com/cloudbox-community/notification-center/main/screenshots/dashboard.png',
        'https://raw.githubusercontent.com/cloudbox-community/notification-center/main/screenshots/templates.png',
        'https://raw.githubusercontent.com/cloudbox-community/notification-center/main/screenshots/settings.png'
    ],
    'https://demo.cloudbox.dev/plugins/notification-center',
    ARRAY['email:send', 'sms:send', 'webhook:send', 'api:write', 'storage:write'],
    '{"nodemailer": "^6.9.0", "twilio": "^4.19.0", "cloudbox-sdk": "^1.0.0"}',
    'subscription',
    9.99,
    'USD',
    51,
    4.7,
    22,
    false,
    'published',
    '{"github_stars": 203, "github_forks": 28, "last_updated": "2024-08-12T09:15:00Z", "compatibility": ["cloudbox >= 1.0.0"], "size": "3.2 MB", "billing_cycle": "monthly"}'
) ON CONFLICT (plugin_name) DO NOTHING;

-- Add Database Manager to marketplace
INSERT INTO plugin_marketplace (
    plugin_name, repository, version, description, author, category, tags, license,
    website, support_email, screenshots, demo_url, permissions, dependencies,
    pricing_model, price, currency, installation_count, rating, review_count,
    featured, status, metadata
) VALUES (
    'database-manager',
    'https://github.com/trusted-dev/cloudbox-tools',
    '3.0.1',
    'Advanced database management interface with query builder, schema visualization, and data export capabilities. Supports PostgreSQL, MySQL, and MongoDB with enterprise-grade security and performance.',
    'TrustedDev',
    'Database',
    ARRAY['database', 'sql', 'query-builder', 'schema', 'mongodb', 'postgresql', 'mysql'],
    'GPL-3.0',
    'https://trusteddev.com/database-manager',
    'support@trusteddev.com',
    ARRAY[
        'https://raw.githubusercontent.com/trusted-dev/cloudbox-tools/main/screenshots/query-builder.png',
        'https://raw.githubusercontent.com/trusted-dev/cloudbox-tools/main/screenshots/schema-view.png',
        'https://raw.githubusercontent.com/trusted-dev/cloudbox-tools/main/screenshots/export.png'
    ],
    'https://demo.cloudbox.dev/plugins/database-manager',
    ARRAY['database:read', 'database:write', 'database:admin', 'export:data'],
    '{"pg": "^8.11.0", "mysql2": "^3.6.0", "mongodb": "^5.7.0", "cloudbox-sdk": "^1.0.0"}',
    'paid',
    49.99,
    'USD',
    23,
    4.9,
    12,
    false,
    'published',
    '{"github_stars": 67, "github_forks": 15, "last_updated": "2024-08-08T16:45:00Z", "compatibility": ["cloudbox >= 1.0.0"], "size": "8.1 MB", "enterprise_features": ["advanced_security", "audit_logs", "backup_integration"]}'
) ON CONFLICT (plugin_name) DO NOTHING;

-- Add File Manager to marketplace
INSERT INTO plugin_marketplace (
    plugin_name, repository, version, description, author, category, tags, license,
    website, support_email, screenshots, demo_url, permissions, dependencies,
    pricing_model, price, currency, installation_count, rating, review_count,
    featured, status, metadata
) VALUES (
    'file-manager',
    'https://github.com/cloudbox/plugins',
    '2.3.0',
    'Comprehensive file management system with drag-and-drop upload, bulk operations, and preview capabilities. Organize your project files with ease using our intuitive interface and powerful search functionality.',
    'CloudBox Core Team',
    'File Management',
    ARRAY['files', 'upload', 'management', 'drag-drop', 'preview', 'search', 'organization'],
    'MIT',
    'https://cloudbox.dev/plugins/file-manager',
    'support@cloudbox.dev',
    ARRAY[
        'https://raw.githubusercontent.com/cloudbox/plugins/main/file-manager/screenshots/overview.png',
        'https://raw.githubusercontent.com/cloudbox/plugins/main/file-manager/screenshots/upload.png',
        'https://raw.githubusercontent.com/cloudbox/plugins/main/file-manager/screenshots/preview.png'
    ],
    'https://demo.cloudbox.dev/plugins/file-manager',
    ARRAY['filesystem:read', 'filesystem:write', 'filesystem:delete', 'upload:files'],
    '{"react-dropzone": "^14.2.0", "cloudbox-sdk": "^1.0.0"}',
    'free',
    0.00,
    'USD',
    127,
    4.5,
    47,
    true,
    'published',
    '{"github_stars": 289, "github_forks": 45, "last_updated": "2024-08-14T11:20:00Z", "compatibility": ["cloudbox >= 1.0.0"], "size": "4.8 MB", "official_plugin": true}'
) ON CONFLICT (plugin_name) DO NOTHING;

-- Add some sample plugin installations for the demo project
INSERT INTO plugin_installations (
    plugin_name, plugin_version, project_id, status, installation_path,
    installed_by, installed_at, last_enabled_at,
    config, environment
) VALUES 
    (
        'file-manager',
        '2.3.0',
        1,
        'enabled',
        '/plugins/file-manager',
        1,
        NOW() - INTERVAL '1 week',
        NOW() - INTERVAL '1 day',
        '{"max_file_size": "100MB", "allowed_types": ["image/*", "text/*", "application/pdf"]}',
        '{"UPLOAD_DIR": "/uploads", "MAX_FILES": "1000"}'
    ),
    (
        'cloudbox-script-runner',
        '1.0.0',
        1,
        'enabled',
        '/plugins/script-runner',
        1,
        NOW() - INTERVAL '3 days',
        NOW() - INTERVAL '1 hour',
        '{"timeout": 30000, "max_memory": "512MB", "allowed_languages": ["javascript", "python"]}',
        '{"NODE_ENV": "production", "PYTHON_PATH": "/usr/bin/python3"}'
    )
ON CONFLICT (plugin_name, project_id) DO NOTHING;

-- Add corresponding plugin states
INSERT INTO plugin_states (
    plugin_name, project_id, current_status, last_health_check, health_status,
    health_details, state_changed_at, state_changed_by
) VALUES 
    (
        'file-manager',
        1,
        'running',
        NOW() - INTERVAL '5 minutes',
        'healthy',
        '{"memory_usage": "45MB", "cpu_usage": "2.1%", "uptime": 604800}',
        NOW() - INTERVAL '1 day',
        1
    ),
    (
        'cloudbox-script-runner',
        1,
        'running',
        NOW() - INTERVAL '2 minutes',
        'healthy',
        '{"memory_usage": "32MB", "cpu_usage": "1.8%", "uptime": 259200}',
        NOW() - INTERVAL '1 hour',
        1
    )
ON CONFLICT (plugin_name, project_id) DO NOTHING;

-- Add some download statistics
INSERT INTO plugin_downloads (
    plugin_name, plugin_version, project_id, user_id, download_source,
    download_status, file_size, download_time_ms, checksum_verified,
    signature_verified, started_at, completed_at
) VALUES 
    (
        'cloudbox-script-runner',
        '1.0.0',
        1,
        1,
        'https://github.com/ekoppen/cloudbox-script-runner/archive/v1.0.0.zip',
        'completed',
        2457600, -- 2.4 MB
        3450,
        true,
        true,
        NOW() - INTERVAL '3 days',
        NOW() - INTERVAL '3 days' + INTERVAL '3.45 seconds'
    ),
    (
        'file-manager',
        '2.3.0',
        1,
        1,
        'https://github.com/cloudbox/plugins/archive/file-manager-v2.3.0.zip',
        'completed',
        5033164, -- 4.8 MB
        5230,
        true,
        true,
        NOW() - INTERVAL '1 week',
        NOW() - INTERVAL '1 week' + INTERVAL '5.23 seconds'
    );

-- Update registry download and install counts based on actual data
UPDATE plugin_registry 
SET 
    download_count = (SELECT COUNT(*) FROM plugin_downloads WHERE plugin_name = plugin_registry.name AND download_status = 'completed'),
    install_count = (SELECT COUNT(*) FROM plugin_installations WHERE plugin_name = plugin_registry.name)
WHERE EXISTS (SELECT 1 FROM plugin_downloads WHERE plugin_name = plugin_registry.name);

COMMENT ON TABLE plugin_registry IS 'Registry of all available plugins with metadata and security information - populated with test data';
COMMENT ON TABLE plugin_marketplace IS 'Marketplace entries for plugins with enhanced metadata for discovery - populated with test data';
COMMENT ON TABLE approved_repositories IS 'Dynamic whitelist of approved plugin source repositories - includes test repositories';