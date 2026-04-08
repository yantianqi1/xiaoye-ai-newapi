-- Admin-configured model ID used by /api/tools/reverse-prompt.
-- The reverse-prompt handler now calls upstream NewAPI /v1/chat/completions
-- using the user's own API key (X-User-Api-Key) and this model.
INSERT IGNORE INTO platform_config (config_key, config_value) VALUES ('reverse_prompt_model', '');
