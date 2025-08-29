-- Setup Default CORS Configuration for CloudBox Projects
-- This script ensures all projects have proper CORS headers and wildcard localhost support

-- Create a function to ensure proper CORS configuration for new projects
CREATE OR REPLACE FUNCTION ensure_cors_config() 
RETURNS TRIGGER AS $$
BEGIN
    -- Check if CORS config already exists for this project
    IF NOT EXISTS (SELECT 1 FROM cors_configs WHERE project_id = NEW.id) THEN
        -- Create default CORS configuration with all necessary headers
        INSERT INTO cors_configs (
            project_id,
            allowed_origins,
            allowed_methods,
            allowed_headers,
            allow_credentials,
            max_age,
            created_at,
            updated_at
        ) VALUES (
            NEW.id,
            ARRAY['http://localhost:*', 'https://localhost:*']::text[],
            ARRAY['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS']::text[],
            ARRAY['*', 'Content-Type', 'Authorization', 'X-API-Key', 'Session-Token', 'session-token', 'X-CloudBox-Client', 'x-cloudbox-client']::text[],
            false,
            3600,
            NOW(),
            NOW()
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically add CORS config for new projects
DROP TRIGGER IF EXISTS tr_projects_ensure_cors ON projects;
CREATE TRIGGER tr_projects_ensure_cors
    AFTER INSERT ON projects
    FOR EACH ROW
    EXECUTE FUNCTION ensure_cors_config();

-- Update existing CORS configurations to include all necessary headers
UPDATE cors_configs 
SET 
    allowed_headers = ARRAY['*', 'Content-Type', 'Authorization', 'X-API-Key', 'Session-Token', 'session-token', 'X-CloudBox-Client', 'x-cloudbox-client']::text[],
    updated_at = NOW()
WHERE 
    -- Only update if headers are missing Session-Token or x-cloudbox-client
    NOT ('Session-Token' = ANY(allowed_headers))
    OR NOT ('x-cloudbox-client' = ANY(allowed_headers))
    OR allowed_headers = ARRAY['*']::text[];

-- Add localhost wildcard to existing configs that don't have it
UPDATE cors_configs 
SET 
    allowed_origins = 
        CASE 
            WHEN 'http://localhost:*' = ANY(allowed_origins) THEN allowed_origins
            ELSE allowed_origins || ARRAY['http://localhost:*']::text[]
        END,
    updated_at = NOW()
WHERE 
    NOT ('http://localhost:*' = ANY(allowed_origins));

-- Add https localhost wildcard support for secure development
UPDATE cors_configs 
SET 
    allowed_origins = 
        CASE 
            WHEN 'https://localhost:*' = ANY(allowed_origins) THEN allowed_origins
            ELSE allowed_origins || ARRAY['https://localhost:*']::text[]
        END,
    updated_at = NOW()
WHERE 
    NOT ('https://localhost:*' = ANY(allowed_origins));

-- Show current CORS configurations
SELECT 
    p.id as project_id,
    p.name as project_name,
    c.allowed_origins,
    c.allowed_headers
FROM projects p
LEFT JOIN cors_configs c ON c.project_id = p.id
ORDER BY p.id;