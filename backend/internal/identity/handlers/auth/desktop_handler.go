package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"his-system/internal/identity/application/command"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/middleware"
	"his-system/pkg/response"
)

func handleErr(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*appErrors.AppError); ok {
		return response.Fail(c, appErr)
	}
	return response.Fail(c, &appErrors.AppError{
		Code:    "INTERNAL_ERROR",
		Status:  500,
		Message: err.Error(),
	})
}

type DesktopAuthHandler struct {
	initLoginHandler     *command.InitLoginHandler
	completeLoginHandler *command.CompleteLoginHandler
	refreshTokenHandler  *command.RefreshTokenHandler
	logoutHandler        *command.LogoutHandler
	setupMFAHandler      *command.SetupMFAHandler
	verifyMFAHandler     *command.VerifyMFAHandler
	updateLangHandler    *command.UpdateLanguageHandler
}

func NewDesktopAuthHandler(
	init *command.InitLoginHandler,
	complete *command.CompleteLoginHandler,
	refresh *command.RefreshTokenHandler,
	logout *command.LogoutHandler,
	setupMFA *command.SetupMFAHandler,
	verifyMFA *command.VerifyMFAHandler,
	updateLang *command.UpdateLanguageHandler,
) *DesktopAuthHandler {
	return &DesktopAuthHandler{
		initLoginHandler:     init,
		completeLoginHandler: complete,
		refreshTokenHandler:  refresh,
		logoutHandler:        logout,
		setupMFAHandler:      setupMFA,
		verifyMFAHandler:     verifyMFA,
		updateLangHandler:    updateLang,
	}
}

type InitLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// InitLogin godoc
// @Summary Initialize Desktop Login
// @Description Authenticates user and returns a challenge string
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Param request body InitLoginRequest true "Login Credentials"
// @Success 200 {object} command.InitLoginResult
// @Failure 401 {object} response.Response
// @Failure 429 {object} response.Response
// @Router /auth/login/init [post]
func (h *DesktopAuthHandler) InitLogin(c *fiber.Ctx) error {
	var req InitLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request payload"})
	}

	cmd := command.InitLoginCommand{
		Username: req.Username,
		Password: req.Password,
		ClientIP: c.IP(),
	}

	res, err := h.initLoginHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, res)
}

type CompleteLoginRequest struct {
	ChallengeString   string `json:"challenge_string"`
	Signature         string `json:"signature"`
	PublicKeyPEM      string `json:"public_key_pem"`
	MFAToken          string `json:"mfa_token"`
	DeviceFingerprint string `json:"device_fingerprint"`
}

// CompleteLogin godoc
// @Summary Complete Desktop Login
// @Description Verifies challenge signature and issues tokens
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Param request body CompleteLoginRequest true "Challenge Response"
// @Success 200 {object} command.CompleteLoginResult
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /auth/login/complete [post]
func (h *DesktopAuthHandler) CompleteLogin(c *fiber.Ctx) error {
	var req CompleteLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400})
	}

	cmd := command.CompleteLoginCommand{
		ChallengeString:   req.ChallengeString,
		Signature:         req.Signature,
		PublicKeyPEM:      req.PublicKeyPEM,
		MFAToken:          req.MFAToken,
		DeviceFingerprint: req.DeviceFingerprint,
		ClientIP:          c.IP(),
	}

	res, err := h.completeLoginHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, res)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	Signature    string `json:"signature"`
	PublicKeyPEM string `json:"public_key_pem"`
}

// RefreshToken godoc
// @Summary Refresh Access Token
// @Description Rotates tokens using valid Refresh Token and device signature
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh Payload"
// @Success 200 {object} command.RefreshTokenResult
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *DesktopAuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400})
	}

	cmd := command.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
		Signature:    req.Signature,
		PublicKeyPEM: req.PublicKeyPEM,
	}

	res, err := h.refreshTokenHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, res)
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout godoc
// @Summary Logout Desktop
// @Description Invalidates the refresh token
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh Token"
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *DesktopAuthHandler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400})
	}

	cmd := command.LogoutCommand{RefreshToken: req.RefreshToken}
	_ = h.logoutHandler.Handle(c.Context(), cmd)

	return response.OK(c, "Logged out successfully")
}

type SetupMFARequest struct {
	// Typically passed via claims/context
}

// SetupMFA godoc
// @Summary Setup TOTP MFA
// @Description Generates TOTP secret and backup codes. Requires valid access token.
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} command.SetupMFAResult
// @Failure 401 {object} response.Response
// @Router /auth/mfa/setup [post]
func (h *DesktopAuthHandler) SetupMFA(c *fiber.Ctx) error {
	// Assume a generic JWT middleware has populated c.Locals("claims") or user context
	// For now, let's implement a quick mock extraction or rely on Fiber context
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return response.Fail(c, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401})
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Fail(c, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Missing claims in context"})
	}

	cmd := command.SetupMFACommand{
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	res, err := h.setupMFAHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, res)
}

type VerifyMFARequest struct {
	UserID   string `json:"user_id"` // Alternatively username
	TOTPCode string `json:"totp_code"`
}

// VerifyMFA godoc
// @Summary Verify TOTP MFA
// @Description Validates TOTP code to complete login flow
// @Tags Auth (Desktop)
// @Accept json
// @Produce json
// @Param request body VerifyMFARequest true "MFA Code"
// @Success 200 {object} command.VerifyMFAResult
// @Failure 401 {object} response.Response
// @Router /auth/mfa/verify [post]
func (h *DesktopAuthHandler) VerifyMFA(c *fiber.Ctx) error {
	var req VerifyMFARequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400})
	}

	cmd := command.VerifyMFACommand{
		Username: req.UserID,
		Code:     req.TOTPCode,
	}

	res, err := h.verifyMFAHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, res)
}

type UpdateLanguageReq struct {
	Language string `json:"language"`
}

func (h *DesktopAuthHandler) UpdateLanguage(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Fail(c, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401})
	}

	var req UpdateLanguageReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	err := h.updateLangHandler.Handle(c.Context(), command.UpdateLanguageCommand{
		UserID:   claims.UserID,
		Language: req.Language,
	})
	if err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, fiber.Map{"success": true})
}