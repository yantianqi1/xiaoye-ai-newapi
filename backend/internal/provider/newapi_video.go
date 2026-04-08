package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"google-ai-proxy/internal/db"
)

// NewAPIVideoProvider 通过上游 NewAPI (OpenAI 兼容网关) 调用其 /v1/video/generations
// 实现任意 OpenAI 兼容渠道 (kling / vidu / jimeng / sora / veo / doubao …) 的视频生成。
// 平台管理员在 platform_config 中配置:
//   - newapi_base_url   上游 NewAPI 地址 (如 https://your-newapi.com 或带 /v1)
//   - newapi_api_key    平台级 API Key (Bearer token)
// 任意在 platform_models 表中 type='video' 的模型都会走此 Provider。
type NewAPIVideoProvider struct{}

// ---- 上游请求 / 响应结构 (参考 /root/data/new-api/dto/video.go TaskSubmitReq & OpenAIVideo) ----

type newapiVideoSubmitReq struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Mode     string                 `json:"mode,omitempty"`
	Image    string                 `json:"image,omitempty"`
	Images   []string               `json:"images,omitempty"`
	Duration int                    `json:"duration,omitempty"`
	Size     string                 `json:"size,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type newapiOpenAIVideo struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"task_id,omitempty"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress,omitempty"`
	VideoURL    string                 `json:"video_url,omitempty"`
	DownloadURL string                 `json:"download_url,omitempty"`
	Error       interface{}            `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ---- 配置读取 ----

func getNewAPIBaseURLFromDB() string {
	return strings.TrimRight(strings.TrimSpace(db.GetConfig("newapi_base_url")), "/")
}

func getNewAPIKeyFromDB() string {
	return strings.TrimSpace(db.GetConfig("newapi_api_key"))
}

func newapiJoinURL(base, path string) string {
	if strings.HasSuffix(base, "/v1") {
		return base + path
	}
	return base + "/v1" + path
}

// ---- VideoProvider 接口实现 ----

func NewNewAPIVideoProvider() *NewAPIVideoProvider { return &NewAPIVideoProvider{} }

func (p *NewAPIVideoProvider) GetProviderName() string { return "newapi" }

func (p *NewAPIVideoProvider) IsAvailable() bool {
	return getNewAPIBaseURLFromDB() != "" && getNewAPIKeyFromDB() != ""
}

// GetSupportedModels 返回 DB 中 type='video' 的 platform_models。
// 不做硬编码——管理员在后台新增的任意模型都会出现在这里。
func (p *NewAPIVideoProvider) GetSupportedModels() []VideoModel {
	var rows []db.PlatformModel
	if err := db.DB.Where("enabled = ? AND type = ?", true, "video").
		Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil
	}
	models := make([]VideoModel, 0, len(rows))
	for _, r := range rows {
		models = append(models, VideoModel{
			ID:       r.ModelID,
			Name:     r.Name,
			Provider: "newapi",
		})
	}
	return models
}

// CalculateCredits 提供一个通用的钻石计算公式。不同上游渠道真实计费不同，
// 如需精确计费应在 platform_models 表中为每个模型单独记录价格。
func (p *NewAPIVideoProvider) CalculateCredits(resolution string, duration int, generateAudio bool) int {
	basePerSecond := map[string]float64{
		"480p":  8.0,
		"540p":  10.0,
		"720p":  12.0,
		"1080p": 20.0,
		"4k":    40.0,
	}
	base, ok := basePerSecond[strings.ToLower(resolution)]
	if !ok {
		base = 12.0
	}
	if duration <= 0 {
		duration = 5
	}
	credits := base * float64(duration)
	if generateAudio {
		credits *= 1.2
	}
	return int(credits)
}

// CreateVideoTask 调用上游 NewAPI 创建视频任务
func (p *NewAPIVideoProvider) CreateVideoTask(req VideoGenerateRequest) (*VideoTaskResult, error) {
	baseURL := getNewAPIBaseURLFromDB()
	apiKey := getNewAPIKeyFromDB()
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("NewAPI 未配置 (newapi_base_url / newapi_api_key)")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model 不能为空")
	}

	body := newapiVideoSubmitReq{
		Model:    req.Model,
		Prompt:   req.Prompt,
		Mode:     req.Mode,
		Duration: req.Duration,
		Size:     mapRatioResolutionToSize(req.Ratio, req.Resolution),
		Metadata: map[string]interface{}{
			"ratio":          req.Ratio,
			"resolution":     req.Resolution,
			"generate_audio": req.GenerateAudio,
		},
	}

	// 图生视频 / 参考图
	if req.FirstFrame != "" {
		body.Image = "data:image/png;base64," + req.FirstFrame
	}
	if len(req.ReferenceImages) > 0 {
		imgs := make([]string, 0, len(req.ReferenceImages))
		for _, b64 := range req.ReferenceImages {
			imgs = append(imgs, "data:image/png;base64,"+b64)
		}
		body.Images = imgs
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	endpoint := newapiJoinURL(baseURL, "/video/generations")
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[NewAPIVideo] 请求失败: %v", err)
		return nil, fmt.Errorf("上游请求失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Printf("[NewAPIVideo] 创建任务失败 status=%d body=%s", resp.StatusCode, string(raw))
		return nil, fmt.Errorf("上游返回错误 (%d): %s", resp.StatusCode, string(raw))
	}

	var parsed newapiOpenAIVideo
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body=%s", err, string(raw))
	}
	taskID := parsed.ID
	if taskID == "" {
		taskID = parsed.TaskID
	}
	if taskID == "" {
		return nil, fmt.Errorf("上游未返回 task id, body=%s", string(raw))
	}

	log.Printf("[NewAPIVideo] 任务创建成功: model=%s task_id=%s", req.Model, taskID)
	return &VideoTaskResult{
		TaskID: taskID,
		Status: mapNewAPIStatus(parsed.Status),
	}, nil
}

// GetVideoTaskStatus 查询任务状态
func (p *NewAPIVideoProvider) GetVideoTaskStatus(taskID string) (*VideoTaskStatusResponse, error) {
	baseURL := getNewAPIBaseURLFromDB()
	apiKey := getNewAPIKeyFromDB()
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("NewAPI 未配置 (newapi_base_url / newapi_api_key)")
	}

	endpoint := newapiJoinURL(baseURL, "/video/generations/"+taskID)
	httpReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("上游返回错误 (%d): %s", resp.StatusCode, string(raw))
	}

	var parsed newapiOpenAIVideo
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("解析任务状态失败: %w, body=%s", err, string(raw))
	}

	out := &VideoTaskStatusResponse{
		TaskID: taskID,
		Status: mapNewAPIStatus(parsed.Status),
	}

	// 视频下载地址：优先 video_url → download_url → /v1/videos/{id}/content
	if parsed.VideoURL != "" {
		out.VideoURL = parsed.VideoURL
	} else if parsed.DownloadURL != "" {
		out.VideoURL = parsed.DownloadURL
	} else if out.Status == "succeeded" {
		out.VideoURL = newapiJoinURL(baseURL, "/videos/"+taskID+"/content")
	}

	// 错误信息
	if parsed.Error != nil {
		switch v := parsed.Error.(type) {
		case string:
			out.ErrorMessage = v
		case map[string]interface{}:
			if msg, ok := v["message"].(string); ok {
				out.ErrorMessage = msg
			}
			if code, ok := v["code"].(string); ok {
				out.ErrorCode = code
			}
		}
	}

	rawResp := map[string]interface{}{}
	_ = json.Unmarshal(raw, &rawResp)
	out.RawResponse = rawResp
	return out, nil
}

// mapNewAPIStatus 将上游状态映射为本项目通用状态
// 上游可能返回: queued, in_progress, processing, running, completed, succeeded, failed, canceled
func mapNewAPIStatus(s string) string {
	switch strings.ToLower(s) {
	case "queued", "pending":
		return "queued"
	case "in_progress", "processing", "running":
		return "running"
	case "completed", "succeeded", "success":
		return "succeeded"
	case "failed", "error", "canceled", "cancelled":
		return "failed"
	case "":
		return "queued"
	default:
		return s
	}
}

// mapRatioResolutionToSize 根据 ratio + resolution 推算 size 字符串 (width x height)
// 上游 OpenAI 兼容接口 size 字段常用形如 "1280x720"。
func mapRatioResolutionToSize(ratio, resolution string) string {
	heights := map[string]int{
		"480p": 480, "540p": 540, "720p": 720, "1080p": 1080, "4k": 2160,
	}
	h, ok := heights[strings.ToLower(resolution)]
	if !ok {
		h = 720
	}
	// ratio -> w/h
	var rw, rh int
	switch ratio {
	case "16:9":
		rw, rh = 16, 9
	case "9:16":
		rw, rh = 9, 16
	case "1:1":
		rw, rh = 1, 1
	case "4:3":
		rw, rh = 4, 3
	case "3:4":
		rw, rh = 3, 4
	case "21:9":
		rw, rh = 21, 9
	default:
		rw, rh = 16, 9
	}
	w := h * rw / rh
	return fmt.Sprintf("%dx%d", w, h)
}

func init() {
	RegisterVideoProvider("newapi", NewNewAPIVideoProvider())
}
