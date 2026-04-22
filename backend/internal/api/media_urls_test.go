package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAbsolutizeMediaURLUsesRequestBase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Host = "assets.example.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	ctx.Request = req

	baseURL := requestPublicBaseURL(ctx)
	got := absolutizeMediaURL("/api/uploads/useredit/test.png", baseURL)
	want := "https://assets.example.com/api/uploads/useredit/test.png"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestNormalizeUnifiedGenerateRequestMediaConvertsRelativeURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("POST", "/api/generate", nil)
	req.Host = "yvhj.qtq.cc.cd"
	req.Header.Set("X-Forwarded-Proto", "https")
	ctx.Request = req

	generateReq := UnifiedGenerateRequest{
		Images: []string{"/api/uploads/useredit/input.png"},
		Mask:   "/api/uploads/useredit/mask.png",
	}

	normalizeUnifiedGenerateRequestMedia(ctx, &generateReq)

	wantImage := "https://yvhj.qtq.cc.cd/api/uploads/useredit/input.png"
	if len(generateReq.Images) != 1 || generateReq.Images[0] != wantImage {
		t.Fatalf("expected image %s, got %#v", wantImage, generateReq.Images)
	}

	wantMask := "https://yvhj.qtq.cc.cd/api/uploads/useredit/mask.png"
	if generateReq.Mask != wantMask {
		t.Fatalf("expected mask %s, got %s", wantMask, generateReq.Mask)
	}
}
