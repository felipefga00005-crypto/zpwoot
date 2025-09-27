package message

import (
	"strings"
	"time"
)

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

type MediaSource string

const (
	MediaSourceURL    MediaSource = "url"
	MediaSourceBase64 MediaSource = "base64"
	MediaSourceFile   MediaSource = "file"
)

type SendResult struct {
	MessageID string    `json:"messageId"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type SendMessageRequest struct {
	To       string      `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Type     MessageType `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Body     string      `json:"body,omitempty" example:"Hello World!"`
	Caption  string      `json:"caption,omitempty" example:"Image caption"`
	File     string      `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename string      `json:"filename,omitempty" example:"document.pdf"`
	MimeType string      `json:"mimeType,omitempty" example:"image/jpeg"`

	Latitude  float64 `json:"latitude,omitempty" example:"-23.5505"`
	Longitude float64 `json:"longitude,omitempty" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP"`

	ContactName  string       `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string       `json:"contactPhone,omitempty" example:"+5511999999999"`
	ContextInfo  *ContextInfo `json:"contextInfo,omitempty"`
}

type ContextInfo struct {
	StanzaID    string `json:"stanzaId" validate:"required" example:"ABCD1234abcd"`
	Participant string `json:"participant,omitempty" example:"5511999999999@s.whatsapp.net"`
}

type SendMessageResponse struct {
	MessageID string    `json:"messageId" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

type MediaInfo struct {
	MimeType string `json:"mimeType"`
	FileSize int64  `json:"fileSize"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"` // for audio/video in seconds
}

type LocationMessage struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	Name      string  `json:"name,omitempty"`
}

type ContactMessage struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

func (req *SendMessageRequest) IsMediaMessage() bool {
	return req.Type != MessageTypeText && req.Type != MessageTypeLocation && req.Type != MessageTypeContact
}

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
