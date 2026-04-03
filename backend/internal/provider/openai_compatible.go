package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google-ai-proxy/internal/config"
)

const (
	OpenAICompatibleImageModelID = "openai-compatible-image"
	defaultOpenAICompatImageSize = "1K"
	defaultOpenAICompatTimeout   = 300 * time.Second
)

var openAICompatibleImageSizes = map[string]string{
	"1:1": "1024x1024",
	"2:3": "1024x1536",
	"3:2": "1536x1024",
}

type OpenAICompatibleImageProvider struct {
	baseURLGetter     func() string
	apiKeyGetter      func() string
	modelGetter       func() string
	displayNameGetter func() string
	httpClient        *http.Client
}

func init() {
	Register(OpenAICompatibleImageModelID, newConfiguredOpenAICompatibleImageProvider())
}

func newConfiguredOpenAICompatibleImageProvider() *OpenAICompatibleImageProvider {
	return &OpenAICompatibleImageProvider{
		baseURLGetter:     config.GetOpenAICompatBaseURL,
		apiKeyGetter:      config.GetOpenAICompatAPIKey,
		modelGetter:       config.GetOpenAICompatImageModel,
		displayNameGetter: config.GetOpenAICompatImageDisplayName,
		httpClient:        newOpenAICompatibleHTTPClient(),
	}
}

func newOpenAICompatibleImageProvider(baseURL, apiKey, modelName, displayName string, client *http.Client) *OpenAICompatibleImageProvider {
	if client == nil {
		client = newOpenAICompatibleHTTPClient()
	}
	return &OpenAICompatibleImageProvider{
		baseURLGetter:     func() string { return baseURL },
		apiKeyGetter:      func() string { return apiKey },
		modelGetter:       func() string { return modelName },
		displayNameGetter: func() string { return displayName },
		httpClient:        client,
	}
}

func (p *OpenAICompatibleImageProvider) ID() string {
	return OpenAICompatibleImageModelID
}

func (p *OpenAICompatibleImageProvider) Name() string {
	if displayName := strings.TrimSpace(p.displayNameGetter()); displayName != "" {
		return displayName
	}
	if modelName := strings.TrimSpace(p.modelGetter()); modelName != "" {
		return modelName
	}
	return "OpenAI Compatible Image"
}

func (p *OpenAICompatibleImageProvider) Provider() string {
	return "openai-compatible"
}

func (p *OpenAICompatibleImageProvider) IsAvailable() bool {
	return p.baseURL() != "" && p.apiKey() != "" && p.modelName() != ""
}

func (p *OpenAICompatibleImageProvider) GenerateImage(prompt string, opts ImageOptions) (*ImageResult, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("服务未配置，请联系管理员")
	}

	size, err := mapOpenAICompatibleImageSize(opts.AspectRatio, opts.ImageSize)
	if err != nil {
		return nil, err
	}
	if len(opts.InputImages) == 0 && opts.MaskImage == "" {
		return p.generateImage(prompt, size)
	}
	return p.editImage(prompt, size, opts)
}

func (p *OpenAICompatibleImageProvider) generateImage(prompt, size string) (*ImageResult, error) {
	reqBody := openAICompatibleImageRequest{
		Model:          p.modelName(),
		Prompt:         prompt,
		ResponseFormat: "b64_json",
		Size:           size,
		N:              1,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := buildOpenAICompatibleEndpoint(p.baseURL(), "/images/generations")
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey())
	httpReq.Header.Set("Content-Type", "application/json")
	return p.doRequest(httpReq)
}

func (p *OpenAICompatibleImageProvider) editImage(prompt, size string, opts ImageOptions) (*ImageResult, error) {
	if len(opts.InputImages) == 0 {
		return nil, fmt.Errorf("图片编辑至少需要一张输入图")
	}

	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)
	if err := writeEditForm(writer, p.modelName(), prompt, size, opts); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint := buildOpenAICompatibleEndpoint(p.baseURL(), "/images/edits")
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bodyBuffer)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey())
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	return p.doRequest(httpReq)
}

func (p *OpenAICompatibleImageProvider) doRequest(httpReq *http.Request) (*ImageResult, error) {
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("生成失败: %s", sanitizeHTTPError(err))
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[OpenAICompat] status=%d body=%s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("生成失败 (状态码 %d)", resp.StatusCode)
	}
	return parseOpenAICompatibleImageResponse(bodyBytes)
}

func (p *OpenAICompatibleImageProvider) baseURL() string {
	if p.baseURLGetter == nil {
		return ""
	}
	return strings.TrimSpace(p.baseURLGetter())
}

func (p *OpenAICompatibleImageProvider) apiKey() string {
	if p.apiKeyGetter == nil {
		return ""
	}
	return strings.TrimSpace(p.apiKeyGetter())
}

func (p *OpenAICompatibleImageProvider) modelName() string {
	if p.modelGetter == nil {
		return ""
	}
	return strings.TrimSpace(p.modelGetter())
}

func newOpenAICompatibleHTTPClient() *http.Client {
	proxyValue := strings.TrimSpace(os.Getenv("HTTP_PROXY"))
	if proxyValue == "" {
		return &http.Client{Timeout: defaultOpenAICompatTimeout}
	}

	proxyURL, err := url.Parse(proxyValue)
	if err != nil {
		log.Printf("[OpenAICompat] 代理地址解析失败: %v", err)
		return &http.Client{Timeout: defaultOpenAICompatTimeout}
	}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   defaultOpenAICompatTimeout,
	}
}

func mapOpenAICompatibleImageSize(aspectRatio, imageSize string) (string, error) {
	size := strings.TrimSpace(imageSize)
	if size == "" {
		size = defaultOpenAICompatImageSize
	}
	if size != defaultOpenAICompatImageSize {
		return "", fmt.Errorf("OpenAI 兼容图片模型仅支持 %s 尺寸", defaultOpenAICompatImageSize)
	}

	ratio := strings.TrimSpace(aspectRatio)
	if ratio == "" {
		ratio = "1:1"
	}
	mapped, ok := openAICompatibleImageSizes[ratio]
	if !ok {
		return "", fmt.Errorf("OpenAI 兼容图片模型不支持画幅比 %s", ratio)
	}
	return mapped, nil
}

func buildOpenAICompatibleEndpoint(baseURL, path string) string {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(normalized, path) {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + path
	}
	return normalized + "/v1" + path
}

func writeEditForm(writer *multipart.Writer, modelName, prompt, size string, opts ImageOptions) error {
	fields := map[string]string{
		"model":           modelName,
		"prompt":          prompt,
		"response_format": "b64_json",
		"size":            size,
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("write field %s: %w", key, err)
		}
	}

	for index, encoded := range opts.InputImages {
		fieldName := "image"
		if len(opts.InputImages) > 1 {
			fieldName = "image[]"
		}
		if err := writeBase64ImagePart(writer, fieldName, fmt.Sprintf("image-%d", index+1), encoded); err != nil {
			return err
		}
	}
	if opts.MaskImage != "" {
		if err := writeBase64ImagePart(writer, "mask", "mask", opts.MaskImage); err != nil {
			return err
		}
	}
	return nil
}
