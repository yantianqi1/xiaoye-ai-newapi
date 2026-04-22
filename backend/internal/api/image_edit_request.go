package api

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
)

type imageBinary struct {
	data        []byte
	contentType string
}

func downloadImageBinary(imageURL string) (*imageBinary, error) {
	data, contentType, err := downloadImageBytes(imageURL)
	if err != nil {
		return nil, err
	}

	return &imageBinary{data: data, contentType: contentType}, nil
}

func createImageEditMultipartBody(model, prompt, size string, imageURLs []string, maskURL string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writeImageEditFields(writer, model, prompt, size); err != nil {
		return nil, "", err
	}
	if err := writeImageEditImages(writer, imageURLs); err != nil {
		return nil, "", err
	}
	if err := writeImageEditMask(writer, maskURL); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("关闭 multipart 失败: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

func writeImageEditFields(writer *multipart.Writer, model, prompt, size string) error {
	fields := map[string]string{
		"model":           model,
		"prompt":          prompt,
		"response_format": "b64_json",
		"size":            size,
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("写入字段 %s 失败: %w", key, err)
		}
	}
	return nil
}

func writeImageEditImages(writer *multipart.Writer, imageURLs []string) error {
	for index, imageURL := range imageURLs {
		fieldName := "image"
		if len(imageURLs) > 1 {
			fieldName = "image[]"
		}
		binary, err := downloadImageBinary(imageURL)
		if err != nil {
			return fmt.Errorf("下载输入图片失败: %w", err)
		}
		if err := writeImageBinaryPart(writer, fieldName, fmt.Sprintf("image-%d", index+1), binary); err != nil {
			return err
		}
	}
	return nil
}

func writeImageEditMask(writer *multipart.Writer, maskURL string) error {
	if strings.TrimSpace(maskURL) == "" {
		return nil
	}

	binary, err := downloadImageBinary(maskURL)
	if err != nil {
		return fmt.Errorf("下载mask图片失败: %w", err)
	}
	return writeImageBinaryPart(writer, "mask", "mask", binary)
}

func writeImageBinaryPart(writer *multipart.Writer, fieldName, filePrefix string, binary *imageBinary) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s%s"`, fieldName, filePrefix, imageMimeExtension(binary.contentType)))
	header.Set("Content-Type", binary.contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("创建 multipart 分片失败: %w", err)
	}
	if _, err := part.Write(binary.data); err != nil {
		return fmt.Errorf("写入图片内容失败: %w", err)
	}
	return nil
}

func imageMimeExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return filepath.Ext("." + strings.TrimPrefix(contentType, "image/"))
	}
}
