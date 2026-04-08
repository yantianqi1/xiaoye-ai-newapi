package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google-ai-proxy/internal/db"

	"github.com/gin-gonic/gin"
)

// ReversePromptRequest 图片反推提示词请求。
type ReversePromptRequest struct {
	Image       string `json:"image" binding:"required"`
	Language    string `json:"language,omitempty"`
	TargetModel string `json:"target_model,omitempty"`
}

// ReversePromptResponse 图片反推提示词响应。
type ReversePromptResponse struct {
	Prompt string                 `json:"prompt"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// ReversePrompt 图片反推提示词：走上游 NewAPI 的 /v1/chat/completions。
//
// 设计：
//   - 上游地址由管理员在 platform_config.newapi_base_url 配置（与其它 BYOK 路径共用）。
//   - 使用的视觉模型 ID 由管理员在 platform_config.reverse_prompt_model 配置。
//   - 鉴权使用用户自己的 API Key（请求头 X-User-Api-Key），与图像/视频 BYOK 一致。
//   - 因此本接口不再扣除平台钻石。
func ReversePrompt(c *gin.Context) {
	startTime := time.Now()
	userID := c.GetUint64("userID")
	userIDStr := strconv.FormatUint(userID, 10)

	var req ReversePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp := gin.H{"error": "invalid request"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusBadRequest, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	req.Image = strings.TrimSpace(req.Image)
	if req.Image == "" {
		resp := gin.H{"error": "image is required"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusBadRequest, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	req.Language = normalizeReversePromptLanguage(req.Language)
	req.TargetModel = strings.TrimSpace(req.TargetModel)
	if req.TargetModel == "" {
		req.TargetModel = "Nanobanana Pro"
	}

	if _, ok := getActiveUser(c, userID); !ok {
		return
	}

	apiKey := getUserAPIKey(c)
	if apiKey == "" {
		resp := gin.H{"error": "请先设置 API Key"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusBadRequest, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	baseURL := getNewAPIBaseURL()
	if baseURL == "" {
		resp := gin.H{"error": "平台未配置上游服务地址，请联系管理员"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusServiceUnavailable, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	modelID := strings.TrimSpace(db.GetConfig("reverse_prompt_model"))
	if modelID == "" {
		resp := gin.H{"error": "管理员尚未配置图片反推模型"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusServiceUnavailable, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	prompt, usage, err := callReversePromptAPI(req, baseURL, apiKey, modelID)
	if err != nil {
		log.Printf("[ReversePrompt] failed [user:%d]: %v", userID, err)
		resp := gin.H{"error": "图片反推失败，请重试"}
		logAPICall("/api/tools/reverse-prompt", nil, http.StatusBadGateway, resp, time.Since(startTime), userIDStr)
		c.JSON(http.StatusBadGateway, resp)
		return
	}

	resp := ReversePromptResponse{
		Prompt: prompt,
		Meta: map[string]interface{}{
			"provider":          "newapi",
			"model":             modelID,
			"language":          req.Language,
			"target_model":      req.TargetModel,
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
			"latency_ms":        time.Since(startTime).Milliseconds(),
		},
	}
	logAPICall("/api/tools/reverse-prompt", nil, http.StatusOK, resp, time.Since(startTime), userIDStr)
	c.JSON(http.StatusOK, resp)
}

type reversePromptChatRequest struct {
	Model    string                     `json:"model"`
	Messages []reversePromptChatMessage `json:"messages"`
}

type reversePromptChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type reversePromptChatContentPart struct {
	Type     string                     `json:"type"`
	Text     string                     `json:"text,omitempty"`
	ImageURL *reversePromptChatImageURL `json:"image_url,omitempty"`
}

type reversePromptChatImageURL struct {
	URL string `json:"url"`
}

type reversePromptChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type reversePromptUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func callReversePromptAPI(req ReversePromptRequest, baseURL, apiKey, modelID string) (string, reversePromptUsage, error) {
	var emptyUsage reversePromptUsage

	systemPrompt := buildReversePromptSystemPrompt(req.Language, req.TargetModel)

	imageURL := req.Image
	if !strings.HasPrefix(imageURL, "data:") && !strings.HasPrefix(imageURL, "http") {
		imageURL = "data:image/jpeg;base64," + imageURL
	}

	userContent := []reversePromptChatContentPart{
		{
			Type:     "image_url",
			ImageURL: &reversePromptChatImageURL{URL: imageURL},
		},
		{
			Type: "text",
			Text: "请分析这张图片，反推出可以生成该图片的提示词。",
		},
	}

	chatReq := reversePromptChatRequest{
		Model: modelID,
		Messages: []reversePromptChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
	}

	bodyBytes, err := json.Marshal(chatReq)
	if err != nil {
		return "", emptyUsage, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := buildUpstreamURL(baseURL, "/chat/completions")
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", emptyUsage, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := newProxyHTTPClient(60 * time.Second)
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", emptyUsage, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", emptyUsage, fmt.Errorf("read response: %w", err)
	}
	if httpResp.StatusCode >= http.StatusBadRequest {
		return "", emptyUsage, fmt.Errorf("upstream status=%d body=%s", httpResp.StatusCode, string(respBytes))
	}

	var parsed reversePromptChatResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", emptyUsage, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", emptyUsage, errors.New("upstream returned empty choices")
	}

	prompt := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if prompt == "" {
		return "", emptyUsage, errors.New("upstream returned empty content")
	}

	return prompt, reversePromptUsage{
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
		TotalTokens:      parsed.Usage.TotalTokens,
	}, nil
}

func buildReversePromptSystemPrompt(language, targetModel string) string {
	langInstruction := "请使用中文输出提示词。"
	if language == "en" {
		langInstruction = "Please output the prompt in English."
	}

	lines := []string{
		"你是一位专业的 AI 图像提示词反推专家。",
		"你的任务是：根据用户提供的图片，反推出能够尽可能还原该图片的 AI 绘画提示词。",
		"",
		"输出要求：",
		"1. 仅输出提示词文本，不要输出任何解释、标题、前缀或额外说明。",
		"2. 提示词应包含：主体描述、场景/背景、构图/视角、光影/色彩、艺术风格/质感。",
		"3. 按重要性排列，核心主体和动作在前，氛围和细节在后。",
		"4. 使用清晰、可执行的描述，避免模糊词汇。",
		"5. 提示词长度适中，既要覆盖画面关键元素，又不过度冗长。",
		"",
		langInstruction,
		fmt.Sprintf("目标生成模型为 %s，请按照该模型的提示词风格和最佳实践来输出。", targetModel),
	}

	return strings.Join(lines, "\n")
}

func normalizeReversePromptLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "en", "english":
		return "en"
	default:
		return "zh"
	}
}
