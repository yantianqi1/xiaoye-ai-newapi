package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadImageAcceptsMultipartFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldBinaryUpload := uploadBinaryImage
	oldBase64Upload := uploadBase64Image
	uploadBinaryImage = func(imageData []byte, licenseID string, directory string) (string, error) {
		if string(imageData) != "fake-png-data" {
			t.Fatalf("unexpected image data: %q", string(imageData))
		}
		if licenseID != "42" {
			t.Fatalf("unexpected license ID: %s", licenseID)
		}
		if directory != "useredit" {
			t.Fatalf("unexpected directory: %s", directory)
		}
		return "https://example.com/uploaded.png", nil
	}
	uploadBase64Image = func(string, string, string) (string, error) {
		t.Fatal("multipart upload should not use base64 path")
		return "", nil
	}
	t.Cleanup(func() {
		uploadBinaryImage = oldBinaryUpload
		uploadBase64Image = oldBase64Upload
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "sample.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("fake-png-data")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/user/upload/image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Request = req
	ctx.Set("userID", uint64(42))

	UploadImage(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp UploadImageResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.URL != "https://example.com/uploaded.png" {
		t.Fatalf("unexpected response URL: %s", resp.URL)
	}
}

func TestUploadImageAcceptsBase64JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldBinaryUpload := uploadBinaryImage
	oldBase64Upload := uploadBase64Image
	uploadBinaryImage = func([]byte, string, string) (string, error) {
		t.Fatal("json upload should not use multipart path")
		return "", nil
	}
	uploadBase64Image = func(image string, licenseID string, directory string) (string, error) {
		if image != "ZmFrZS1iYXNlNjQ=" {
			t.Fatalf("unexpected base64 payload: %s", image)
		}
		if licenseID != "7" {
			t.Fatalf("unexpected license ID: %s", licenseID)
		}
		if directory != "useredit" {
			t.Fatalf("unexpected directory: %s", directory)
		}
		return "https://example.com/uploaded-base64.png", nil
	}
	t.Cleanup(func() {
		uploadBinaryImage = oldBinaryUpload
		uploadBase64Image = oldBase64Upload
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/user/upload/image",
		bytes.NewBufferString(`{"image":"ZmFrZS1iYXNlNjQ="}`),
	)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Set("userID", uint64(7))

	UploadImage(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp UploadImageResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.URL != "https://example.com/uploaded-base64.png" {
		t.Fatalf("unexpected response URL: %s", resp.URL)
	}
}
