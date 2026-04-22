package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyChatImageGenerateEmbedsReferenceImagesAsDataURI(t *testing.T) {
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("fake-png-binary"))
	}))
	defer imageServer.Close()

	var captured struct {
		Model    string `json:"model"`
		Messages []struct {
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url,omitempty"`
			} `json:"content"`
		} `json:"messages"`
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":[{"inline_data":{"data":"ZmFrZS1pbWFnZQ=="}}]}}]}`))
	}))
	defer upstream.Close()

	result, err := proxyChatImageGenerate(http.DefaultClient, upstream.URL, "test-key", "gemini-2.5-flash-image", "make it brighter", []string{imageServer.URL + "/input.png"})
	if err != nil {
		t.Fatalf("proxyChatImageGenerate returned error: %v", err)
	}

	if result.Base64Data != "ZmFrZS1pbWFnZQ==" {
		t.Fatalf("unexpected result payload: %s", result.Base64Data)
	}
	if len(captured.Messages) != 1 || len(captured.Messages[0].Content) < 2 {
		t.Fatalf("unexpected content payload: %#v", captured.Messages)
	}

	imagePart := captured.Messages[0].Content[0]
	if imagePart.Type != "image_url" {
		t.Fatalf("expected first content part to be image_url, got %s", imagePart.Type)
	}
	if !strings.HasPrefix(imagePart.ImageURL.URL, "data:image/png;base64,") {
		t.Fatalf("expected data URI image, got %s", imagePart.ImageURL.URL)
	}
	wantBinary := []byte("fake-png-binary")
	gotBase64 := strings.TrimPrefix(imagePart.ImageURL.URL, "data:image/png;base64,")
	gotBinary, err := base64.StdEncoding.DecodeString(gotBase64)
	if err != nil {
		t.Fatalf("decode embedded base64: %v", err)
	}
	if string(gotBinary) != string(wantBinary) {
		t.Fatalf("expected embedded image %q, got %q", string(wantBinary), string(gotBinary))
	}
}
