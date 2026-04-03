package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
)

func writeBase64ImagePart(writer *multipart.Writer, fieldName, filePrefix, encoded string) error {
	data, mimeType, extension, err := decodeBase64Image(encoded)
	if err != nil {
		return err
	}

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s%s"`, fieldName, filePrefix, extension))
	header.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("create multipart part: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write multipart body: %w", err)
	}
	return nil
}

func decodeBase64Image(encoded string) ([]byte, string, string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", "", fmt.Errorf("decode image base64: %w", err)
	}
	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, "", "", fmt.Errorf("input data is not an image: %s", mimeType)
	}
	return data, mimeType, mimeExtension(mimeType), nil
}

func mimeExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return filepath.Ext("." + strings.TrimPrefix(mimeType, "image/"))
	}
}

func parseOpenAICompatibleImageResponse(bodyBytes []byte) (*ImageResult, error) {
	var parsed openAICompatibleImageResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("API 响应解析失败: %v", err)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("响应中没有候选结果")
	}
	if parsed.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("响应中未返回 b64_json 图像数据")
	}
	return &ImageResult{Data: parsed.Data[0].B64JSON, MimeType: "image/png"}, nil
}

type openAICompatibleImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	ResponseFormat string `json:"response_format"`
	Size           string `json:"size"`
	N              int    `json:"n"`
}

type openAICompatibleImageResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
}
