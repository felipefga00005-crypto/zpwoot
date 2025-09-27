package session

import (
	"context"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
)

// UseCase defines the session use case interface
type UseCase interface {
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error)
	ListSessions(ctx context.Context, req *ListSessionsRequest) (*ListSessionsResponse, error)
	GetSessionInfo(ctx context.Context, sessionID string) (*SessionInfoResponse, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ConnectSession(ctx context.Context, sessionID string) error
	LogoutSession(ctx context.Context, sessionID string) error
	GetQRCode(ctx context.Context, sessionID string) (*QRCodeResponse, error)
	PairPhone(ctx context.Context, sessionID string, req *PairPhoneRequest) error
	SetProxy(ctx context.Context, sessionID string, req *SetProxyRequest) error
	GetProxy(ctx context.Context, sessionID string) (*ProxyResponse, error)
}

// useCaseImpl implements the session use case
type useCaseImpl struct {
	sessionRepo    ports.SessionRepository
	WameowMgr      ports.WameowManager
	sessionService *session.Service
}

// NewUseCase creates a new session use case
func NewUseCase(
	sessionRepo ports.SessionRepository,
	WameowMgr ports.WameowManager,
	sessionService *session.Service,
) UseCase {
	return &useCaseImpl{
		sessionRepo:    sessionRepo,
		WameowMgr:      WameowMgr,
		sessionService: sessionService,
	}
}

// CreateSession creates a new Wameow session
func (uc *useCaseImpl) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// Convert DTO to domain request
	domainReq := req.ToCreateSessionRequest()

	// Create session using domain service
	sess, err := uc.sessionService.CreateSession(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	var proxyConfig *ProxyConfig
	if sess.ProxyConfig != nil {
		proxyConfig = &ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}

	response := &CreateSessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		ProxyConfig: proxyConfig,
		CreatedAt:   sess.CreatedAt,
	}

	return response, nil
}

// ListSessions retrieves a list of sessions with optional filtering
func (uc *useCaseImpl) ListSessions(ctx context.Context, req *ListSessionsRequest) (*ListSessionsResponse, error) {
	// Convert DTO to domain request
	domainReq := &session.ListSessionsRequest{
		IsConnected: req.IsConnected,
		DeviceJid:   req.DeviceJid,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	// Set defaults
	if domainReq.Limit == 0 {
		domainReq.Limit = 20
	}

	// Get sessions from domain service
	sessions, total, err := uc.sessionService.ListSessions(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain entities to response DTOs
	sessionResponses := make([]SessionInfoResponse, len(sessions))
	for i, sess := range sessions {
		// Create SessionInfo from Session
		sessionInfo := &session.SessionInfo{
			Session: sess,
		}
		sessionResponses[i] = *FromSessionInfo(sessionInfo)
	}

	response := &ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    total,
		Limit:    domainReq.Limit,
		Offset:   domainReq.Offset,
	}

	return response, nil
}

// GetSessionInfo retrieves detailed information about a specific session
func (uc *useCaseImpl) GetSessionInfo(ctx context.Context, sessionID string) (*SessionInfoResponse, error) {
	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Create SessionInfo from Session
	sessionInfo := &session.SessionInfo{
		Session: sess,
	}

	// Convert domain entity to response DTO
	response := FromSessionInfo(sessionInfo)
	return response, nil
}

// DeleteSession removes a session permanently
func (uc *useCaseImpl) DeleteSession(ctx context.Context, sessionID string) error {
	return uc.sessionService.DeleteSession(ctx, sessionID)
}

// ConnectSession establishes connection with Wameow
func (uc *useCaseImpl) ConnectSession(ctx context.Context, sessionID string) error {
	return uc.sessionService.ConnectSession(ctx, sessionID)
}

// LogoutSession logs out from Wameow
func (uc *useCaseImpl) LogoutSession(ctx context.Context, sessionID string) error {
	return uc.sessionService.LogoutSession(ctx, sessionID)
}

// GetQRCode retrieves the current QR code for pairing
func (uc *useCaseImpl) GetQRCode(ctx context.Context, sessionID string) (*QRCodeResponse, error) {
	// Get QR code from domain service
	qrCode, err := uc.sessionService.GetQRCode(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Convert domain entity to response DTO
	response := FromQRCodeResponse(qrCode)
	return response, nil
}

// PairPhone pairs a phone number with the session
func (uc *useCaseImpl) PairPhone(ctx context.Context, sessionID string, req *PairPhoneRequest) error {
	// This would typically use the Wameow manager to pair the phone
	// For now, just return nil (not implemented)
	return nil
}

// SetProxy configures proxy for the session
func (uc *useCaseImpl) SetProxy(ctx context.Context, sessionID string, req *SetProxyRequest) error {
	domainProxyConfig := &session.ProxyConfig{
		Type:     req.ProxyConfig.Type,
		Host:     req.ProxyConfig.Host,
		Port:     req.ProxyConfig.Port,
		Username: req.ProxyConfig.Username,
		Password: req.ProxyConfig.Password,
	}
	return uc.sessionService.SetProxy(ctx, sessionID, domainProxyConfig)
}

// GetProxy retrieves proxy configuration for the session
func (uc *useCaseImpl) GetProxy(ctx context.Context, sessionID string) (*ProxyResponse, error) {
	// Get proxy config from domain service
	proxyConfig, err := uc.sessionService.GetProxy(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var appProxyConfig *ProxyConfig
	if proxyConfig != nil {
		appProxyConfig = &ProxyConfig{
			Type:     proxyConfig.Type,
			Host:     proxyConfig.Host,
			Port:     proxyConfig.Port,
			Username: proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	response := &ProxyResponse{
		ProxyConfig: appProxyConfig,
	}

	return response, nil
}
