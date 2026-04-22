package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
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

func TestProxyImageEditUsesMultipartForm(t *testing.T) {
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("fake-png-binary"))
	}))
	defer imageServer.Close()

	var contentType string
	var promptValue string
	var modelValue string
	var responseFormat string
	var sizeValue string
	var imageCount int
	var maskCount int

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data;") {
			t.Fatalf("expected multipart content type, got %s", contentType)
		}
		if err := parseMultipartRequest(r); err != nil {
			t.Fatalf("parse multipart request: %v", err)
		}

		promptValue = r.FormValue("prompt")
		modelValue = r.FormValue("model")
		responseFormat = r.FormValue("response_format")
		sizeValue = r.FormValue("size")
		imageCount = len(r.MultipartForm.File["image"])
		maskCount = len(r.MultipartForm.File["mask"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"ZWRpdGVkLWltYWdl"}]}`))
	}))
	defer upstream.Close()

	result, err := proxyImageEdit(
		http.DefaultClient,
		upstream.URL,
		"test-key",
		"gpt-image-1",
		"replace the background",
		"1024x1024",
		[]string{imageServer.URL + "/input.png"},
		imageServer.URL+"/mask.png",
	)
	if err != nil {
		t.Fatalf("proxyImageEdit returned error: %v", err)
	}

	if promptValue != "replace the background" {
		t.Fatalf("expected prompt to be forwarded, got %q", promptValue)
	}
	if modelValue != "gpt-image-1" {
		t.Fatalf("expected model gpt-image-1, got %q", modelValue)
	}
	if responseFormat != "b64_json" {
		t.Fatalf("expected b64_json, got %q", responseFormat)
	}
	if sizeValue != "1024x1024" {
		t.Fatalf("expected size 1024x1024, got %q", sizeValue)
	}
	if imageCount != 1 {
		t.Fatalf("expected one image upload, got %d", imageCount)
	}
	if maskCount != 1 {
		t.Fatalf("expected one mask upload, got %d", maskCount)
	}
	if result.Base64Data != "ZWRpdGVkLWltYWdl" {
		t.Fatalf("unexpected result payload: %s", result.Base64Data)
	}
}

func parseMultipartRequest(r *http.Request) error {
	reader, err := r.MultipartReader()
	if err != nil {
		return err
	}
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		return err
	}
	r.MultipartForm = form
	r.Form = map[string][]string{}
	for key, values := range form.Value {
		r.Form[key] = append(r.Form[key], values...)
	}
	for key, files := range form.File {
		if len(files) == 0 {
			continue
		}
		file, err := files[0].Open()
		if err != nil {
			return err
		}
		if _, err := io.ReadAll(file); err != nil {
			file.Close()
			return err
		}
		file.Close()
		r.Form[key] = append(r.Form[key], files[0].Filename)
	}
	return nil
}
