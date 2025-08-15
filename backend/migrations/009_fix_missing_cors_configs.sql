-- Migration to fix missing CORS configurations for existing projects
-- This ensures all projects have default CORS configurations

-- Insert default CORS configurations for projects that don't have one
INSERT INTO cors_configs (project_id, allowed_origins, allowed_methods, allowed_headers, exposed_headers, allow_credentials, max_age, created_at, updated_at)
SELECT 
    p.id as project_id,
    ARRAY['*'] as allowed_origins,
    ARRAY['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'] as allowed_methods,
    ARRAY['*'] as allowed_headers,
    ARRAY[]::text[] as exposed_headers,
    false as allow_credentials,
    3600 as max_age,
    NOW() as created_at,
    NOW() as updated_at
FROM projects p
LEFT JOIN cors_configs c ON p.id = c.project_id
WHERE c.project_id IS NULL;