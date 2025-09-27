package session

import (
	"time"

	domainSession "zpwoot/internal/domain/session"
)

// ProxyConfig represents proxy configuration for sessions
type ProxyConfig struct {
	Type     string `json:"type" example:"http"` // http, socks5
	Host     string `json:"host" example:"proxy.example.com"`
	Port     int    `json:"port" example:"8080"`
	Username string `json:"username,omitempty" example:"proxyuser"`
	Password string `json:"password,omitempty" example:"proxypass123"`
} //@name ProxyConfig

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	Name        string       `json:"name" validate:"required,min=3,max=50" example:"my-whatsapp-session"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name CreateSessionRequest

// CreateSessionResponse represents the response after creating a session
type CreateSessionResponse struct {
	ID          string       `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	Name        string       `json:"name" example:"my-whatsapp-session"`
	IsConnected bool         `json:"isConnected" example:"false"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	CreatedAt   time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name CreateSessionResponse

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Name        *string      `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Session Name"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name UpdateSessionRequest

// ListSessionsRequest represents the request to list sessions
type ListSessionsRequest struct {
	IsConnected *bool   `json:"isConnected,omitempty" query:"isConnected" example:"true"`
	DeviceJid   *string `json:"deviceJid,omitempty" query:"deviceJid" example:"5511999999999@s.Wameow.net"`
	Limit       int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset      int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0" example:"0"`
} //@name ListSessionsRequest

// ListSessionsResponse represents the response for listing sessions
type ListSessionsResponse struct {
	Sessions []SessionInfoResponse `json:"sessions"`
	Total    int                   `json:"total" example:"10"`
	Limit    int                   `json:"limit" example:"20"`
	Offset   int                   `json:"offset" example:"0"`
} //@name ListSessionsResponse

// SessionInfoResponse represents session information in responses
type SessionInfoResponse struct {
	Session    *SessionResponse    `json:"session"`
	DeviceInfo *DeviceInfoResponse `json:"deviceInfo,omitempty"`
} //@name SessionInfoResponse

// SessionResponse represents a session in responses
type SessionResponse struct {
	ID              string       `json:"id" example:"session-123"`
	Name            string       `json:"name" example:"my-Wameow-session"`
	DeviceJid       string       `json:"deviceJid,omitempty" example:"5511999999999@s.Wameow.net"`
	IsConnected     bool         `json:"isConnected" example:"false"`
	ConnectionError *string      `json:"connectionError,omitempty" example:"Connection timeout"`
	ProxyConfig     *ProxyConfig `json:"proxyConfig,omitempty"`
	CreatedAt       time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time    `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
	ConnectedAt     *time.Time   `json:"connectedAt,omitempty" example:"2024-01-01T00:00:30Z"`
} //@name SessionResponse

// DeviceInfoResponse represents device information in responses
type DeviceInfoResponse struct {
	Platform    string `json:"platform" example:"android"`
	DeviceModel string `json:"deviceModel" example:"Samsung Galaxy S21"`
	OSVersion   string `json:"osVersion" example:"11"`
	AppVersion  string `json:"appVersion" example:"2.21.4.18"`
} //@name DeviceInfoResponse

// PairPhoneRequest represents the request to pair a phone
type PairPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,e164" example:"+5511987654321"`
} //@name PairPhoneRequest

// QRCodeResponse represents the QR code response
type QRCodeResponse struct {
	QRCode    string    `json:"qrCode" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="`
	ExpiresAt time.Time `json:"expiresAt" example:"2024-01-01T00:01:00Z"`
	Timeout   int       `json:"timeoutSeconds" example:"60"`
} //@name QRCodeResponse

// SetProxyRequest represents the request to set proxy configuration
type SetProxyRequest struct {
	ProxyConfig ProxyConfig `json:"proxyConfig"`
} //@name SetProxyRequest

// ProxyResponse represents the proxy configuration response
type ProxyResponse struct {
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name ProxyResponse

// Conversion methods

// ToCreateSessionRequest converts to domain request
func (r *CreateSessionRequest) ToCreateSessionRequest() *domainSession.CreateSessionRequest {
	var proxyConfig *domainSession.ProxyConfig
	if r.ProxyConfig != nil {
		proxyConfig = &domainSession.ProxyConfig{
			Type:     r.ProxyConfig.Type,
			Host:     r.ProxyConfig.Host,
			Port:     r.ProxyConfig.Port,
			Username: r.ProxyConfig.Username,
			Password: r.ProxyConfig.Password,
		}
	}
	return &domainSession.CreateSessionRequest{
		Name:        r.Name,
		ProxyConfig: proxyConfig,
	}
}

// FromSession converts from domain session to response
func FromSession(s *domainSession.Session) *SessionResponse {
	var proxyConfig *ProxyConfig
	if s.ProxyConfig != nil {
		proxyConfig = &ProxyConfig{
			Type:     s.ProxyConfig.Type,
			Host:     s.ProxyConfig.Host,
			Port:     s.ProxyConfig.Port,
			Username: s.ProxyConfig.Username,
			Password: s.ProxyConfig.Password,
		}
	}

	response := &SessionResponse{
		ID:              s.ID.String(),
		Name:            s.Name,
		IsConnected:     s.IsConnected,
		ConnectionError: s.ConnectionError,
		ProxyConfig:     proxyConfig,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		ConnectedAt:     s.ConnectedAt,
	}

	// Only include deviceJid if it's not empty (obtained after connection)
	if s.DeviceJid != "" {
		response.DeviceJid = s.DeviceJid
	}

	return response
}

// FromSessionInfo converts from domain session info to response
func FromSessionInfo(si *domainSession.SessionInfo) *SessionInfoResponse {
	response := &SessionInfoResponse{}

	if si.Session != nil {
		response.Session = FromSession(si.Session)
	}

	if si.DeviceInfo != nil {
		response.DeviceInfo = &DeviceInfoResponse{
			Platform:    si.DeviceInfo.Platform,
			DeviceModel: si.DeviceInfo.DeviceModel,
			OSVersion:   si.DeviceInfo.OSVersion,
			AppVersion:  si.DeviceInfo.AppVersion,
		}
	}

	return response
}

// FromQRCodeResponse converts from domain QR code response
func FromQRCodeResponse(qr *domainSession.QRCodeResponse) *QRCodeResponse {
	return &QRCodeResponse{
		QRCode:    qr.QRCode,
		ExpiresAt: qr.ExpiresAt,
		Timeout:   qr.Timeout,
	}
}
