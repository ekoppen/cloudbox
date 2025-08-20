-- Create missing tables for user authentication and app functionality

-- Create refresh_tokens table for persistent login
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Token information
    token VARCHAR(255) NOT NULL UNIQUE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,

    -- Session metadata
    ip_address INET,
    user_agent TEXT,

    -- User relation
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Create app_users table for application users (different from CloudBox admin users)
CREATE TABLE IF NOT EXISTS app_users (
    id VARCHAR(255) PRIMARY KEY, -- UUID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- User credentials
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    username VARCHAR(255),

    -- User metadata
    profile_data JSONB DEFAULT '{}',
    preferences JSONB DEFAULT '{}',

    -- Status
    is_active BOOLEAN DEFAULT true,
    is_email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_seen_at TIMESTAMP WITH TIME ZONE,

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Security
    login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP WITH TIME ZONE,
    email_verification_token VARCHAR(255)
);

-- Create app_sessions table for user sessions
CREATE TABLE IF NOT EXISTS app_sessions (
    id VARCHAR(255) PRIMARY KEY, -- UUID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Session info
    user_id VARCHAR(255) NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Session metadata
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Status
    is_active BOOLEAN DEFAULT true,
    last_activity TIMESTAMP WITH TIME ZONE
);

-- Create channel_members table for channel membership
CREATE TABLE IF NOT EXISTS channel_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Membership info
    channel_id VARCHAR(255) NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- owner, admin, member

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Status
    is_active BOOLEAN DEFAULT true,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_read_at TIMESTAMP WITH TIME ZONE,
    is_muted BOOLEAN DEFAULT false,

    -- Permissions
    can_read BOOLEAN DEFAULT true,
    can_write BOOLEAN DEFAULT true,
    can_invite BOOLEAN DEFAULT false,
    can_moderate BOOLEAN DEFAULT false,

    -- Unique constraint for user per channel
    CONSTRAINT uq_channel_member UNIQUE (channel_id, user_id)
);

-- Create message_reactions table for message reactions
CREATE TABLE IF NOT EXISTS message_reactions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Reaction info
    message_id VARCHAR(255) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    emoji VARCHAR(100) NOT NULL, -- emoji unicode or :name:

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Unique constraint for user reaction per message
    CONSTRAINT uq_user_message_emoji UNIQUE (message_id, user_id, emoji)
);

-- Create message_read table for read receipts
CREATE TABLE IF NOT EXISTS message_read (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Read receipt info
    message_id VARCHAR(255) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Unique constraint for user read per message
    CONSTRAINT uq_user_message_read UNIQUE (message_id, user_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_app_users_email ON app_users(email);
CREATE INDEX IF NOT EXISTS idx_app_users_username ON app_users(username);
CREATE INDEX IF NOT EXISTS idx_app_users_project_id ON app_users(project_id);
CREATE INDEX IF NOT EXISTS idx_app_sessions_user_id ON app_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_app_sessions_token ON app_sessions(token);
CREATE INDEX IF NOT EXISTS idx_app_sessions_project_id ON app_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_channel_id ON channel_members(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_user_id ON channel_members(user_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_project_id ON channel_members(project_id);
CREATE INDEX IF NOT EXISTS idx_message_reactions_message_id ON message_reactions(message_id);
CREATE INDEX IF NOT EXISTS idx_message_reactions_user_id ON message_reactions(user_id);
CREATE INDEX IF NOT EXISTS idx_message_reactions_project_id ON message_reactions(project_id);
CREATE INDEX IF NOT EXISTS idx_message_read_message_id ON message_read(message_id);
CREATE INDEX IF NOT EXISTS idx_message_read_user_id ON message_read(user_id);
CREATE INDEX IF NOT EXISTS idx_message_read_project_id ON message_read(project_id);

-- Add triggers to update updated_at timestamp
CREATE TRIGGER update_refresh_tokens_updated_at 
    BEFORE UPDATE ON refresh_tokens 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_app_users_updated_at 
    BEFORE UPDATE ON app_users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_app_sessions_updated_at 
    BEFORE UPDATE ON app_sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_channel_members_updated_at 
    BEFORE UPDATE ON channel_members 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_message_reactions_updated_at 
    BEFORE UPDATE ON message_reactions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_message_read_updated_at 
    BEFORE UPDATE ON message_read 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE refresh_tokens IS 'Refresh tokens for persistent user login sessions';
COMMENT ON TABLE app_users IS 'Application users (different from CloudBox admin users)';
COMMENT ON TABLE app_sessions IS 'User session tracking for application users';
COMMENT ON TABLE channel_members IS 'Channel membership and permissions';
COMMENT ON TABLE message_reactions IS 'Message reactions (emojis)';
COMMENT ON TABLE message_read IS 'Message read receipts';