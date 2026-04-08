-- Platform models: admin-managed model list for BYOK mode
CREATE TABLE IF NOT EXISTS platform_models (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  model_id    VARCHAR(100) NOT NULL UNIQUE COMMENT 'Model ID sent to upstream (e.g. gpt-image-1)',
  name        VARCHAR(100) NOT NULL COMMENT 'Display name',
  type        VARCHAR(20)  NOT NULL DEFAULT 'image' COMMENT 'image or video',
  icon_url    VARCHAR(255) DEFAULT '' COMMENT 'Icon URL (optional)',
  sort_order  INT DEFAULT 0 COMMENT 'Display order',
  enabled     BOOLEAN DEFAULT TRUE COMMENT 'Is enabled',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Platform config: global settings (e.g. NewAPI base URL)
CREATE TABLE IF NOT EXISTS platform_config (
  config_key   VARCHAR(100) NOT NULL PRIMARY KEY,
  config_value TEXT NOT NULL,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed default NewAPI base URL
INSERT IGNORE INTO platform_config (config_key, config_value) VALUES ('newapi_base_url', '');
