package api

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func requestPublicBaseURL(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}

	scheme := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}

	return scheme + "://" + host
}

func absolutizeMediaURL(rawURL, baseURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" || baseURL == "" {
		return value
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.IsAbs() {
		return value
	}

	if strings.HasPrefix(value, "/") {
		return strings.TrimRight(baseURL, "/") + value
	}

	return value
}

func absolutizeMediaURLs(urls []string, baseURL string) []string {
	if len(urls) == 0 {
		return urls
	}

	normalized := make([]string, len(urls))
	for i, rawURL := range urls {
		normalized[i] = absolutizeMediaURL(rawURL, baseURL)
	}
	return normalized
}

func normalizeUnifiedGenerateRequestMedia(c *gin.Context, req *UnifiedGenerateRequest) {
	if req == nil {
		return
	}

	baseURL := requestPublicBaseURL(c)
	req.Images = absolutizeMediaURLs(req.Images, baseURL)
	req.Mask = absolutizeMediaURL(req.Mask, baseURL)
}
