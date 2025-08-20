-- Create files table for file storage system

-- First create buckets table for organizing files
CREATE TABLE IF NOT EXISTS buckets (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Bucket information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    max_file_size BIGINT DEFAULT 52428800, -- 50MB default
    allowed_types JSONB DEFAULT '[]', -- Array of MIME types
    is_public BOOLEAN DEFAULT false,

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Statistics
    file_count BIGINT DEFAULT 0,
    total_size BIGINT DEFAULT 0,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create files table
CREATE TABLE IF NOT EXISTS files (
    id VARCHAR(255) PRIMARY KEY, -- UUID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- File metadata
    original_name VARCHAR(500) NOT NULL,
    file_name VARCHAR(500) NOT NULL, -- Stored filename
    file_path TEXT NOT NULL, -- Full path on disk
    mime_type VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    checksum VARCHAR(255), -- MD5 or SHA256

    -- Storage info
    bucket_name VARCHAR(255) NOT NULL,
    folder_path TEXT, -- Path within bucket (empty for root)
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Access control
    is_public BOOLEAN DEFAULT false,
    author VARCHAR(255), -- User/API key that uploaded

    -- URLs
    public_url TEXT,
    private_url TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_buckets_project_id ON buckets(project_id);
CREATE INDEX IF NOT EXISTS idx_buckets_name ON buckets(name);
CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id);
CREATE INDEX IF NOT EXISTS idx_files_bucket_name ON files(bucket_name);
CREATE INDEX IF NOT EXISTS idx_files_folder_path ON files(folder_path);
CREATE INDEX IF NOT EXISTS idx_files_mime_type ON files(mime_type);
CREATE INDEX IF NOT EXISTS idx_files_author ON files(author);
CREATE INDEX IF NOT EXISTS idx_files_is_public ON files(is_public);

-- Add triggers to update updated_at timestamp
CREATE TRIGGER update_buckets_updated_at 
    BEFORE UPDATE ON buckets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_files_updated_at 
    BEFORE UPDATE ON files 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Now add the foreign key constraint for organizations.logo_file_id
ALTER TABLE organizations 
ADD CONSTRAINT fk_organizations_logo_file 
FOREIGN KEY (logo_file_id) REFERENCES files(id) ON DELETE SET NULL;

-- Add comments for documentation
COMMENT ON TABLE buckets IS 'File storage buckets for organizing files';
COMMENT ON TABLE files IS 'Uploaded files storage';
COMMENT ON COLUMN files.id IS 'UUID identifier for the file';
COMMENT ON COLUMN files.original_name IS 'Original filename as uploaded';
COMMENT ON COLUMN files.file_name IS 'Stored filename on disk';
COMMENT ON COLUMN files.file_path IS 'Full path to file on storage system';
COMMENT ON COLUMN files.bucket_name IS 'Bucket where file is stored';
COMMENT ON COLUMN files.folder_path IS 'Folder path within bucket';
COMMENT ON COLUMN files.author IS 'User or API key that uploaded the file';