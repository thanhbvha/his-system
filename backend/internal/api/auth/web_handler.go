package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"his-system/internal/identity/application/command"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

type WebAuthHandler struct {
	sendOTPHandler         *command.SendOTPHandler
	verifyOTPHandler       *command.VerifyOTPHandler
	registerPatientHandler *command.RegisterPatientHandler
	refreshWebHandler      *command.RefreshWebHandler
	logoutWebHandler       *command.LogoutWebHandler
}

func NewWebAuthHandler(
	sendOTPHandler *command.SendOTPHandler,
	verifyOTPHandler *command.VerifyOTPHandler,
	registerPatientHandler *command.RegisterPatientHandler,
	refreshWebHandler *command.RefreshWebHandler,
	logoutWebHandler *command.LogoutWebHandler,
) *WebAuthHandler {
	return &WebAuthHandler{
		sendOTPHandler:         sendOTPHandler,
		verifyOTPHandler:       verifyOTPHandler,
		registerPatientHandler: registerPatientHandler,
		refreshWebHandler:      refreshWebHandler,
		logoutWebHandler:       logoutWebHandler,
	}
}

// setRefreshCookie sets the HTTP-only cookie for Web
func setRefreshCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Domain:   "",               // Adjust as needed
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		Secure:   true,             // Requires HTTPS
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// clearRefreshCookie clears the HTTP-only cookie
func clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

type SendOTPRequest struct {
	Phone string `json:"phone"`
}

// SendOTP godoc
// @Summary Send OTP via SMS/Zalo
// @Description Generates and sends OTP to patient's phone
// @Tags Auth (Web)
// @Accept json
// @Produce json
// @Param request body SendOTPRequest true "Phone number"
// @Success 200 {object} response.Response
// @Failure 400,422,429 {object} response.Response
// @Router /auth/otp/send [post]
func (h *WebAuthHandler) SendOTP(c *fiber.Ctx) error {
	var req SendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Payload không hợp lệ"})
	}

	cmd := command.SendOTPCommand{
		Phone:    req.Phone,
		ClientIP: c.IP(),
	}

	if err := h.sendOTPHandler.Handle(c.Context(), cmd); err != nil {
		return handleErr(c, err)
	}

	return response.OK(c, fiber.Map{"success": true, "message": "OTP đã được gửi"})
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verifies patient's OTP and issues tokens if user exists
// @Tags Auth (Web)
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP data"
// @Success 200 {object} command.VerifyOTPResult
// @Failure 401,429 {object} response.Response
// @Router /auth/otp/verify [post]
func (h *WebAuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Payload không hợp lệ"})
	}

	cmd := command.VerifyOTPCommand{
		Phone: req.Phone,
		OTP:   req.OTP,
	}

	res, rt, err := h.verifyOTPHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	if !res.NeedsRegister && rt != "" {
		setRefreshCookie(c, rt)
	}

	return response.OK(c, res)
}

type RegisterPatientRequest struct {
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	DOB      string `json:"dob"`    // "2006-01-02"
	Gender   string `json:"gender"` // "male", "female", "other"
	Email    string `json:"email"`
}

// Register godoc
// @Summary Register Patient
// @Description Registers a new patient after OTP verification
// @Tags Auth (Web)
// @Accept json
// @Produce json
// @Param request body RegisterPatientRequest true "Patient info"
// @Success 200 {object} command.RegisterPatientResult
// @Failure 400,409 {object} response.Response
// @Router /auth/register [post]
func (h *WebAuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterPatientRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Payload không hợp lệ"})
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Ngày sinh không đúng định dạng YYYY-MM-DD"})
	}

	cmd := command.RegisterPatientCommand{
		Phone:    req.Phone,
		FullName: req.FullName,
		DOB:      dob,
		Gender:   req.Gender,
		Email:    req.Email,
	}

	res, rt, err := h.registerPatientHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	setRefreshCookie(c, rt)

	return response.OK(c, res)
}

// RefreshWeb godoc
// @Summary Refresh Token for Web
// @Description Rotates HttpOnly refresh cookie and issues new access token
// @Tags Auth (Web)
// @Produce json
// @Success 200 {object} command.RefreshWebResult
// @Failure 401 {object} response.Response
// @Router /auth/web/refresh [post]
func (h *WebAuthHandler) RefreshWeb(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	if cookie == "" {
		return response.Fail(c, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Yêu cầu refresh token"})
	}

	cmd := command.RefreshWebCommand{RefreshToken: cookie}
	res, newRt, err := h.refreshWebHandler.Handle(c.Context(), cmd)
	if err != nil {
		return handleErr(c, err)
	}

	setRefreshCookie(c, newRt)
	return response.OK(c, res)
}

// LogoutWeb godoc
// @Summary Logout Web
// @Description Clears refresh token cookie
// @Tags Auth (Web)
// @Produce json
// @Success 200 {object} response.Response
// @Router /auth/web/logout [post]
func (h *WebAuthHandler) LogoutWeb(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	cmd := command.LogoutWebCommand{RefreshToken: cookie}

	if err := h.logoutWebHandler.Handle(c.Context(), cmd); err != nil {
		return handleErr(c, err)
	}

	clearRefreshCookie(c)
	return response.OK(c, fiber.Map{"success": true, "message": "Đăng xuất thành công"})
}
