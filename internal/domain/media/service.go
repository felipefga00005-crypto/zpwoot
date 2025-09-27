package media

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"

	"zpwoot/platform/logger"
)

// Service defines the interface for media domain service
type Service interface {
	DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error)
	GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfo, error)
	ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error)
	ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error)
	GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error)
	ReadCachedFile(ctx context.Context, filePath string) ([]byte, error)
}

// WhatsAppClient defines the interface for WhatsApp operations
type WhatsAppClient interface {
	IsLoggedIn() bool
	DownloadMedia(ctx context.Context, messageID string) ([]byte, string, error) // returns data, mimeType, error
	GetMessageInfo(ctx context.Context, messageID string) (*MessageInfo, error)
}

// MessageInfo represents information about a WhatsApp message
type MessageInfo struct {
	ID        string
	FromJID   string
	Timestamp time.Time
	MediaType string
	MimeType  string
	FileSize  int64
	Filename  string
	Caption   string
	HasMedia  bool
}

// CacheManager defines the interface for cache operations
type CacheManager interface {
	SaveFile(ctx context.Context, data []byte, filename string) (string, error)
	ReadFile(ctx context.Context, filePath string) ([]byte, error)
	DeleteFile(ctx context.Context, filePath string) error
	ListFiles(ctx context.Context, pattern string) ([]string, error)
	GetFileInfo(ctx context.Context, filePath string) (os.FileInfo, error)
	CleanupOldFiles(ctx context.Context, olderThan time.Duration) (int, int64, error)
}

type serviceImpl struct {
	whatsappClient WhatsAppClient
	cacheManager   CacheManager
	logger         *logger.Logger
	cacheDir       string
	maxFileSize    int64
}

// NewService creates a new media domain service
func NewService(whatsappClient WhatsAppClient, cacheManager CacheManager, logger *logger.Logger, cacheDir string) Service {
	return &serviceImpl{
		whatsappClient: whatsappClient,
		cacheManager:   cacheManager,
		logger:         logger,
		cacheDir:       cacheDir,
		maxFileSize:    100 * 1024 * 1024, // 100MB default
	}
}

// DownloadMedia downloads media from a WhatsApp message
func (s *serviceImpl) DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error) {
	if err := ValidateDownloadRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Downloading media from WhatsApp", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
		"media_type": req.MediaType,
	})

	// Check if WhatsApp client is logged in
	if !s.whatsappClient.IsLoggedIn() {
		return nil, ErrClientNotLoggedIn
	}

	// Get message info first to validate media type if specified
	msgInfo, err := s.whatsappClient.GetMessageInfo(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get message info", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrMessageNotFound
	}

	if !msgInfo.HasMedia {
		return nil, ErrNoMediaInMessage
	}

	// Validate media type if specified
	if req.MediaType != "" && msgInfo.MediaType != req.MediaType {
		return nil, ErrMediaTypeMismatch
	}

	// Download media from WhatsApp
	data, mimeType, err := s.whatsappClient.DownloadMedia(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to download media", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrDownloadFailed
	}

	if int64(len(data)) > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	// Generate filename
	filename := s.generateFilename(req.MessageID, mimeType, msgInfo.Filename)

	// Save to cache
	filePath, err := s.cacheManager.SaveFile(ctx, data, filename)
	if err != nil {
		s.logger.WarnWithFields("Failed to cache media file", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		// Continue without caching
		filePath = ""
	}

	s.logger.InfoWithFields("Media downloaded successfully", map[string]interface{}{
		"message_id": req.MessageID,
		"file_size":  len(data),
		"mime_type":  mimeType,
		"cached":     filePath != "",
	})

	return &DownloadMediaResponse{
		Data:      data,
		MimeType:  mimeType,
		FileSize:  int64(len(data)),
		Filename:  filename,
		MediaType: msgInfo.MediaType,
		FilePath:  filePath,
	}, nil
}

// GetMediaInfo gets information about media in a message without downloading it
func (s *serviceImpl) GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfo, error) {
	if err := ValidateMediaInfoRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Getting media info", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
	})

	// Check if WhatsApp client is logged in
	if !s.whatsappClient.IsLoggedIn() {
		return nil, ErrClientNotLoggedIn
	}

	// Get message info
	msgInfo, err := s.whatsappClient.GetMessageInfo(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get message info", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrMessageNotFound
	}

	if !msgInfo.HasMedia {
		return nil, ErrNoMediaInMessage
	}

	return &MediaInfo{
		MessageID: msgInfo.ID,
		MediaType: msgInfo.MediaType,
		MimeType:  msgInfo.MimeType,
		FileSize:  msgInfo.FileSize,
		Filename:  msgInfo.Filename,
		Caption:   msgInfo.Caption,
		Timestamp: msgInfo.Timestamp,
		FromJID:   msgInfo.FromJID,
	}, nil
}

// ListCachedMedia lists cached media files
func (s *serviceImpl) ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error) {
	if err := ValidateListCachedMediaRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Listing cached media", map[string]interface{}{
		"session_id": req.SessionID,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"media_type": req.MediaType,
	})

	// This is a simplified implementation
	// In a real implementation, you would query the cache database/storage
	pattern := "*"
	if req.MediaType != "" {
		pattern = fmt.Sprintf("*_%s_*", req.MediaType)
	}

	files, err := s.cacheManager.ListFiles(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list cached files: %w", err)
	}

	// Apply pagination
	total := len(files)
	start := req.Offset
	end := start + req.Limit

	if start >= total {
		return &ListCachedMediaResponse{
			Items:     []CachedMediaItem{},
			Total:     total,
			Limit:     req.Limit,
			Offset:    req.Offset,
			HasMore:   false,
			TotalSize: 0,
		}, nil
	}

	if end > total {
		end = total
	}

	items := make([]CachedMediaItem, 0, end-start)
	var totalSize int64

	for i := start; i < end; i++ {
		filePath := files[i]
		info, err := s.cacheManager.GetFileInfo(ctx, filePath)
		if err != nil {
			continue
		}

		// Parse filename to extract metadata (simplified)
		filename := filepath.Base(filePath)

		item := CachedMediaItem{
			MessageID:  extractMessageIDFromFilename(filename),
			MediaType:  extractMediaTypeFromFilename(filename),
			MimeType:   extractMimeTypeFromFilename(filename),
			FileSize:   info.Size(),
			Filename:   filename,
			CachedAt:   info.ModTime(),
			LastAccess: info.ModTime(), // Simplified
			ExpiresAt:  info.ModTime().Add(24 * time.Hour),
			FilePath:   filePath,
		}

		items = append(items, item)
		totalSize += info.Size()
	}

	return &ListCachedMediaResponse{
		Items:     items,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
		HasMore:   end < total,
		TotalSize: totalSize,
	}, nil
}

// ClearCache clears cached media files
func (s *serviceImpl) ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error) {
	if err := ValidateClearCacheRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Clearing media cache", map[string]interface{}{
		"session_id": req.SessionID,
		"older_than": req.OlderThan,
		"media_type": req.MediaType,
	})

	olderThan := time.Duration(req.OlderThan) * time.Hour
	filesDeleted, spaceFreed, err := s.cacheManager.CleanupOldFiles(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup cache: %w", err)
	}

	s.logger.InfoWithFields("Cache cleared successfully", map[string]interface{}{
		"files_deleted": filesDeleted,
		"space_freed":   spaceFreed,
	})

	return &ClearCacheResponse{
		FilesDeleted: filesDeleted,
		SpaceFreed:   spaceFreed,
	}, nil
}

// GetMediaStats gets statistics about media usage
func (s *serviceImpl) GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error) {
	s.logger.InfoWithFields("Getting media stats", map[string]interface{}{
		"session_id": req.SessionID,
	})

	// This is a simplified implementation
	// In a real implementation, you would query actual statistics
	stats := MediaStats{
		TotalFiles:    0,
		TotalSize:     0,
		ImageFiles:    0,
		VideoFiles:    0,
		AudioFiles:    0,
		DocumentFiles: 0,
		CacheHitRate:  0.85,
		AvgFileSize:   524288, // 512KB
	}

	return &GetMediaStatsResponse{
		SessionID: req.SessionID,
		Stats:     stats,
		UpdatedAt: time.Now(),
	}, nil
}

// ReadCachedFile reads a cached file
func (s *serviceImpl) ReadCachedFile(ctx context.Context, filePath string) ([]byte, error) {
	return s.cacheManager.ReadFile(ctx, filePath)
}

// Helper functions

func (s *serviceImpl) generateFilename(messageID, mimeType, originalFilename string) string {
	if originalFilename != "" {
		return originalFilename
	}

	// Generate filename based on message ID and MIME type
	ext := ""
	if mimeType != "" {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[0]
		}
	}

	return fmt.Sprintf("%s%s", messageID, ext)
}

// Simplified helper functions for filename parsing
func extractMessageIDFromFilename(filename string) string {
	// This is a simplified implementation
	// In a real implementation, you would parse the filename properly
	return filename
}

func extractMediaTypeFromFilename(filename string) string {
	// This is a simplified implementation
	return "unknown"
}

func extractMimeTypeFromFilename(filename string) string {
	// This is a simplified implementation
	ext := filepath.Ext(filename)
	return mime.TypeByExtension(ext)
}
