package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

// downloadImageAsBase64 downloads an image from URL and returns base64 encoded data.
func downloadImageAsBase64(imageURL string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
