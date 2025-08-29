-- Add API request logs table for tracking usage statistics
CREATE TABLE IF NOT EXISTS api_request_logs (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Request details
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    full_path TEXT NOT NULL,
    user_agent TEXT,
    ip_address INET,
    
    -- Response details
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    response_size_bytes INTEGER DEFAULT 0,
    
    -- Authentication info
    api_key_id INTEGER REFERENCES api_keys(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Timing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add API route statistics summary table for faster queries
CREATE TABLE IF NOT EXISTS api_route_stats (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Route identification
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    
    -- Daily aggregated stats
    date DATE NOT NULL,
    total_requests INTEGER DEFAULT 0,
    success_requests INTEGER DEFAULT 0, -- 2xx responses
    error_requests INTEGER DEFAULT 0,   -- 4xx + 5xx responses
    avg_response_time_ms REAL DEFAULT 0,
    total_response_size_bytes BIGINT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint to prevent duplicates
    UNIQUE(project_id, method, endpoint, date)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_api_logs_project_created ON api_request_logs(project_id, created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_endpoint ON api_request_logs(project_id, endpoint, created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_status ON api_request_logs(project_id, status_code, created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_user ON api_request_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_api_key ON api_request_logs(api_key_id, created_at);

CREATE INDEX IF NOT EXISTS idx_route_stats_project_date ON api_route_stats(project_id, date);
CREATE INDEX IF NOT EXISTS idx_route_stats_endpoint ON api_route_stats(project_id, endpoint, date);

-- Add trigger to update api_route_stats when api_request_logs are inserted
CREATE OR REPLACE FUNCTION update_api_route_stats() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO api_route_stats (
        project_id, method, endpoint, date, 
        total_requests, success_requests, error_requests,
        avg_response_time_ms, total_response_size_bytes
    ) 
    VALUES (
        NEW.project_id, 
        NEW.method, 
        NEW.endpoint, 
        DATE(NEW.created_at),
        1,
        CASE WHEN NEW.status_code >= 200 AND NEW.status_code < 300 THEN 1 ELSE 0 END,
        CASE WHEN NEW.status_code >= 400 THEN 1 ELSE 0 END,
        NEW.response_time_ms,
        NEW.response_size_bytes
    )
    ON CONFLICT (project_id, method, endpoint, date)
    DO UPDATE SET
        total_requests = api_route_stats.total_requests + 1,
        success_requests = api_route_stats.success_requests + 
            CASE WHEN NEW.status_code >= 200 AND NEW.status_code < 300 THEN 1 ELSE 0 END,
        error_requests = api_route_stats.error_requests + 
            CASE WHEN NEW.status_code >= 400 THEN 1 ELSE 0 END,
        avg_response_time_ms = (
            (api_route_stats.avg_response_time_ms * (api_route_stats.total_requests - 1) + NEW.response_time_ms) 
            / api_route_stats.total_requests
        ),
        total_response_size_bytes = api_route_stats.total_response_size_bytes + NEW.response_size_bytes,
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_api_route_stats
    AFTER INSERT ON api_request_logs
    FOR EACH ROW EXECUTE FUNCTION update_api_route_stats();