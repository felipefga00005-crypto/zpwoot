package message

import (
	"time"

	"zpwoot/internal/domain/message"
)

type SendMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Type     string `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Body     string `json:"body,omitempty" example:"Hello World!"`
	Caption  string `json:"caption,omitempty" example:"Image caption"`
	File     string `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename string `json:"filename,omitempty" example:"document.pdf"` // Only used for document type, not for audio
	MimeType string `json:"mimeType,omitempty" example:"image/jpeg"`

	Latitude  float64 `json:"latitude,omitempty" example:"-23.5505"`
	Longitude float64 `json:"longitude,omitempty" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP"`

	ContactName  string       `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string       `json:"contactPhone,omitempty" example:"+5511999999999"`
	ContextInfo  *ContextInfo `json:"contextInfo,omitempty"`
} //@name SendMessageRequest

type SendMessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name SendMessageResponse

func FromDomainRequest(req *message.SendMessageRequest) *SendMessageRequest {
	return &SendMessageRequest{
		To:           req.To,
		Type:         string(req.Type),
		Body:         req.Body,
		Caption:      req.Caption,
		File:         req.File,
		Filename:     req.Filename,
		MimeType:     req.MimeType,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Address:      req.Address,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
	}
}

func (r *SendMessageRequest) ToDomainRequest() *message.SendMessageRequest {
	return &message.SendMessageRequest{
		To:           r.To,
		Type:         message.MessageType(r.Type),
		Body:         r.Body,
		Caption:      r.Caption,
		File:         r.File,
		Filename:     r.Filename,
		MimeType:     r.MimeType,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		Address:      r.Address,
		ContactName:  r.ContactName,
		ContactPhone: r.ContactPhone,
	}
}

func FromDomainResponse(resp *message.SendMessageResponse) *SendMessageResponse {
	return &SendMessageResponse{
		ID:        resp.MessageID,
		Status:    resp.Status,
		Timestamp: resp.Timestamp,
	}
}

func (r *SendMessageResponse) ToDomainResponse() *message.SendMessageResponse {
	return &message.SendMessageResponse{
		MessageID: r.ID,
		Status:    r.Status,
		Timestamp: r.Timestamp,
	}
}

type ButtonMessageRequest struct {
	To      string   `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body    string   `json:"body" validate:"required" example:"Please choose one of the options below:"`
	Buttons []Button `json:"buttons" validate:"required,min=1,max=3"`
} //@name ButtonMessageRequest

type Button struct {
	ID   string `json:"id" example:"btn_yes"`
	Text string `json:"text" validate:"required" example:"Yes, I agree"`
} //@name Button

type ListMessageRequest struct {
	To         string    `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string    `json:"body" validate:"required" example:"Please select one of the available options:"`
	ButtonText string    `json:"buttonText" validate:"required" example:"Select Option"`
	Sections   []Section `json:"sections" validate:"required,min=1"`
} //@name ListMessageRequest

type Section struct {
	Title string `json:"title" example:"Available Services"`
	Rows  []Row  `json:"rows" validate:"required,min=1,max=10"`
} //@name Section

type Row struct {
	ID          string `json:"id" example:"service_support"`
	Title       string `json:"title" validate:"required" example:"Customer Support"`
	Description string `json:"description" example:"Get help from our support team"`
} //@name Row

type MediaMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/media.file"`
	Caption  string `json:"caption" example:"Media caption"`
	MimeType string `json:"mimeType" example:"application/octet-stream"`
	Filename string `json:"filename" example:"media.file"`
} //@name MediaMessageRequest

type ImageMessageRequest struct {
	To          string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/image.jpg"`
	Caption     string       `json:"caption" example:"Beautiful sunset photo"`
	MimeType    string       `json:"mimeType" example:"image/jpeg"`
	Filename    string       `json:"filename" example:"sunset.jpg"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name ImageMessageRequest

type VideoMessageRequest struct {
	To          string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/video.mp4"`
	Caption     string       `json:"caption" example:"Check out this amazing video!"`
	MimeType    string       `json:"mimeType" example:"video/mp4"`
	Filename    string       `json:"filename" example:"amazing_video.mp4"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name VideoMessageRequest

type AudioMessageRequest struct {
	To          string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/audio.ogg"`
	Caption     string       `json:"caption" example:"Voice message"`
	MimeType    string       `json:"mimeType" example:"audio/ogg"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
	// Note: AudioMessage in WhatsApp protocol doesn't support filename field
	// Use DocumentMessage for files that need a filename
} //@name AudioMessageRequest

type DocumentMessageRequest struct {
	To          string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/document.pdf"`
	Caption     string       `json:"caption" example:"Important document"`
	MimeType    string       `json:"mimeType" example:"application/pdf"`
	Filename    string       `json:"filename" validate:"required" example:"important_document.pdf"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name DocumentMessageRequest

type LocationMessageRequest struct {
	To        string  `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Address   string  `json:"address" example:"Avenida Paulista, 1578 - Bela Vista, São Paulo - SP, Brazil"`
} //@name LocationMessageRequest

type ContactMessageRequest struct {
	To           string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ContactName  string `json:"contactName" validate:"required" example:"Maria Silva"`
	ContactPhone string `json:"contactPhone" validate:"required" example:"+5511987654321"`
} //@name ContactMessageRequest

type ContactInfo struct {
	Name         string `json:"name" validate:"required" example:"João Santos"`
	Phone        string `json:"phone" validate:"required" example:"+5511987654321"`
	Email        string `json:"email,omitempty" example:"joao.santos@email.com"`                 // Not displayed in WhatsApp
	Organization string `json:"organization,omitempty" example:"Tech Solutions Ltda"`            // Displayed in WhatsApp
	Title        string `json:"title,omitempty" example:"Software Engineer"`                     // Not displayed in WhatsApp
	Website      string `json:"website,omitempty" example:"https://joaosantos.dev"`              // Not displayed in WhatsApp
	Address      string `json:"address,omitempty" example:"Rua das Flores, 123 - São Paulo, SP"` // Not displayed in WhatsApp
} //@name ContactInfo

type ContactListMessageRequest struct {
	To       string        `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Contacts []ContactInfo `json:"contacts" validate:"required,min=1,max=10"`
} //@name ContactListMessageRequest

type ContactListMessageResponse struct {
	TotalContacts int                 `json:"totalContacts" example:"3"`
	SuccessCount  int                 `json:"successCount" example:"3"`
	FailureCount  int                 `json:"failureCount" example:"0"`
	Results       []ContactSendResult `json:"results"`
	Timestamp     string              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
} //@name ContactListMessageResponse

type ContactSendResult struct {
	ContactName string `json:"contactName" example:"João Santos"`
	MessageID   string `json:"messageId,omitempty" example:"3EB07F264CA1B4AD714A3F"`
	Status      string `json:"status" example:"sent"`
	Error       string `json:"error,omitempty"`
} //@name ContactSendResult

type ReactionMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	Reaction  string `json:"reaction" validate:"required" example:"👍"`
} //@name ReactionMessageRequest

type PresenceMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Presence string `json:"presence" validate:"required,oneof=typing recording online offline paused" example:"typing"`
} //@name PresenceMessageRequest

type EditMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	NewBody   string `json:"newBody" validate:"required" example:"Updated message text"`
} //@name EditMessageRequest

type DeleteMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	ForAll    bool   `json:"forAll" example:"true"`
} //@name DeleteMessageRequest

type MessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name MessageResponse

type ReactionResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Reaction  string    `json:"reaction" example:"👍"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name ReactionResponse

type PresenceResponse struct {
	Status    string    `json:"status" example:"sent"`
	Presence  string    `json:"presence" example:"typing"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name PresenceResponse

type EditResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"edited"`
	NewBody   string    `json:"newBody" example:"Updated message text"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name EditResponse

type DeleteResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"deleted"`
	ForAll    bool      `json:"forAll" example:"true"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name DeleteResponse

type BusinessProfileRequest struct {
	To           string `json:"to" validate:"required" example:"5511987654321@s.whatsapp.net"`
	Name         string `json:"name" validate:"required" example:"Empresa Teste Ltda"`
	Phone        string `json:"phone" validate:"required" example:"+5511987654321"`
	Email        string `json:"email,omitempty" example:"contato@empresateste.com.br"`
	Organization string `json:"organization,omitempty" example:"Empresa Teste Ltda"`
	Title        string `json:"title,omitempty" example:"Atendimento ao Cliente"`
	Website      string `json:"website,omitempty" example:"https://www.empresateste.com.br"`
	Address      string `json:"address,omitempty" example:"Rua Teste, 123 - São Paulo, SP"`
} //@name BusinessProfileRequest

type TextMessageRequest struct {
	To          string       `json:"to" validate:"required" example:"5511987654321@s.whatsapp.net"`
	Body        string       `json:"body" validate:"required" example:"Hello, this is a text message"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name TextMessageRequest

type ContextInfo struct {
	StanzaID    string `json:"stanzaId" validate:"required" example:"ABCD1234abcd"`
	Participant string `json:"participant,omitempty" example:"5511999999999@s.whatsapp.net"`
} //@name ContextInfo
