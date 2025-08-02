-- Organizations table for grouping projects
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Organization details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#3B82F6' NOT NULL, -- Hex color code
    
    -- Status
    is_active BOOLEAN DEFAULT true NOT NULL,
    
    -- Owner (the user who created this organization)
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Project count (will be updated via triggers)
    project_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Constraints
    CONSTRAINT uq_organization_name_active UNIQUE (name, user_id) DEFERRABLE INITIALLY DEFERRED
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_organizations_user_id ON organizations(user_id);
CREATE INDEX IF NOT EXISTS idx_organizations_is_active ON organizations(is_active);
CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);

-- Update projects table to reference organizations
ALTER TABLE projects ADD COLUMN IF NOT EXISTS organization_id INTEGER;

-- Add foreign key constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_projects_organization' 
        AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects ADD CONSTRAINT fk_projects_organization 
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Create index on organization_id in projects table
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects(organization_id);

-- Function to update project count
CREATE OR REPLACE FUNCTION update_organization_project_count()
RETURNS TRIGGER AS $$
BEGIN
    -- Update project count for old organization if any
    IF TG_OP = 'UPDATE' AND OLD.organization_id IS NOT NULL AND OLD.organization_id != NEW.organization_id THEN
        UPDATE organizations 
        SET project_count = (
            SELECT COUNT(*) FROM projects 
            WHERE organization_id = OLD.organization_id AND deleted_at IS NULL
        ) 
        WHERE id = OLD.organization_id;
    END IF;
    
    -- Update project count for new organization if any
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') AND NEW.organization_id IS NOT NULL THEN
        UPDATE organizations 
        SET project_count = (
            SELECT COUNT(*) FROM projects 
            WHERE organization_id = NEW.organization_id AND deleted_at IS NULL
        ) 
        WHERE id = NEW.organization_id;
    END IF;
    
    -- Update project count for deleted project's organization
    IF TG_OP = 'DELETE' AND OLD.organization_id IS NOT NULL THEN
        UPDATE organizations 
        SET project_count = (
            SELECT COUNT(*) FROM projects 
            WHERE organization_id = OLD.organization_id AND deleted_at IS NULL
        ) 
        WHERE id = OLD.organization_id;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers to update project count
DROP TRIGGER IF EXISTS tr_projects_org_count_insert ON projects;
CREATE TRIGGER tr_projects_org_count_insert
    AFTER INSERT ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_organization_project_count();

DROP TRIGGER IF EXISTS tr_projects_org_count_update ON projects;
CREATE TRIGGER tr_projects_org_count_update
    AFTER UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_organization_project_count();

DROP TRIGGER IF EXISTS tr_projects_org_count_delete ON projects;
CREATE TRIGGER tr_projects_org_count_delete
    AFTER UPDATE OF deleted_at ON projects
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
    EXECUTE FUNCTION update_organization_project_count();

-- Comments
COMMENT ON TABLE organizations IS 'Organizations for grouping and managing projects';
COMMENT ON COLUMN organizations.name IS 'Display name of the organization';
COMMENT ON COLUMN organizations.description IS 'Optional description of the organization';
COMMENT ON COLUMN organizations.color IS 'Hex color code for UI representation';
COMMENT ON COLUMN organizations.user_id IS 'User who owns/created this organization';
COMMENT ON COLUMN organizations.project_count IS 'Cached count of active projects in this organization';