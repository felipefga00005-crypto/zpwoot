package message

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

type MediaProcessor struct {
	logger  *logger.Logger
	tempDir string
	maxSize int64 // Maximum file size in bytes
	timeout time.Duration
}

func NewMediaProcessor(logger *logger.Logger) *MediaProcessor {
	return &MediaProcessor{
		logger:  logger,
		tempDir: os.TempDir(),
		maxSize: 100 * 1024 * 1024, // 100MB default
		timeout: 30 * time.Second,
	}
}

type ProcessedMedia struct {
	FilePath string
	MimeType string
	FileSize int64
	Cleanup  func() error
}

func (mp *MediaProcessor) ProcessMedia(ctx context.Context, file string) (*ProcessedMedia, error) {
	if file == "" {
		return nil, fmt.Errorf("file content is empty")
	}

	if strings.HasPrefix(file, "data:") {
		return mp.processBase64(file)
	}

	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		return mp.processURL(ctx, file)
	}

	return nil, fmt.Errorf("unsupported file format: must be URL or base64")
}

func (mp *MediaProcessor) processBase64(data string) (*ProcessedMedia, error) {
	mp.logger.Debug("Processing base64 media")

	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid base64 data format")
	}

	mimeType := "application/octet-stream"
	if strings.Contains(parts[0], ":") && strings.Contains(parts[0], ";") {
		mimePart := strings.Split(parts[0], ":")[1]
		mimeType = strings.Split(mimePart, ";")[0]
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	if int64(len(decoded)) > mp.maxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	tempFile, err := os.CreateTemp(mp.tempDir, "whatsmeow-media-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if _, err := tempFile.Write(decoded); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	mp.logger.InfoWithFields("Base64 media processed", map[string]interface{}{
		"file_path": tempFile.Name(),
		"mime_type": mimeType,
		"file_size": len(decoded),
	})

	return &ProcessedMedia{
		FilePath: tempFile.Name(),
		MimeType: mimeType,
		FileSize: int64(len(decoded)),
		Cleanup: func() error {
			return os.Remove(tempFile.Name())
		},
	}, nil
}

func (mp *MediaProcessor) processURL(ctx context.Context, url string) (*ProcessedMedia, error) {
	mp.logger.InfoWithFields("Processing URL media", map[string]interface{}{
		"url": url,
	})

	client := &http.Client{
		Timeout: mp.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "zpwoot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	if resp.ContentLength > mp.maxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	tempFile, err := os.CreateTemp(mp.tempDir, "whatsmeow-media-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	written, err := io.CopyN(tempFile, resp.Body, mp.maxSize+1)
	if err != nil && err != io.EOF {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to copy data to temporary file: %w", err)
	}

	if written > mp.maxSize {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	mp.logger.InfoWithFields("URL media processed", map[string]interface{}{
		"url":       url,
		"file_path": tempFile.Name(),
		"mime_type": mimeType,
		"file_size": written,
	})

	return &ProcessedMedia{
		FilePath: tempFile.Name(),
		MimeType: mimeType,
		FileSize: written,
		Cleanup: func() error {
			return os.Remove(tempFile.Name())
		},
	}, nil
}

func DetectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".avi":  "video/avi",
		".mov":  "video/quicktime",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".txt":  "text/plain",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

func ValidateMessageRequest(req *SendMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("recipient (to) is required")
	}

	if req.Type == "" {
		return fmt.Errorf("message type is required")
	}

	switch req.Type {
	case MessageTypeText:
		if req.Body == "" {
			return fmt.Errorf("body is required for text messages")
		}
	case MessageTypeImage, MessageTypeAudio, MessageTypeVideo, MessageTypeDocument, MessageTypeSticker:
		if req.File == "" {
			return fmt.Errorf("file is required for %s messages", req.Type)
		}
	case MessageTypeLocation:
		if req.Latitude == 0 || req.Longitude == 0 {
			return fmt.Errorf("latitude and longitude are required for location messages")
		}
	case MessageTypeContact:
		if req.ContactName == "" || req.ContactPhone == "" {
			return fmt.Errorf("contact name and phone are required for contact messages")
		}
	default:
		return fmt.Errorf("unsupported message type: %s", req.Type)
	}

	return nil
}
