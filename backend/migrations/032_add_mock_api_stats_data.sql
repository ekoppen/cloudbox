-- Add some mock API statistics data for testing
INSERT INTO api_request_logs (project_id, method, endpoint, full_path, status_code, response_time_ms, response_size_bytes, created_at) VALUES
-- Project 6 (PhotoPortfolio) mock data
(6, 'GET', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 401, 150, 45, NOW() - INTERVAL '1 hour'),
(6, 'GET', '/p/{id}/api/discovery/routes', '/p/6/api/discovery/routes', 401, 75, 32, NOW() - INTERVAL '2 hours'),
(6, 'GET', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 401, 120, 45, NOW() - INTERVAL '3 hours'),
(6, 'POST', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 401, 200, 78, NOW() - INTERVAL '4 hours'),
(6, 'GET', '/p/{id}/api/discovery/routes', '/p/6/api/discovery/routes', 401, 90, 32, NOW() - INTERVAL '5 hours'),
-- Add some successful requests (mock data as if authenticated)
(6, 'GET', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 200, 85, 1024, NOW() - INTERVAL '6 hours'),
(6, 'POST', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 201, 180, 2048, NOW() - INTERVAL '7 hours'),
(6, 'GET', '/p/{id}/api/discovery/routes', '/p/6/api/discovery/routes', 200, 45, 512, NOW() - INTERVAL '8 hours'),
-- Yesterday's data
(6, 'GET', '/p/{id}/api/storage/images/files', '/p/6/api/storage/images/files', 200, 95, 1536, NOW() - INTERVAL '25 hours'),
(6, 'GET', '/p/{id}/api/discovery/routes', '/p/6/api/discovery/routes', 200, 55, 256, NOW() - INTERVAL '26 hours');

-- Add aggregated route stats (this would normally be done by the trigger)
INSERT INTO api_route_stats (project_id, method, endpoint, date, total_requests, success_requests, error_requests, avg_response_time_ms, total_response_size_bytes) VALUES
-- Today's stats
(6, 'GET', '/p/{id}/api/storage/images/files', CURRENT_DATE, 4, 2, 2, 137.5, 2650),
(6, 'GET', '/p/{id}/api/discovery/routes', CURRENT_DATE, 3, 1, 2, 70, 320),
(6, 'POST', '/p/{id}/api/storage/images/files', CURRENT_DATE, 2, 1, 1, 190, 2126),
-- Yesterday's stats
(6, 'GET', '/p/{id}/api/storage/images/files', CURRENT_DATE - INTERVAL '1 day', 1, 1, 0, 95, 1536),
(6, 'GET', '/p/{id}/api/discovery/routes', CURRENT_DATE - INTERVAL '1 day', 1, 1, 0, 55, 256);