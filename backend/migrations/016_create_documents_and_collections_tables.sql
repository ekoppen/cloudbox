-- Create collections and documents tables for dynamic data storage

-- Create collections table
CREATE TABLE IF NOT EXISTS collections (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Collection information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB DEFAULT '{}', -- JSON schema for validation
    indexes JSONB DEFAULT '[]', -- Database indexes configuration

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Statistics
    document_count BIGINT DEFAULT 0,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Unique constraint for collection name per project
    CONSTRAINT uq_collection_name_project UNIQUE (name, project_id)
);

-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(255) PRIMARY KEY, -- UUID or custom ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Collection info
    collection_name VARCHAR(255) NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Document data (JSON)
    data JSONB NOT NULL DEFAULT '{}',

    -- Metadata
    version INTEGER DEFAULT 1,
    author VARCHAR(255) -- User/API key that created/modified
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_collections_project_id ON collections(project_id);
CREATE INDEX IF NOT EXISTS idx_collections_name ON collections(name);
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_documents_collection_name ON documents(collection_name);
CREATE INDEX IF NOT EXISTS idx_documents_author ON documents(author);
CREATE INDEX IF NOT EXISTS idx_documents_version ON documents(version);

-- GIN index for JSONB data search
CREATE INDEX IF NOT EXISTS idx_documents_data_gin ON documents USING GIN (data);

-- Add triggers to update updated_at timestamp
CREATE TRIGGER update_collections_updated_at 
    BEFORE UPDATE ON collections 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at 
    BEFORE UPDATE ON documents 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update collection document count and last_modified
CREATE OR REPLACE FUNCTION update_collection_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Update document count and last_modified for the collection
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        UPDATE collections 
        SET 
            document_count = (
                SELECT COUNT(*) FROM documents 
                WHERE collection_name = NEW.collection_name 
                AND project_id = NEW.project_id 
                AND deleted_at IS NULL
            ),
            last_modified = NOW()
        WHERE name = NEW.collection_name AND project_id = NEW.project_id;
        
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE collections 
        SET 
            document_count = (
                SELECT COUNT(*) FROM documents 
                WHERE collection_name = OLD.collection_name 
                AND project_id = OLD.project_id 
                AND deleted_at IS NULL
            ),
            last_modified = NOW()
        WHERE name = OLD.collection_name AND project_id = OLD.project_id;
        
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to update collection statistics
CREATE TRIGGER tr_documents_collection_stats_insert
    AFTER INSERT ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_collection_stats();

CREATE TRIGGER tr_documents_collection_stats_update
    AFTER UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_collection_stats();

CREATE TRIGGER tr_documents_collection_stats_delete
    AFTER UPDATE OF deleted_at ON documents
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
    EXECUTE FUNCTION update_collection_stats();

-- Add comments for documentation
COMMENT ON TABLE collections IS 'Dynamic data collections (like database tables)';
COMMENT ON TABLE documents IS 'Documents stored in collections with JSON data';
COMMENT ON COLUMN collections.schema IS 'JSON schema for document validation';
COMMENT ON COLUMN collections.indexes IS 'Database indexes configuration';
COMMENT ON COLUMN documents.data IS 'Document data as JSON';
COMMENT ON COLUMN documents.collection_name IS 'Name of the collection this document belongs to';
COMMENT ON COLUMN documents.version IS 'Document version for optimistic locking';
COMMENT ON COLUMN documents.author IS 'User or API key that created/modified the document';