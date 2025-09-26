package message

import (
	"strings"
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
	MessageTypeSticker  MessageType = "sticker"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
)

// MediaSource represents how media content is provided
type MediaSource string

const (
	MediaSourceURL    MediaSource = "url"
	MediaSourceBase64 MediaSource = "base64"
	MediaSourceFile   MediaSource = "file"
)

// SendResult represents the result of sending a message
type SendResult struct {
	MessageID string    `json:"messageId"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	To          string      `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Type        MessageType `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Body        string      `json:"body,omitempty" example:"Hello World!"`
	Caption     string      `json:"caption,omitempty" example:"Image caption"`
	File        string      `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename    string      `json:"filename,omitempty" example:"document.pdf"`
	MimeType    string      `json:"mimeType,omitempty" example:"image/jpeg"`
	
	// Location specific fields
	Latitude  float64 `json:"latitude,omitempty" example:"-23.5505"`
	Longitude float64 `json:"longitude,omitempty" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP"`
	
	// Contact specific fields
	ContactName  string `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string `json:"contactPhone,omitempty" example:"+5511999999999"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	MessageID string    `json:"messageId" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// MediaInfo represents information about media content
type MediaInfo struct {
	MimeType string `json:"mimeType"`
	FileSize int64  `json:"fileSize"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"` // for audio/video in seconds
}

// LocationMessage represents a location message
type LocationMessage struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	Name      string  `json:"name,omitempty"`
}

// ContactMessage represents a contact message
type ContactMessage struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// IsMediaMessage returns true if the message contains media
func (req *SendMessageRequest) IsMediaMessage() bool {
	return req.Type != MessageTypeText && req.Type != MessageTypeLocation && req.Type != MessageTypeContact
}

// GetMediaSource determines the media source type
func (req *SendMessageRequest) GetMediaSource() MediaSource {
	if req.File == "" {
		return ""
	}
	
	if strings.HasPrefix(req.File, "data:") {
		return MediaSourceBase64
	}
	
	if strings.HasPrefix(req.File, "http://") || strings.HasPrefix(req.File, "https://") {
		return MediaSourceURL
	}
	
	return MediaSourceFile
}
