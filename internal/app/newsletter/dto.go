package newsletter

import (
	"time"
	"zpwoot/internal/domain/newsletter"
)

// CreateNewsletterRequest - Request para criar newsletter
type CreateNewsletterRequest struct {
	Name        string `json:"name" validate:"required,max=25"`
	Description string `json:"description,omitempty" validate:"max=512"`
}

// CreateNewsletterResponse - Response da criação de newsletter
type CreateNewsletterResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	InviteCode  string    `json:"inviteCode"`
	State       string    `json:"state"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
}

// GetNewsletterInfoRequest - Request para obter info de newsletter
type GetNewsletterInfoRequest struct {
	JID string `json:"jid" validate:"required"`
}

// GetNewsletterInfoWithInviteRequest - Request para obter info via convite
type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey" validate:"required"`
}

// NewsletterInfoResponse - Response com informações do newsletter
type NewsletterInfoResponse struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	InviteCode      string                      `json:"inviteCode"`
	SubscriberCount int                         `json:"subscriberCount"`
	State           string                      `json:"state"`
	Role            string                      `json:"role"`
	Muted           bool                        `json:"muted"`
	MuteState       string                      `json:"muteState"`
	Verified        bool                        `json:"verified"`
	VerificationState string                    `json:"verificationState"`
	CreationTime    time.Time                   `json:"creationTime"`
	UpdateTime      time.Time                   `json:"updateTime"`
	Picture         *ProfilePictureInfo         `json:"picture,omitempty"`
	Preview         *ProfilePictureInfo         `json:"preview,omitempty"`
}

// ProfilePictureInfo - Informações da foto do perfil
type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

// FollowNewsletterRequest - Request para seguir newsletter
type FollowNewsletterRequest struct {
	JID string `json:"jid" validate:"required"`
}

// UnfollowNewsletterRequest - Request para deixar de seguir newsletter
type UnfollowNewsletterRequest struct {
	JID string `json:"jid" validate:"required"`
}

// SubscribedNewslettersResponse - Response com newsletters seguidos
type SubscribedNewslettersResponse struct {
	Newsletters []NewsletterInfoResponse `json:"newsletters"`
	Total       int                      `json:"total"`
}

// NewsletterActionResponse - Response genérica para ações
type NewsletterActionResponse struct {
	JID       string    `json:"jid"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversion methods

// ToDomain converts CreateNewsletterRequest to domain entity
func (req *CreateNewsletterRequest) ToDomain() *newsletter.CreateNewsletterRequest {
	return &newsletter.CreateNewsletterRequest{
		Name:        req.Name,
		Description: req.Description,
	}
}

// FromDomain converts domain NewsletterInfo to NewsletterInfoResponse
func (resp *NewsletterInfoResponse) FromDomain(info *newsletter.NewsletterInfo) {
	resp.ID = info.ID
	resp.Name = info.Name
	resp.Description = info.Description
	resp.InviteCode = info.InviteCode
	resp.SubscriberCount = info.SubscriberCount
	resp.State = string(info.State)
	resp.Role = string(info.Role)
	resp.Muted = info.Muted
	resp.MuteState = string(info.MuteState)
	resp.Verified = info.Verified
	resp.VerificationState = string(info.VerificationState)
	resp.CreationTime = info.CreationTime
	resp.UpdateTime = info.UpdateTime
	
	if info.Picture != nil {
		resp.Picture = &ProfilePictureInfo{
			URL:    info.Picture.URL,
			ID:     info.Picture.ID,
			Type:   info.Picture.Type,
			Direct: info.Picture.Direct,
		}
	}
	
	if info.Preview != nil {
		resp.Preview = &ProfilePictureInfo{
			URL:    info.Preview.URL,
			ID:     info.Preview.ID,
			Type:   info.Preview.Type,
			Direct: info.Preview.Direct,
		}
	}
}

// FromDomainList converts a list of domain NewsletterInfo to NewsletterInfoResponse list
func FromDomainList(infos []*newsletter.NewsletterInfo) []NewsletterInfoResponse {
	responses := make([]NewsletterInfoResponse, len(infos))
	for i, info := range infos {
		responses[i].FromDomain(info)
	}
	return responses
}

// ToDomain converts GetNewsletterInfoRequest to domain entity
func (req *GetNewsletterInfoRequest) ToDomain() *newsletter.GetNewsletterInfoRequest {
	return &newsletter.GetNewsletterInfoRequest{
		JID: req.JID,
	}
}

// ToDomain converts GetNewsletterInfoWithInviteRequest to domain entity
func (req *GetNewsletterInfoWithInviteRequest) ToDomain() *newsletter.GetNewsletterInfoWithInviteRequest {
	return &newsletter.GetNewsletterInfoWithInviteRequest{
		InviteKey: req.InviteKey,
	}
}

// ToDomain converts FollowNewsletterRequest to domain entity
func (req *FollowNewsletterRequest) ToDomain() *newsletter.FollowNewsletterRequest {
	return &newsletter.FollowNewsletterRequest{
		JID: req.JID,
	}
}

// ToDomain converts UnfollowNewsletterRequest to domain entity
func (req *UnfollowNewsletterRequest) ToDomain() *newsletter.UnfollowNewsletterRequest {
	return &newsletter.UnfollowNewsletterRequest{
		JID: req.JID,
	}
}

// Validation methods

// Validate validates the CreateNewsletterRequest
func (req *CreateNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the GetNewsletterInfoRequest
func (req *GetNewsletterInfoRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the GetNewsletterInfoWithInviteRequest
func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the FollowNewsletterRequest
func (req *FollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the UnfollowNewsletterRequest
func (req *UnfollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Helper functions

// NewCreateNewsletterResponse creates a new CreateNewsletterResponse from domain data
func NewCreateNewsletterResponse(info *newsletter.NewsletterInfo) *CreateNewsletterResponse {
	return &CreateNewsletterResponse{
		ID:          info.ID,
		Name:        info.Name,
		Description: info.Description,
		InviteCode:  info.InviteCode,
		State:       string(info.State),
		Role:        string(info.Role),
		CreatedAt:   info.CreationTime,
	}
}

// NewNewsletterInfoResponse creates a new NewsletterInfoResponse from domain data
func NewNewsletterInfoResponse(info *newsletter.NewsletterInfo) *NewsletterInfoResponse {
	resp := &NewsletterInfoResponse{}
	resp.FromDomain(info)
	return resp
}

// NewSubscribedNewslettersResponse creates a new SubscribedNewslettersResponse
func NewSubscribedNewslettersResponse(infos []*newsletter.NewsletterInfo) *SubscribedNewslettersResponse {
	newsletters := FromDomainList(infos)
	return &SubscribedNewslettersResponse{
		Newsletters: newsletters,
		Total:       len(newsletters),
	}
}

// NewNewsletterActionResponse creates a new NewsletterActionResponse
func NewNewsletterActionResponse(jid, status, message string) *NewsletterActionResponse {
	return &NewsletterActionResponse{
		JID:       jid,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// Success response helpers

// NewSuccessFollowResponse creates a success response for follow action
func NewSuccessFollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter followed successfully")
}

// NewSuccessUnfollowResponse creates a success response for unfollow action
func NewSuccessUnfollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter unfollowed successfully")
}

// Error response helpers

// NewErrorResponse creates an error response
func NewErrorResponse(jid, message string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "error", message)
}
