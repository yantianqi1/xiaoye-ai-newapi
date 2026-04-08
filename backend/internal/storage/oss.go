package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var client *oss.Client
var bucket *oss.Bucket
var bucketName string
var publicDomain string

// localStorageDir is the fallback local directory when OSS is not configured.
const localStorageDir = "/app/uploads"

// useLocalStorage indicates whether to fall back to local file storage.
var useLocalStorage bool

// InitOSS initializes the Aliyun OSS client using environment variable credentials
func InitOSS() error {
	endpoint := os.Getenv("OSS_ENDPOINT")
	region := os.Getenv("OSS_REGION")
	bucketNameEnv := os.Getenv("OSS_BUCKET_NAME")
	publicDomainEnv := os.Getenv("OSS_PUBLIC_DOMAIN") // Optional: custom domain for public URLs

	if endpoint == "" || region == "" || bucketNameEnv == "" {
		return fmt.Errorf("缺少必要的 OSS 配置: OSS_ENDPOINT, OSS_REGION, OSS_BUCKET_NAME")
	}

	bucketName = bucketNameEnv

	// Clean up publicDomain - remove any protocol prefix
	publicDomain = publicDomainEnv
	if publicDomain != "" {
		publicDomain = strings.TrimPrefix(publicDomain, "https://")
		publicDomain = strings.TrimPrefix(publicDomain, "http://")
		publicDomain = strings.TrimRight(publicDomain, "/")
	}

	// Get credentials from environment variables
	accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")

	if accessKeyID == "" || accessKeySecret == "" {
		return fmt.Errorf("缺少必要的 OSS 凭证: OSS_ACCESS_KEY_ID=%s, OSS_ACCESS_KEY_SECRET=%s",
			func() string {
				if accessKeyID == "" {
					return "未设置"
				}
				return "已设置"
			}(),
			func() string {
				if accessKeySecret == "" {
					return "未设置"
				}
				return "已设置"
			}())
	}

	// Create OSS client with credentials and endpoint
	var err error
	client, err = oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return fmt.Errorf("初始化 OSS 客户端失败: %v", err)
	}

	bucket, err = client.Bucket(bucketName)
	if err != nil {
		return fmt.Errorf("获取 OSS bucket 失败: %v", err)
	}

	log.Printf("OSS 初始化成功: bucket=%s, region=%s, endpoint=%s", bucketName, region, endpoint)
	return nil
}

// buildPublicURL builds the public accessible URL for an object
func buildPublicURL(objectKey string) string {
	// If custom domain is configured, use it
	if publicDomain != "" {
		return fmt.Sprintf("https://%s/%s", publicDomain, objectKey)
	}

	// Otherwise build from bucket and endpoint
	// Endpoint format: oss-cn-hangzhou.aliyuncs.com
	// URL should be: https://bucket.oss-cn-hangzhou.aliyuncs.com/objectKey
	endpoint := os.Getenv("OSS_ENDPOINT")
	if endpoint != "" && !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://") {
		return fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, objectKey)
	}

	// If endpoint has protocol, strip it and build URL
	cleanEndpoint := strings.TrimPrefix(endpoint, "https://")
	cleanEndpoint = strings.TrimPrefix(cleanEndpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", bucketName, cleanEndpoint, objectKey)
}

// InitLocalStorage sets up local file storage as fallback.
func InitLocalStorage() {
	if err := os.MkdirAll(localStorageDir, 0755); err != nil {
		log.Printf("warning: failed to create local storage dir %s: %v", localStorageDir, err)
		return
	}
	useLocalStorage = true
	log.Printf("本地文件存储已启用: %s", localStorageDir)
}

// LocalStorageDir returns the local uploads directory path for serving static files.
func LocalStorageDir() string {
	return localStorageDir
}

// uploadLocal saves binary data to a local file and returns a URL path.
func uploadLocal(data []byte, directory, ext string) (string, error) {
	subDir := fmt.Sprintf("%s/%s", localStorageDir, directory)
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	timestamp := time.Now().UnixNano() / 1e6
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	filename := fmt.Sprintf("%d_%s%s", timestamp, hex.EncodeToString(randBytes), ext)
	filePath := fmt.Sprintf("%s/%s", subDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	// Return a path that will be served by the backend at /api/uploads/...
	urlPath := fmt.Sprintf("/api/uploads/%s/%s", directory, filename)
	log.Printf("本地文件保存成功: %s", urlPath)
	return urlPath, nil
}

// UploadImageData uploads image data (binary) to OSS or local storage.
func UploadImageData(imageData []byte, licenseID string, directory string) (string, error) {
	if bucket == nil && !useLocalStorage {
		return "", fmt.Errorf("OSS 客户端未初始化且本地存储未启用")
	}

	if bucket == nil {
		return uploadLocal(imageData, directory, ".png")
	}

	// Generate unique filename with random suffix to prevent collision
	timestamp := time.Now().UnixNano() / 1e6 // milliseconds
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	objectKey := fmt.Sprintf("%s/%d_%s.png", directory, timestamp, hex.EncodeToString(randBytes))

	// Upload to OSS
	err := bucket.PutObject(objectKey, bytes.NewReader(imageData))
	if err != nil {
		log.Printf("上传图像到 OSS 失败 [用户:%s]: %v", licenseID, err)
		return "", fmt.Errorf("上传失败: %v", err)
	}

	// Get public URL
	publicURL := buildPublicURL(objectKey)
	log.Printf("图像上传成功: %s -> %s", objectKey, publicURL)
	return publicURL, nil
}

// UploadBase64Image uploads base64 encoded image to OSS
// base64Data: base64 encoded image data
// licenseID: user license ID
// directory: target directory (e.g., "banana" or "useredit")
// returns: public URL of the uploaded image
func UploadBase64Image(base64Data string, licenseID string, directory string) (string, error) {
	if bucket == nil && !useLocalStorage {
		return "", fmt.Errorf("OSS 客户端未初始化且本地存储未启用")
	}

	// Decode base64 to binary
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("解码 base64 图像失败 [用户:%s]: %v", licenseID, err)
		return "", fmt.Errorf("解码失败: %v", err)
	}

	// Upload using the same method
	return UploadImageData(imageData, licenseID, directory)
}

// UploadVideoData uploads video binary data to OSS and returns public URL.
func UploadVideoData(videoData []byte, userID string, extension string) (string, error) {
	if bucket == nil && !useLocalStorage {
		return "", fmt.Errorf("OSS client is not initialized")
	}
	if len(videoData) == 0 {
		return "", fmt.Errorf("video data is empty")
	}

	ext := strings.ToLower(strings.TrimSpace(extension))
	if ext == "" {
		ext = ".mp4"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	if bucket == nil {
		return uploadLocal(videoData, "videos/"+userID, ext)
	}

	timestamp := time.Now().UnixNano() / 1e6
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	objectKey := fmt.Sprintf("userupload/videos/%s/%d_%s%s", userID, timestamp, hex.EncodeToString(randBytes), ext)

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "video/mp4"
	}

	err := bucket.PutObject(objectKey, bytes.NewReader(videoData), oss.ContentType(contentType))
	if err != nil {
		log.Printf("upload video to OSS failed [user:%s]: %v", userID, err)
		return "", fmt.Errorf("upload failed: %v", err)
	}

	publicURL := buildPublicURL(objectKey)
	log.Printf("video upload success: %s -> %s (size: %d bytes)", objectKey, publicURL, len(videoData))
	return publicURL, nil
}

// DownloadAndUploadVideo 下载视频并上传到OSS
// videoURL: 原始视频URL
// userID: 用户ID
// headers: 可选的请求头（如 Google API 需要 x-goog-api-key）
// returns: OSS上的永久URL
func DownloadAndUploadVideo(videoURL string, userID string, headers ...map[string]string) (string, error) {
	if bucket == nil && !useLocalStorage {
		return "", fmt.Errorf("OSS 客户端未初始化且本地存储未启用")
	}
	log.Printf("downloading video: %s", videoURL)
	// 下载视频（Google 等需要代理的 URL 自动走代理）
	client := &http.Client{Timeout: 300 * time.Second}
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
	}
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求失败: %v", err)
	}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("下载视频失败 [用户:%s]: %v", userID, err)
		return "", fmt.Errorf("下载视频失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载视频失败: HTTP %d", resp.StatusCode)
	}

	// 读取视频数据
	videoData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取视频数据失败 [用户:%s]: %v", userID, err)
		return "", fmt.Errorf("读取视频数据失败: %v", err)
	}

	if bucket == nil {
		return uploadLocal(videoData, "videos/"+userID, ".mp4")
	}

	// 生成唯一文件名（含随机后缀防碰撞）
	timestamp := time.Now().UnixNano() / 1e6
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	objectKey := fmt.Sprintf("videos/%s/%d_%s.mp4", userID, timestamp, hex.EncodeToString(randBytes))

	// 上传到OSS
	err = bucket.PutObject(objectKey, bytes.NewReader(videoData))
	if err != nil {
		log.Printf("上传视频到 OSS 失败 [用户:%s]: %v", userID, err)
		return "", fmt.Errorf("上传视频失败: %v", err)
	}

	// 获取公开URL
	publicURL := buildPublicURL(objectKey)
	log.Printf("视频上传成功: %s -> %s (大小: %d bytes)", objectKey, publicURL, len(videoData))
	return publicURL, nil
}
