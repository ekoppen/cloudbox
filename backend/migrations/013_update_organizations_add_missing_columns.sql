-- Update organizations table to add missing columns from the Organization model

-- Add contact information columns
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS website VARCHAR(255);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS phone VARCHAR(100);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS contact_person VARCHAR(255);

-- Add logo and branding columns
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS logo_url TEXT;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS logo_file_id VARCHAR(255);

-- Add address information columns
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS city VARCHAR(255);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS country VARCHAR(255);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS postal_code VARCHAR(20);

-- Add foreign key constraint for logo_file_id (references files table that will be created)
-- This will be enabled after the files table is created in the next migration

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_organizations_email ON organizations(email);
CREATE INDEX IF NOT EXISTS idx_organizations_city ON organizations(city);
CREATE INDEX IF NOT EXISTS idx_organizations_country ON organizations(country);

-- Add comments for documentation
COMMENT ON COLUMN organizations.website IS 'Organization website URL';
COMMENT ON COLUMN organizations.email IS 'Organization contact email';
COMMENT ON COLUMN organizations.phone IS 'Organization contact phone number';
COMMENT ON COLUMN organizations.contact_person IS 'Organization contact person name';
COMMENT ON COLUMN organizations.logo_url IS 'Organization logo URL';
COMMENT ON COLUMN organizations.logo_file_id IS 'Reference to uploaded logo file';
COMMENT ON COLUMN organizations.address IS 'Organization physical address';
COMMENT ON COLUMN organizations.city IS 'Organization city';
COMMENT ON COLUMN organizations.country IS 'Organization country';
COMMENT ON COLUMN organizations.postal_code IS 'Organization postal/zip code';