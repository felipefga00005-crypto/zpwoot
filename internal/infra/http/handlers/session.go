package handlers

import (
	"fmt"
	"strings"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/session"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type SessionHandler struct {
	logger          *logger.Logger
	sessionUC       session.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewSessionHandler(appLogger *logger.Logger, sessionUC session.UseCase, sessionRepo helpers.SessionRepository) *SessionHandler {
	return &SessionHandler{
		logger:          appLogger,
		sessionUC:       sessionUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func NewSessionHandlerWithoutUseCase(appLogger *logger.Logger, sessionRepo helpers.SessionRepository) *SessionHandler {
	return &SessionHandler{
		logger:          appLogger,
		sessionUC:       nil, // Will be nil until properly wired
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func (h *SessionHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == "session not found" || err == domainSession.ErrSessionNotFound {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Failed to resolve session")
	}

	return sess, nil
}

// @Summary Create a new Wameow session
// @Description Creates a new Wameow session with the provided configuration. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body session.CreateSessionRequest true "Session creation request"
// @Success 201 {object} session.CreateSessionResponse "Session created successfully"
// @Failure 400 {object} object "Invalid request body or parameters"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/create [post]
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	h.logger.Info("Creating new session")

	if h.sessionUC == nil {
		h.logger.Error("Session use case not initialized")
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Session service not available",
		})
	}

	var req session.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if isValid, errorMsg := h.sessionResolver.ValidateSessionName(req.Name); !isValid {
		h.logger.WarnWithFields("Invalid session name provided", map[string]interface{}{
			"name":  req.Name,
			"error": errorMsg,
		})

		suggested := h.sessionResolver.SuggestValidName(req.Name)
		return c.Status(400).JSON(fiber.Map{
			"error":         "Invalid session name",
			"message":       errorMsg,
			"suggestedName": suggested,
			"namingRules": []string{
				"Must be 3-50 characters long",
				"Must start with a letter",
				"Can contain letters, numbers, hyphens, and underscores",
				"Cannot use reserved names (create, list, info, etc.)",
			},
		})
	}

	result, err := h.sessionUC.CreateSession(c.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create session: " + err.Error())

		if strings.Contains(err.Error(), "Session already exists") {
			return c.Status(409).JSON(fiber.Map{
				"success": false,
				"error":   "Session already exists",
				"message": fmt.Sprintf("A session with the name '%s' already exists. Please choose a different name.", req.Name),
			})
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to create session"))
	}

	response := common.NewSuccessResponse(result, "Session created successfully")
	return c.Status(201).JSON(response)
}

// @Summary List all Wameow sessions
// @Description Retrieves a list of all Wameow sessions with optional filtering. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Filter by session status" Enums(created,connecting,connected,disconnected,error,logged_out) example("connected")
// @Param deviceJid query string false "Filter by device JID" example("5511999999999@s.Wameow.net")
// @Param limit query int false "Limit number of results" minimum(1) maximum(100) default(20) example(20)
// @Param offset query int false "Offset for pagination" minimum(0) default(0) example(0)
// @Success 200 {object} session.ListSessionsResponse "Sessions retrieved successfully"
// @Failure 400 {object} object "Invalid request parameters"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/list [get]
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	h.logger.Info("Listing sessions")

	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	var req session.ListSessionsRequest

	if isConnectedStr := c.Query("isConnected"); isConnectedStr != "" {
		switch isConnectedStr {
		case "true":
			isConnected := true
			req.IsConnected = &isConnected
		case "false":
			isConnected := false
			req.IsConnected = &isConnected
		}
	}

	if deviceJid := c.Query("deviceJid"); deviceJid != "" {
		req.DeviceJid = &deviceJid
	}

	if limit := c.QueryInt("limit", 20); limit > 0 && limit <= 100 {
		req.Limit = limit
	} else {
		req.Limit = 20
	}

	if offset := c.QueryInt("offset", 0); offset >= 0 {
		req.Offset = offset
	}

	result, err := h.sessionUC.ListSessions(c.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list sessions: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to list sessions"))
	}

	response := common.NewSuccessResponse(result, "Sessions retrieved successfully")
	return c.JSON(response)
}

// @Summary Get session information
// @Description Retrieves detailed information about a specific Wameow session including connection status and device info. You can use either the session UUID or session name. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} session.SessionInfoResponse "Session info retrieved successfully"
// @Failure 400 {object} object "Invalid session identifier"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/info [get]
func (h *SessionHandler) GetSessionInfo(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Getting session info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.sessionUC.GetSessionInfo(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get session info: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get session info"))
	}

	response := common.NewSuccessResponse(result, "Session info retrieved successfully")
	return c.JSON(response)
}

// @Summary Delete a Wameow session
// @Description Permanently removes a Wameow session and all associated data. This action cannot be undone. You can use either the session UUID or session name. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} object "Session deleted successfully"
// @Failure 400 {object} object "Invalid session identifier"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/delete [delete]
func (h *SessionHandler) DeleteSession(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Deleting session", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	err := h.sessionUC.DeleteSession(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to delete session: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to delete session"))
	}

	response := common.NewSuccessResponse(nil, "Session deleted successfully")
	return c.JSON(response)
}

// @Summary Connect Wameow session
// @Description Establishes connection with Wameow for the specified session. Will generate QR code if not paired. You can use either the session UUID or session name. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} object "Connection initiated successfully"
// @Failure 400 {object} object "Invalid session identifier"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/connect [post]
func (h *SessionHandler) ConnectSession(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Connecting session", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	err := h.sessionUC.ConnectSession(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to connect session: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to connect session"))
	}

	response := common.NewSuccessResponse(nil, "Session connection initiated successfully")
	return c.JSON(response)
}

// @Summary Logout Wameow session
// @Description Logs out from Wameow for the specified session. You can use either the session UUID or session name. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} object "Session logged out successfully"
// @Failure 400 {object} object "Invalid session identifier"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/logout [post]
func (h *SessionHandler) LogoutSession(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Logging out session", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	err := h.sessionUC.LogoutSession(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to logout session: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to logout session"))
	}

	response := common.NewSuccessResponse(nil, "Session logged out successfully")
	return c.JSON(response)
}

// @Summary Get QR code for session pairing
// @Description Retrieves the current QR code for pairing a Wameow session. The QR code expires after 60 seconds. You can use either the session UUID or session name. Requires API key authentication.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth

// @Param id path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=session.QRCodeResponse} "QR code retrieved successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid session identifier"
// @Failure 401 {object} common.ErrorResponse "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} common.ErrorResponse "Session not found or no QR code available"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /sessions/{sessionId}/qr [get]
func (h *SessionHandler) GetQRCode(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Getting QR code", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.sessionUC.GetQRCode(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get QR code: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get QR code"))
	}

	response := common.NewSuccessResponse(result, "QR code retrieved successfully")
	return c.JSON(response)
}

func (h *SessionHandler) PairPhone(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	var req session.PairPhoneRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse pair phone request: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	ctx := c.Context()
	err := h.sessionUC.PairPhone(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to pair phone: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to pair phone"))
	}

	response := common.NewSuccessResponse(nil, "Phone pairing initiated successfully")
	return c.JSON(response)
}

func (h *SessionHandler) SetProxy(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Setting proxy", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	var req session.SetProxyRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	err := h.sessionUC.SetProxy(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to set proxy: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to set proxy"))
	}

	response := common.NewSuccessResponse(nil, "Proxy configuration updated successfully")
	return c.JSON(response)
}

func (h *SessionHandler) GetProxy(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Getting proxy config", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.sessionUC.GetProxy(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get proxy: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get proxy"))
	}

	response := common.NewSuccessResponse(result, "Proxy configuration retrieved successfully")
	return c.JSON(response)
}
