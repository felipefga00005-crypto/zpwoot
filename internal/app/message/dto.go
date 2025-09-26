package message

import (
	"time"

	"zpwoot/internal/domain/message"
)

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	To          string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Type        string `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Body        string `json:"body,omitempty" example:"Hello World!"`
	Caption     string `json:"caption,omitempty" example:"Image caption"`
	File        string `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename    string `json:"filename,omitempty" example:"document.pdf"`
	MimeType    string `json:"mimeType,omitempty" example:"image/jpeg"`
	
	// Location specific fields
	Latitude  float64 `json:"latitude,omitempty" example:"-23.5505"`
	Longitude float64 `json:"longitude,omitempty" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP"`
	
	// Contact specific fields
	ContactName  string `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string `json:"contactPhone,omitempty" example:"+5511999999999"`
} // @name SendMessageRequest

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name SendMessageResponse

// FromDomainRequest converts domain request to DTO request
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

// ToDomainRequest converts DTO request to domain request
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

// FromDomainResponse converts domain response to DTO response
func FromDomainResponse(resp *message.SendMessageResponse) *SendMessageResponse {
	return &SendMessageResponse{
		ID:        resp.MessageID,
		Status:    resp.Status,
		Timestamp: resp.Timestamp,
	}
}

// ToDomainResponse converts DTO response to domain response
func (r *SendMessageResponse) ToDomainResponse() *message.SendMessageResponse {
	return &message.SendMessageResponse{
		MessageID: r.ID,
		Status:    r.Status,
		Timestamp: r.Timestamp,
	}
}

// ButtonMessageRequest represents a button message request
type ButtonMessageRequest struct {
	To      string   `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body    string   `json:"body" validate:"required" example:"Choose an option:"`
	Buttons []Button `json:"buttons" validate:"required,min=1,max=3"`
} // @name ButtonMessageRequest

// Button represents a button in a button message
type Button struct {
	ID   string `json:"id" example:"btn_1"`
	Text string `json:"text" validate:"required" example:"Option 1"`
} // @name Button

// ListMessageRequest represents a list message request
type ListMessageRequest struct {
	To         string    `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string    `json:"body" validate:"required" example:"Please select an option:"`
	ButtonText string    `json:"buttonText" validate:"required" example:"View Options"`
	Sections   []Section `json:"sections" validate:"required,min=1"`
} // @name ListMessageRequest

// Section represents a section in a list message
type Section struct {
	Title string `json:"title" example:"Main Options"`
	Rows  []Row  `json:"rows" validate:"required,min=1,max=10"`
} // @name Section

// Row represents a row in a list section
type Row struct {
	ID          string `json:"id" example:"row_1"`
	Title       string `json:"title" validate:"required" example:"Option 1"`
	Description string `json:"description" example:"Description for option 1"`
} // @name Row

// TextMessageRequest represents a text message request
type TextMessageRequest struct {
	To   string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body string `json:"body" validate:"required" example:"Hello World!"`
} // @name TextMessageRequest

// MediaMessageRequest represents a media message request
type MediaMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/image.jpg"`
	Caption  string `json:"caption" example:"Image caption"`
	MimeType string `json:"mimeType" example:"image/jpeg"`
	Filename string `json:"filename" example:"image.jpg"`
} // @name MediaMessageRequest

// LocationMessageRequest represents a location message request
type LocationMessageRequest struct {
	To        string  `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Address   string  `json:"address" example:"São Paulo, SP"`
} // @name LocationMessageRequest

// ContactMessageRequest represents a contact message request
type ContactMessageRequest struct {
	To           string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ContactName  string `json:"contactName" validate:"required" example:"John Doe"`
	ContactPhone string `json:"contactPhone" validate:"required" example:"+5511999999999"`
} // @name ContactMessageRequest

// ReactionMessageRequest represents a reaction message request
type ReactionMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	Reaction  string `json:"reaction" validate:"required" example:"👍"`
} // @name ReactionMessageRequest

// PresenceMessageRequest represents a presence message request
type PresenceMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Presence string `json:"presence" validate:"required,oneof=typing recording online offline paused" example:"typing"`
} // @name PresenceMessageRequest

// EditMessageRequest represents an edit message request
type EditMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	NewBody   string `json:"newBody" validate:"required" example:"Updated message text"`
} // @name EditMessageRequest

// DeleteMessageRequest represents a delete message request
type DeleteMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	ForAll    bool   `json:"forAll" example:"true"`
} // @name DeleteMessageRequest

// MessageResponse represents a standard message response
type MessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name MessageResponse

// ReactionResponse represents a reaction response
type ReactionResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Reaction  string    `json:"reaction" example:"👍"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name ReactionResponse

// PresenceResponse represents a presence response
type PresenceResponse struct {
	Status    string    `json:"status" example:"sent"`
	Presence  string    `json:"presence" example:"typing"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name PresenceResponse

// EditResponse represents an edit message response
type EditResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"edited"`
	NewBody   string    `json:"newBody" example:"Updated message text"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name EditResponse

// DeleteResponse represents a delete message response
type DeleteResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"deleted"`
	ForAll    bool      `json:"forAll" example:"true"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name DeleteResponse
