package api

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const defaultChatImageMimeType = "image/png"

func buildChatImageContentPart(imageRef string) (map[string]interface{}, error) {
	dataURL, err := normalizeChatImageReference(imageRef)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"type": "image_url",
		"image_url": map[string]string{
			"url": dataURL,
		},
	}, nil
}

func normalizeChatImageReference(imageRef string) (string, error) {
	value := strings.TrimSpace(imageRef)
	if value == "" {
		return "", fmt.Errorf("参考图不能为空")
	}
	if strings.HasPrefix(value, "data:image/") {
		return value, nil
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return downloadImageAsDataURI(value)
	}
	if _, err := base64.StdEncoding.DecodeString(value); err == nil {
		return "data:" + defaultChatImageMimeType + ";base64," + value, nil
	}

	return "", fmt.Errorf("不支持的参考图格式")
}
