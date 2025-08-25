-- Create default organization if it doesn't exist
INSERT INTO organizations (name, description, color, is_active, created_at, updated_at, user_id)
SELECT 
  'Default Organization',
  'Default organization for existing projects',
  '#3B82F6',
  true,
  NOW(),
  NOW(),
  (SELECT id FROM users ORDER BY id LIMIT 1)
WHERE NOT EXISTS (
  SELECT 1 FROM organizations WHERE name = 'Default Organization'
);

-- Update projects without organization to use the default organization
UPDATE projects 
SET organization_id = (SELECT id FROM organizations WHERE name = 'Default Organization' LIMIT 1)
WHERE organization_id IS NULL OR organization_id = 0;