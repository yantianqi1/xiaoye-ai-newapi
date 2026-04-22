package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google-ai-proxy/internal/storage"
)

const (
	imageUploadDirectory = "useredit"
	maxImageUploadBytes  = 10 * 1024 * 1024
)

var (
	uploadBinaryImage = storage.UploadImageData
	uploadBase64Image = storage.UploadBase64Image
)

type requestError struct {
	message string
}

func (e requestError) Error() string {
	return e.message
}

// UploadImage 上传图片到 OSS，返回 URL
func UploadImage(c *gin.Context) {
	userID := c.GetUint64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	userIDStr := strconv.FormatUint(userID, 10)
	url, err := uploadUserImage(c, userIDStr)
	if err != nil {
		respondUploadImageError(c, userID, err)
		return
	}
	url = absolutizeMediaURL(url, requestPublicBaseURL(c))

	c.JSON(http.StatusOK, UploadImageResponse{URL: url})
}

func uploadUserImage(c *gin.Context, userID string) (string, error) {
	if c.ContentType() == "multipart/form-data" {
		return uploadMultipartImage(c, userID)
	}
	return uploadJSONImage(c, userID)
}

func uploadMultipartImage(c *gin.Context, userID string) (string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", requestError{message: "缺少图片文件"}
	}
	if file.Size <= 0 {
		return "", requestError{message: "图片文件为空"}
	}
	if file.Size > maxImageUploadBytes {
		return "", requestError{message: "图片文件过大"}
	}

	src, err := file.Open()
	if err != nil {
		return "", requestError{message: "读取图片文件失败"}
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	return uploadBinaryImage(imageData, userID, imageUploadDirectory)
}

func uploadJSONImage(c *gin.Context, userID string) (string, error) {
	var req UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", requestError{message: "请求格式无效"}
	}

	return uploadBase64Image(req.Image, userID, imageUploadDirectory)
}

func respondUploadImageError(c *gin.Context, userID uint64, err error) {
	var reqErr requestError
	if errors.As(err, &reqErr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": reqErr.message})
		return
	}

	log.Printf("上传图片失败 [用户:%d]: %v", userID, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "上传图片失败"})
}

// UploadVideo uploads a user provided video file to OSS and returns public URL.
func UploadVideo(c *gin.Context) {
	userID := c.GetUint64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "please login first"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	if file.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}
	if file.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video file too large"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".mp4", ".mov", ".webm", ".m4v":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported video format"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	defer src.Close()

	videoData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read video data"})
		return
	}

	userIDStr := strconv.FormatUint(userID, 10)
	url, err := storage.UploadVideoData(videoData, userIDStr, ext)
	if err != nil {
		log.Printf("upload video failed [user:%d]: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video"})
		return
	}
	url = absolutizeMediaURL(url, requestPublicBaseURL(c))

	c.JSON(http.StatusOK, UploadImageResponse{URL: url})
}
