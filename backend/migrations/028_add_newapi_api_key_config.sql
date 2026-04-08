-- Platform-managed NewAPI Bearer token, used by NewAPIVideoProvider
-- (backend/internal/provider/newapi_video.go) to call upstream
-- /v1/video/generations on behalf of all users.
INSERT IGNORE INTO platform_config (config_key, config_value) VALUES ('newapi_api_key', '');
