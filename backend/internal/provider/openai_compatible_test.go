package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO7+XzQAAAAASUVORK5CYII="

func TestOpenAICompatibleImageProviderGenerateImageUsesOpenAIImagesEndpoint(t *testing.T) {
	var authHeader string
	var requestPath string
	var requestBody map[string]interface{}

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		authHeader = r.Header.Get("Authorization")
		requestPath = r.URL.Path

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		return jsonResponse(http.StatusOK, `{"data":[{"b64_json":"ZmFrZS1pbWFnZQ=="}]}`), nil
	})}

	provider := newOpenAICompatibleImageProvider("https://newapi.test", "test-key", "gpt-image-1", "NewAPI Image", client)

	result, err := provider.GenerateImage("make a product shot", ImageOptions{
		AspectRatio: "1:1",
		ImageSize:   "1K",
	})
	if err != nil {
		t.Fatalf("GenerateImage returned error: %v", err)
	}

	if authHeader != "Bearer test-key" {
		t.Fatalf("expected bearer auth, got %q", authHeader)
	}
	if requestPath != "/v1/images/generations" {
		t.Fatalf("expected generations path, got %q", requestPath)
	}
	if requestBody["model"] != "gpt-image-1" {
		t.Fatalf("expected upstream model gpt-image-1, got %#v", requestBody["model"])
	}
	if requestBody["response_format"] != "b64_json" {
		t.Fatalf("expected b64_json response format, got %#v", requestBody["response_format"])
	}
	if requestBody["size"] != "1024x1024" {
		t.Fatalf("expected mapped size 1024x1024, got %#v", requestBody["size"])
	}
	if result.Data != "ZmFrZS1pbWFnZQ==" {
		t.Fatalf("expected base64 image payload, got %q", result.Data)
	}
}

func TestOpenAICompatibleImageProviderGenerateImageUsesOpenAIEditsEndpoint(t *testing.T) {
	var authHeader string
	var requestPath string
	var promptValue string
	var modelValue string
	var responseFormat string
	var sizeValue string
	var imageCount int
	var maskCount int

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		authHeader = r.Header.Get("Authorization")
		requestPath = r.URL.Path

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}

		promptValue = r.FormValue("prompt")
		modelValue = r.FormValue("model")
		responseFormat = r.FormValue("response_format")
		sizeValue = r.FormValue("size")
		imageCount = len(r.MultipartForm.File["image"])
		maskCount = len(r.MultipartForm.File["mask"])

		return jsonResponse(http.StatusOK, `{"data":[{"b64_json":"ZWRpdGVkLWltYWdl"}]}`), nil
	})}

	provider := newOpenAICompatibleImageProvider("https://newapi.test", "test-key", "gpt-image-1", "NewAPI Image", client)

	result, err := provider.GenerateImage("replace the background", ImageOptions{
		AspectRatio: "3:2",
		ImageSize:   "1K",
		InputImages: []string{tinyPNGBase64},
		MaskImage:   tinyPNGBase64,
	})
	if err != nil {
		t.Fatalf("GenerateImage returned error: %v", err)
	}

	if authHeader != "Bearer test-key" {
		t.Fatalf("expected bearer auth, got %q", authHeader)
	}
	if requestPath != "/v1/images/edits" {
		t.Fatalf("expected edits path, got %q", requestPath)
	}
	if promptValue != "replace the background" {
		t.Fatalf("expected prompt to be forwarded, got %q", promptValue)
	}
	if modelValue != "gpt-image-1" {
		t.Fatalf("expected upstream model gpt-image-1, got %q", modelValue)
	}
	if responseFormat != "b64_json" {
		t.Fatalf("expected b64_json response format, got %q", responseFormat)
	}
	if sizeValue != "1536x1024" {
		t.Fatalf("expected mapped size 1536x1024, got %q", sizeValue)
	}
	if imageCount != 1 {
		t.Fatalf("expected one uploaded image, got %d", imageCount)
	}
	if maskCount != 1 {
		t.Fatalf("expected one uploaded mask, got %d", maskCount)
	}
	if result.Data != "ZWRpdGVkLWltYWdl" {
		t.Fatalf("expected base64 image payload, got %q", result.Data)
	}
}

func TestOpenAICompatibleImageProviderGenerateImageRejectsUnsupportedAspectRatio(t *testing.T) {
	provider := newOpenAICompatibleImageProvider("https://example.com", "test-key", "gpt-image-1", "NewAPI Image", http.DefaultClient)

	_, err := provider.GenerateImage("unsupported ratio", ImageOptions{
		AspectRatio: "16:9",
		ImageSize:   "1K",
	})
	if err == nil {
		t.Fatal("expected unsupported aspect ratio error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
