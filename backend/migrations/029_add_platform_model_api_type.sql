-- platform_models.api_type: 上游 API 调用风格
--   'task' = OpenAI 兼容异步任务 API (/v1/video/generations + poll)，适合 kling/vidu/jimeng/sora/veo
--   'chat' = chat-completions 同步风格，响应里用 <video src=...> 返回 URL，适合 capcut/dreamina 等封装类渠道
ALTER TABLE platform_models
  ADD COLUMN api_type VARCHAR(20) NOT NULL DEFAULT 'task'
  COMMENT 'Upstream API style: task (async /video/generations) or chat (sync /chat/completions with <video> tag)';
