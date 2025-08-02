-- Create organization_admins table for many-to-many relationship between users and organizations
CREATE TABLE IF NOT EXISTS organization_admins (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- User relation
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Organization relation
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Admin permissions level
    role VARCHAR(50) DEFAULT 'admin' NOT NULL,
    
    -- Status
    is_active BOOLEAN DEFAULT true NOT NULL,
    
    -- Metadata
    assigned_by INTEGER NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    revoked_by INTEGER REFERENCES users(id),
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    -- Note: Unique constraint with WHERE clause will be added separately
    CONSTRAINT uq_user_org_active UNIQUE (user_id, organization_id)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_organization_admins_user_id ON organization_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_organization_admins_organization_id ON organization_admins(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_admins_is_active ON organization_admins(is_active);
CREATE INDEX IF NOT EXISTS idx_organization_admins_deleted_at ON organization_admins(deleted_at);

-- Update the updated_at column automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_organization_admins_updated_at BEFORE UPDATE ON organization_admins FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();