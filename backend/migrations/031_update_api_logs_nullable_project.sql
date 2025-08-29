-- Make project_id nullable in api_request_logs for system-wide logging
ALTER TABLE api_request_logs ALTER COLUMN project_id DROP NOT NULL;