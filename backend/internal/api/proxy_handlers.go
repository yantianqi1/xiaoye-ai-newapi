package api

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google-ai-proxy/internal/db"
	"google-ai-proxy/internal/provider"
	"google-ai-proxy/internal/storage"

	"github.com/gin-gonic/gin"
)

// getNewAPIBaseURL reads the configured upstream NewAPI base URL.
func getNewAPIBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(db.GetConfig("newapi_base_url")), "/")
}

// getUserAPIKey extracts the user's API key from the X-User-Api-Key header.
func getUserAPIKey(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader("X-User-Api-Key"))
}

// newProxyHTTPClient creates an HTTP client with optional proxy support.
func newProxyHTTPClient(timeout time.Duration) *http.Client {
	proxyValue := strings.TrimSpace(os.Getenv("HTTP_PROXY"))
	if proxyValue == "" {
		return &http.Client{Timeout: timeout}
	}
	proxyURL, err := url.Parse(proxyValue)
	if err != nil {
		return &http.Client{Timeout: timeout}
	}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   timeout,
	}
}

// buildUpstreamURL constructs the full upstream URL.
func buildUpstreamURL(baseURL, path string) string {
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + path
	}
	return baseURL + "/v1" + path
}

// UnifiedGenerate is the BYOK proxy handler.
// POST /api/generate
func UnifiedGenerate(c *gin.Context) {
	userID := c.GetUint64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	apiKey := getUserAPIKey(c)
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先设置 API Key"})
		return
	}

	baseURL := getNewAPIBaseURL()
	if baseURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "平台未配置上游服务地址，请联系管理员"})
		return
	}

	user, ok := getActiveUser(c, userID)
	if !ok {
		return
	}
	_ = user

	var req UnifiedGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式无效: " + err.Error()})
		return
	}

	switch req.Type {
	case "image":
		handleProxyImageGenerate(c, userID, apiKey, baseURL, req)
	case "video":
		handleProxyVideoGenerate(c, userID, apiKey, baseURL, req)
	case "ecommerce":
		handleProxyImageGenerate(c, userID, apiKey, baseURL, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的生成类型: " + req.Type})
	}
}

// handleProxyImageGenerate proxies image generation to upstream NewAPI.
func handleProxyImageGenerate(c *gin.Context, userID uint64, apiKey, baseURL string, req UnifiedGenerateRequest) {
	userIDStr := strconv.FormatUint(userID, 10)

	model := req.Model
	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择模型"})
		return
	}

	aspectRatio, _ := req.Params["aspectRatio"].(string)
	imageSize, _ := req.Params["imageSize"].(string)
	if aspectRatio == "" {
		aspectRatio = "1:1"
	}
	if imageSize == "" {
		imageSize = "1K"
	}

	mode := "text-to-image"
	if len(req.Images) > 0 {
		mode = "image-editing"
	}

	// Create generation record first
	params := map[string]interface{}{
		"model":       model,
		"mode":        mode,
		"aspectRatio": aspectRatio,
		"imageSize":   imageSize,
	}
	genReq := CreateGenerationRequest{
		Type:            req.Type,
		Prompt:          req.Prompt,
		ReferenceImages: req.Images,
		Params:          params,
		Images:          []string{},
		Status:          "generating",
	}
	genRecord, err := CreateGeneration(userID, genReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建生成记录失败"})
		return
	}
	genID := genRecord.ID

	// Map aspect ratio to size string
	size := mapAspectRatioToSize(aspectRatio)

	// Background: proxy to upstream
	go func() {
		client := newProxyHTTPClient(300 * time.Second)

		var result *proxyImageResult
		var proxyErr error

		if isChatCompletionsModel(model) {
			// Gemini and other chat-based image models use /chat/completions
			result, proxyErr = proxyChatImageGenerate(client, baseURL, apiKey, model, req.Prompt, req.Images)
		} else if len(req.Images) > 0 || req.Mask != "" {
			result, proxyErr = proxyImageEdit(client, baseURL, apiKey, model, req.Prompt, size, req.Images, req.Mask)
		} else {
			result, proxyErr = proxyImageGeneration(client, baseURL, apiKey, model, req.Prompt, size)
		}

		if proxyErr != nil {
			log.Printf("[Proxy] 图片生成失败 [用户:%d]: %v", userID, proxyErr)
			db.DB.Model(&db.Generation{}).Where("id = ?", genID).Updates(map[string]interface{}{
				"status":     "failed",
				"error_msg":  proxyErr.Error(),
				"updated_at": time.Now(),
			})
			return
		}

		// Upload to OSS
		outputImageURL, err := storage.UploadBase64Image(result.Base64Data, userIDStr, "byok")
		if err != nil {
			log.Printf("[Proxy] 上传结果失败 [用户:%d]: %v", userID, err)
			db.DB.Model(&db.Generation{}).Where("id = ?", genID).Updates(map[string]interface{}{
				"status":     "failed",
				"error_msg":  "上传结果失败",
				"updated_at": time.Now(),
			})
			return
		}

		imagesJSON, _ := json.Marshal([]string{outputImageURL})
		db.DB.Model(&db.Generation{}).Where("id = ?", genID).Updates(map[string]interface{}{
			"images":     string(imagesJSON),
			"status":     "success",
			"updated_at": time.Now(),
		})
		log.Printf("[Proxy] 图片生成成功 [用户:%d] [记录:%d]", userID, genID)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": genID,
		"status":  "generating",
	})
}

// handleProxyVideoGenerate proxies video generation to upstream.
func handleProxyVideoGenerate(c *gin.Context, userID uint64, apiKey, baseURL string, req UnifiedGenerateRequest) {
	model := req.Model
	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择模型"})
		return
	}

	mode, _ := req.Params["mode"].(string)
	resolution, _ := req.Params["resolution"].(string)
	ratio, _ := req.Params["ratio"].(string)
	duration := 5
	if d, ok := req.Params["duration"].(float64); ok {
		duration = int(d)
	}

	if mode == "" {
		mode = "text-to-video"
	}
	if resolution == "" {
		resolution = "720p"
	}
	if ratio == "" {
		ratio = "16:9"
	}

	// 按模型的 api_type 分发:
	//   - chat : 上游把视频生成封装为 /v1/chat/completions, 响应里用 <video src=...> 返回 URL (capcut/dreamina 等)
	//   - task : 上游为 OpenAI 兼容异步任务 API /v1/video/generations (kling/vidu/jimeng/sora 等)
	// 默认 task, 保持历史行为。
	var pm db.PlatformModel
	apiType := "task"
	if err := db.DB.Where("model_id = ? AND enabled = ?", model, true).First(&pm).Error; err == nil {
		if pm.ApiType != "" {
			apiType = pm.ApiType
		}
	}

	if apiType == "chat" {
		handleChatVideoGenerate(c, userID, apiKey, baseURL, model, mode, resolution, ratio, duration, req)
		return
	}

	// 平台托管模式：当管理员已配置 newapi_api_key 且该模型解析到 "newapi" provider 时，
	// 使用 NewAPIVideoProvider（平台级 key + 后台 poller 轮询状态）。否则退回原有 BYOK 直调。
	if providerName := provider.GetProviderByModel(model); providerName == "newapi" {
		if p := provider.GetVideoProvider("newapi"); p != nil && p.IsAvailable() {
			handleNewAPIPlatformVideoGenerate(c, userID, model, mode, resolution, ratio, duration, req, p)
			return
		}
	}

	// Create generation record
	// 注意: byokApiKey 存进 params 让后台 VideoPoller 能独立轮询任务状态
	// (否则必须依赖前端保持页面打开，关页面后 30min 会被超时逻辑标 failed)
	params := map[string]interface{}{
		"model":      model,
		"mode":       mode,
		"resolution": resolution,
		"ratio":      ratio,
		"duration":   duration,
		"provider":   "byok",
		"byokApiKey": apiKey,
	}
	if ga, ok := req.Params["generate_audio"].(bool); ok {
		params["generateAudio"] = ga
	}

	// Build upstream request body (OpenAI-compatible video generation)
	// NOTE: duration 必须是 int，NewAPI TaskSubmitReq.Duration 是 int 类型，
	// 传字符串会被上游直接 400: "cannot unmarshal string into ... duration of type int"
	upstreamReq := map[string]interface{}{
		"model":    model,
		"prompt":   req.Prompt,
		"size":     mapVideoRatioToSize(ratio),
		"duration": duration,
	}

	client := newProxyHTTPClient(60 * time.Second)
	endpoint := buildUpstreamURL(baseURL, "/video/generations")

	bodyBytes, _ := json.Marshal(upstreamReq)
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "上游服务请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Proxy] 视频创建失败 status=%d body=%s", resp.StatusCode, string(respBody))
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("上游服务返回错误 (%d): %s", resp.StatusCode, string(respBody))})
		return
	}

	// Parse response to get task ID
	var videoResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &videoResp); err != nil || videoResp.ID == "" {
		// If the response contains video data directly (non-async), handle it
		c.JSON(http.StatusBadGateway, gin.H{"error": "上游返回格式异常"})
		return
	}

	taskID := videoResp.ID
	genReq := CreateGenerationRequest{
		Type:            "video",
		Prompt:          req.Prompt,
		ReferenceImages: req.Images,
		Params:          params,
		Status:          "queued",
		TaskID:          &taskID,
	}
	genRecord, err := CreateGeneration(userID, genReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存任务记录失败"})
		return
	}

	log.Printf("[Proxy] 视频任务创建成功 [用户:%d] [内部ID:%d] [上游ID:%s]", userID, genRecord.ID, taskID)

	c.JSON(http.StatusOK, gin.H{
		"task_id":          genRecord.ID,
		"provider_task_id": taskID,
		"status":           "queued",
	})
}

// --- proxy helpers ---

type proxyImageResult struct {
	Base64Data string
}

func proxyImageGeneration(client *http.Client, baseURL, apiKey, model, prompt, size string) (*proxyImageResult, error) {
	reqBody := map[string]interface{}{
		"model":           model,
		"prompt":          prompt,
		"response_format": "b64_json",
		"size":            size,
		"n":               1,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	endpoint := buildUpstreamURL(baseURL, "/images/generations")

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Proxy] 上游返回 status=%d body=%s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("上游服务返回错误 (%d)", resp.StatusCode)
	}

	return parseOpenAIImageResponse(respBody)
}

func proxyImageEdit(client *http.Client, baseURL, apiKey, model, prompt, size string, images []string, mask string) (*proxyImageResult, error) {
	// For image editing, download images to base64 then use multipart
	inputBase64s := make([]string, 0, len(images))
	for _, imgURL := range images {
		b64, err := downloadImageAsBase64(imgURL)
		if err != nil {
			return nil, fmt.Errorf("下载输入图片失败: %w", err)
		}
		inputBase64s = append(inputBase64s, b64)
	}

	var maskBase64 string
	if mask != "" {
		var err error
		maskBase64, err = downloadImageAsBase64(mask)
		if err != nil {
			return nil, fmt.Errorf("下载mask图片失败: %w", err)
		}
	}

	// Use JSON approach for edits
	reqBody := map[string]interface{}{
		"model":           model,
		"prompt":          prompt,
		"response_format": "b64_json",
		"size":            size,
		"n":               1,
		"image":           inputBase64s[0],
	}
	if maskBase64 != "" {
		reqBody["mask"] = maskBase64
	}

	bodyBytes, _ := json.Marshal(reqBody)
	endpoint := buildUpstreamURL(baseURL, "/images/edits")

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Proxy] 上游编辑返回 status=%d body=%s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("上游服务返回错误 (%d)", resp.StatusCode)
	}

	return parseOpenAIImageResponse(respBody)
}

// dataURIBase64Re matches data:image URIs embedded in markdown or plain text.
var dataURIBase64Re = regexp.MustCompile(`data:image/[a-zA-Z]+;base64,([A-Za-z0-9+/=]+)`)

func parseOpenAIImageResponse(body []byte) (*proxyImageResult, error) {
	// Try standard OpenAI JSON format first.
	var resp struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err == nil && len(resp.Data) > 0 {
		b64 := resp.Data[0].B64JSON
		if b64 == "" && resp.Data[0].URL != "" {
			data, err := downloadImageAsBase64(resp.Data[0].URL)
			if err != nil {
				return nil, fmt.Errorf("下载结果图片失败: %w", err)
			}
			b64 = data
		}
		if b64 != "" {
			if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
				return &proxyImageResult{Base64Data: b64}, nil
			}
		}
	}

	// Fallback: upstream may return non-JSON text with an embedded data URI
	// e.g. "![image](data:image/jpeg;base64,/9j/4AAQ...)"
	if m := dataURIBase64Re.FindSubmatch(body); len(m) >= 2 {
		b64 := string(m[1])
		if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
			log.Printf("[Proxy] 从非标准响应中提取到 base64 图片数据 (长度: %d)", len(b64))
			return &proxyImageResult{Base64Data: b64}, nil
		}
	}

	// Log a snippet for debugging
	snippet := string(body)
	if len(snippet) > 500 {
		snippet = snippet[:500] + "..."
	}
	return nil, fmt.Errorf("上游未返回可识别的图片数据, 响应片段: %s", snippet)
}

// isChatCompletionsModel returns true for models that generate images via
// the /chat/completions endpoint (e.g. Gemini) rather than /images/generations.
func isChatCompletionsModel(model string) bool {
	m := strings.ToLower(model)
	return strings.Contains(m, "gemini") ||
		strings.Contains(m, "gpt-4o") // gpt-4o also supports image output via chat
}

// proxyChatImageGenerate sends an image generation request through the
// /chat/completions endpoint for models like Gemini that produce images
// as part of a chat response.
func proxyChatImageGenerate(client *http.Client, baseURL, apiKey, model, prompt string, refImages []string) (*proxyImageResult, error) {
	// Build messages array
	var contentParts []map[string]interface{}

	// Add reference images first if any
	for _, imgURL := range refImages {
		contentParts = append(contentParts, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": imgURL,
			},
		})
	}

	// Add text prompt
	contentParts = append(contentParts, map[string]interface{}{
		"type": "text",
		"text": prompt,
	})

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": contentParts,
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	endpoint := buildUpstreamURL(baseURL, "/chat/completions")

	log.Printf("[Proxy] chat completions 图片生成请求: model=%s endpoint=%s", model, endpoint)

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Proxy] chat completions 上游返回 status=%d body=%s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("上游服务返回错误 (%d)", resp.StatusCode)
	}

	// Try to extract image from chat completions response
	// Response may contain inline_data with base64, or data URIs in content
	return parseChatImageResponse(respBody)
}

// parseChatImageResponse extracts image data from a chat completions response.
// Supports: OpenAI-style inline_data, data URIs in text content, and markdown images.
func parseChatImageResponse(body []byte) (*proxyImageResult, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err == nil && len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content

		// Content might be a string
		if s, ok := content.(string); ok {
			if m := dataURIBase64Re.FindStringSubmatch(s); len(m) >= 2 {
				if _, err := base64.StdEncoding.DecodeString(m[1]); err == nil {
					log.Printf("[Proxy] 从 chat completions 文本内容提取到图片 (长度: %d)", len(m[1]))
					return &proxyImageResult{Base64Data: m[1]}, nil
				}
			}
		}

		// Content might be an array of parts (multimodal response)
		if parts, ok := content.([]interface{}); ok {
			for _, part := range parts {
				p, ok := part.(map[string]interface{})
				if !ok {
					continue
				}
				// Check for inline_data (Gemini-style)
				if inlineData, ok := p["inline_data"].(map[string]interface{}); ok {
					if b64, ok := inlineData["data"].(string); ok && b64 != "" {
						if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
							log.Printf("[Proxy] 从 chat completions inline_data 提取到图片 (长度: %d)", len(b64))
							return &proxyImageResult{Base64Data: b64}, nil
						}
					}
				}
				// Check for image_url type
				if t, _ := p["type"].(string); t == "image_url" {
					if imgURL, ok := p["image_url"].(map[string]interface{}); ok {
						if urlStr, ok := imgURL["url"].(string); ok {
							// Could be a data URI
							if m := dataURIBase64Re.FindStringSubmatch(urlStr); len(m) >= 2 {
								if _, err := base64.StdEncoding.DecodeString(m[1]); err == nil {
									return &proxyImageResult{Base64Data: m[1]}, nil
								}
							}
							// Could be an HTTP URL - download it
							if strings.HasPrefix(urlStr, "http") {
								b64, err := downloadImageAsBase64(urlStr)
								if err == nil {
									return &proxyImageResult{Base64Data: b64}, nil
								}
							}
						}
					}
				}
			}
		}
	}

	// Fallback: try to find data URI anywhere in the raw response
	if m := dataURIBase64Re.FindSubmatch(body); len(m) >= 2 {
		b64 := string(m[1])
		if _, err := base64.StdEncoding.DecodeString(b64); err == nil {
			log.Printf("[Proxy] 从 chat completions 原始响应提取到图片 (长度: %d)", len(b64))
			return &proxyImageResult{Base64Data: b64}, nil
		}
	}

	snippet := string(body)
	if len(snippet) > 500 {
		snippet = snippet[:500] + "..."
	}
	return nil, fmt.Errorf("chat completions 响应中未找到图片数据, 响应片段: %s", snippet)
}

func mapAspectRatioToSize(ratio string) string {
	switch ratio {
	case "1:1":
		return "1024x1024"
	case "2:3", "3:4", "9:16":
		return "1024x1536"
	case "3:2", "4:3", "16:9":
		return "1536x1024"
	default:
		return "1024x1024"
	}
}

func mapVideoRatioToSize(ratio string) string {
	switch ratio {
	case "16:9":
		return "1920x1080"
	case "9:16":
		return "1080x1920"
	case "1:1":
		return "1080x1080"
	default:
		return "1920x1080"
	}
}

// handleNewAPIPlatformVideoGenerate 使用平台托管的 NewAPIVideoProvider 创建视频任务。
// 与 BYOK 路径的不同点：
//   - 使用 db 中配置的 newapi_api_key (而非用户 X-User-Api-Key)
//   - 生成记录 params.provider = "newapi"，由后台 StartVideoTaskPoller 自动轮询
//   - 支持首帧图 / 参考图 (透传到上游 image / images 字段)
func handleNewAPIPlatformVideoGenerate(
	c *gin.Context,
	userID uint64,
	model, mode, resolution, ratio string,
	duration int,
	req UnifiedGenerateRequest,
	p provider.VideoProvider,
) {
	generateAudio := false
	if ga, ok := req.Params["generate_audio"].(bool); ok {
		generateAudio = ga
	}

	// 首帧图 / 参考图：从 req.Images 中取 (与前端 BYOK 约定一致，可能是 URL 或 base64)
	// 这里只在字符串不含 "http" 前缀时当作 base64 透传；URL 情况交由前端处理。
	var firstFrame string
	var refImages []string
	for _, img := range req.Images {
		trimmed := strings.TrimSpace(img)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "data:") {
			// 剥离 data URL 前缀，provider 内部会重新拼
			if idx := strings.Index(trimmed, ","); idx >= 0 {
				trimmed = trimmed[idx+1:]
			}
		}
		if firstFrame == "" && (mode == "first-frame" || mode == "first-last-frame" || mode == "image-to-video") {
			firstFrame = trimmed
		} else {
			refImages = append(refImages, trimmed)
		}
	}

	vreq := provider.VideoGenerateRequest{
		Model:           model,
		Prompt:          req.Prompt,
		Mode:            mode,
		Resolution:      resolution,
		Ratio:           ratio,
		Duration:        duration,
		GenerateAudio:   generateAudio,
		FirstFrame:      firstFrame,
		ReferenceImages: refImages,
	}

	result, err := p.CreateVideoTask(vreq)
	if err != nil {
		log.Printf("[Proxy/NewAPI] 创建视频任务失败 [用户:%d]: %v", userID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	params := map[string]interface{}{
		"model":         model,
		"mode":          mode,
		"resolution":    resolution,
		"ratio":         ratio,
		"duration":      duration,
		"generateAudio": generateAudio,
		"provider":      "newapi",
	}

	genReq := CreateGenerationRequest{
		Type:            "video",
		Prompt:          req.Prompt,
		ReferenceImages: req.Images,
		Params:          params,
		Status:          "queued",
		TaskID:          &result.TaskID,
	}
	genRecord, err := CreateGeneration(userID, genReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存任务记录失败"})
		return
	}

	log.Printf("[Proxy/NewAPI] 视频任务创建成功 [用户:%d] [内部ID:%d] [上游ID:%s] [模型:%s]",
		userID, genRecord.ID, result.TaskID, model)

	c.JSON(http.StatusOK, gin.H{
		"task_id":          genRecord.ID,
		"provider_task_id": result.TaskID,
		"status":           "queued",
	})
}

// chatVideoURLRe 从 chat-completion 返回的 content 中提取视频 URL。
// 兼容以下格式:
//   <video controls="controls"> https://foo/bar.mp4 </video>
//   <video src="https://foo/bar.mp4"></video>
//   ![video](https://foo/bar.mp4)
//   直接出现的 https://...mp4 也作为最后兜底
var chatVideoURLRe = regexp.MustCompile(`https?://[^\s"'<>)]+`)

// handleChatVideoGenerate 处理 chat-completions 风格的视频模型 (同步返回 URL)
// 典型上游: newapi 的 capcut / dreamina / 封装型渠道
// 请求: POST /v1/chat/completions { model, messages:[{role:user, content:prompt}] }
// 响应: choices[0].message.content 含 <video>URL</video>
func handleChatVideoGenerate(
	c *gin.Context,
	userID uint64,
	apiKey, baseURL, model, mode, resolution, ratio string,
	duration int,
	req UnifiedGenerateRequest,
) {
	// 先落一条 generating 记录
	params := map[string]interface{}{
		"model":      model,
		"mode":       mode,
		"resolution": resolution,
		"ratio":      ratio,
		"duration":   duration,
		"provider":   "byok-chat",
	}
	if ga, ok := req.Params["generate_audio"].(bool); ok {
		params["generateAudio"] = ga
	}

	genReq := CreateGenerationRequest{
		Type:            "video",
		Prompt:          req.Prompt,
		ReferenceImages: req.Images,
		Params:          params,
		Status:          "generating",
	}
	genRecord, err := CreateGeneration(userID, genReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建生成记录失败"})
		return
	}
	genID := genRecord.ID
	userIDStr := strconv.FormatUint(userID, 10)

	// 后台 SSE 流式调用上游 chat/completions 并累积 content,最后从中抽 URL
	// 必须用 stream=true：上游真实视频生成往往 1-5 分钟，非流式连接极易被中间层切断
	go func() {
		upstreamReq := map[string]interface{}{
			"model": model,
			"messages": []map[string]interface{}{
				{"role": "user", "content": req.Prompt},
			},
			"stream": true,
		}
		bodyBytes, _ := json.Marshal(upstreamReq)
		endpoint := buildUpstreamURL(baseURL, "/chat/completions")

		log.Printf("[Proxy/ChatVideo] 请求 [用户:%d] [记录:%d] model=%s endpoint=%s",
			userID, genID, model, endpoint)

		// 流式: 不能用全局超时,否则长任务会被砍。用无超时 client + Transport 级 IdleConnTimeout
		client := &http.Client{
			Timeout: 0,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 60 * time.Second, // 等首字节最多 1 分钟
				IdleConnTimeout:       90 * time.Second,
			},
		}
		httpReq, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
		if err != nil {
			markChatVideoFailed(genID, "创建请求失败: "+err.Error())
			return
		}
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := client.Do(httpReq)
		if err != nil {
			markChatVideoFailed(genID, "上游请求失败: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("[Proxy/ChatVideo] 上游返回错误 status=%d body=%s", resp.StatusCode, string(respBody))
			markChatVideoFailed(genID, fmt.Sprintf("上游返回错误 (%d): %s", resp.StatusCode, string(respBody)))
			return
		}

		// 解析 SSE: 每行 "data: {json}",忽略 "data: [DONE]"
		// 累积所有 choices[0].delta.content 拼成完整 content
		var contentBuf strings.Builder
		// 兜底硬上限: 8 分钟没收到完整 URL 也算失败
		hardDeadline := time.Now().Add(8 * time.Minute)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			if time.Now().After(hardDeadline) {
				log.Printf("[Proxy/ChatVideo] 硬超时 [记录:%d] 已累积内容: %s", genID, truncate(contentBuf.String(), 500))
				markChatVideoFailed(genID, "上游响应超时（>8 分钟）")
				return
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "" || data == "[DONE]" {
				continue
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			if d := chunk.Choices[0].Delta.Content; d != "" {
				contentBuf.WriteString(d)
			} else if m := chunk.Choices[0].Message.Content; m != "" {
				contentBuf.WriteString(m)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[Proxy/ChatVideo] SSE 读取错误 [记录:%d]: %v 已累积: %s", genID, err, truncate(contentBuf.String(), 500))
			// 继续往下处理 —— 已累积内容里可能已经有 URL
		}

		content := contentBuf.String()
		log.Printf("[Proxy/ChatVideo] 流式累积完成 [记录:%d] 长度=%d 内容=%s", genID, len(content), truncate(content, 500))

		// 从 content 中抽取 URL —— 优先找 .mp4，其次任何 http(s) 链接
		videoURL := ""
		for _, u := range chatVideoURLRe.FindAllString(content, -1) {
			u = strings.TrimRight(u, ".,;:!?")
			if strings.Contains(strings.ToLower(u), ".mp4") || strings.Contains(u, "/video/") {
				videoURL = u
				break
			}
		}
		if videoURL == "" {
			if m := chatVideoURLRe.FindString(content); m != "" {
				videoURL = strings.TrimRight(m, ".,;:!?")
			}
		}
		if videoURL == "" {
			markChatVideoFailed(genID, "上游返回内容未包含视频 URL: "+truncate(content, 300))
			return
		}
		log.Printf("[Proxy/ChatVideo] 提取到视频 URL: %s", videoURL)

		// 下载并转存 OSS
		finalURL, err := storage.DownloadAndUploadVideo(videoURL, userIDStr, nil)
		if err != nil {
			log.Printf("[Proxy/ChatVideo] 转存失败 [记录:%d]: %v，使用原始 URL", genID, err)
			finalURL = videoURL
		}

		db.DB.Model(&db.Generation{}).Where("id = ?", genID).Updates(map[string]interface{}{
			"video_url":  finalURL,
			"status":     "success",
			"updated_at": time.Now(),
		})
		log.Printf("[Proxy/ChatVideo] 视频生成成功 [用户:%d] [记录:%d]", userID, genID)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": genID,
		"status":  "generating",
	})
}

func markChatVideoFailed(genID uint64, msg string) {
	db.DB.Model(&db.Generation{}).Where("id = ?", genID).Updates(map[string]interface{}{
		"status":     "failed",
		"error_msg":  msg,
		"updated_at": time.Now(),
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
