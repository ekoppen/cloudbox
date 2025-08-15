-- Fix project slug unique constraint to allow reuse of deleted project names
-- This migration resolves the issue where users cannot create projects with names 
-- that were previously used by soft-deleted projects.

-- Drop existing unique constraint/index on slug
DROP INDEX IF EXISTS idx_projects_slug;

-- Remove the unique constraint from the original table creation if it exists
-- (This is safe because we're recreating it as a partial index)
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_slug_key;

-- Drop the constraint that may be created by GORM
ALTER TABLE projects DROP CONSTRAINT IF EXISTS idx_projects_slug;

-- Create a partial unique index that only enforces uniqueness for non-deleted projects
-- This allows soft-deleted projects to keep their slug while allowing new projects 
-- to reuse the same slug
CREATE UNIQUE INDEX idx_projects_slug_unique_active 
ON projects(slug) 
WHERE deleted_at IS NULL;

-- Create a regular (non-unique) index for deleted projects to maintain query performance
-- This helps with queries that need to find projects by slug regardless of deletion status
CREATE INDEX idx_projects_slug_all 
ON projects(slug);

-- Verify that our new constraint is working by checking current data
-- This will help identify any existing data issues
SELECT 
    slug, 
    COUNT(*) as total_count,
    COUNT(CASE WHEN deleted_at IS NULL THEN 1 END) as active_count,
    COUNT(CASE WHEN deleted_at IS NOT NULL THEN 1 END) as deleted_count
FROM projects 
GROUP BY slug 
HAVING COUNT(CASE WHEN deleted_at IS NULL THEN 1 END) > 1;

-- If the above query returns any rows, it means there are multiple active projects 
-- with the same slug, which would violate our new constraint. 
-- In that case, manual intervention would be needed before applying this migration.