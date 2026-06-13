package command

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/thanhbvha/go-common/logger"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	"his-system/pkg/crypto"
	appErrors "his-system/pkg/errors"
)

type VerifyOTPCommand struct {
	Phone string
	OTP   string
}

type VerifyOTPResult struct {
	NeedsRegister bool   `json:"needs_register"`
	AccessToken   string `json:"access_token,omitempty"`
}

type VerifyOTPHandler struct {
	patientRepo domain.PatientRepository
	rdb         *commonRedis.Client
	cipher      *crypto.FieldCipher
	signKey     []byte
	encKey      []byte
}

func NewVerifyOTPHandler(patientRepo domain.PatientRepository, rdb *commonRedis.Client, cipher *crypto.FieldCipher, signKey, encKey []byte) *VerifyOTPHandler {
	return &VerifyOTPHandler{
		patientRepo: patientRepo,
		rdb:         rdb,
		cipher:      cipher,
		signKey:     signKey,
		encKey:      encKey,
	}
}

// Handle returns (result, refreshToken, error)
func (h *VerifyOTPHandler) Handle(ctx context.Context, cmd VerifyOTPCommand) (*VerifyOTPResult, string, error) {
	phoneHMAC := h.cipher.HMAC(cmd.Phone)
	otpKey := h.rdb.BuildKey(fmt.Sprintf("otp:%s", phoneHMAC))

	val, err := h.rdb.Get(ctx, otpKey)
	if err != nil || val == "" {
		logger.ErrorAsync("VerifyOTPHandler.Handle: otp expired or not found", slog.String("error", fmt.Sprintf("%v", err)), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Mã OTP đã hết hạn hoặc không tồn tại"}
	}

	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		h.rdb.Delete(ctx, otpKey)
		return nil, "", &appErrors.AppError{Code: "INTERNAL_ERROR", Status: 500, Message: "Dữ liệu OTP không hợp lệ"}
	}

	storedOTP := parts[0]
	attempts, _ := strconv.Atoi(parts[1])

	if attempts >= 5 {
		h.rdb.Delete(ctx, otpKey)
		return nil, "", &appErrors.AppError{Code: "TOO_MANY_REQUESTS", Status: 429, Message: "Bạn đã nhập sai quá 5 lần. Vui lòng yêu cầu mã OTP mới."}
	}

	// Constant time compare
	if subtle.ConstantTimeCompare([]byte(storedOTP), []byte(cmd.OTP)) != 1 {
		// Increment attempt
		newVal := fmt.Sprintf("%s:%d", storedOTP, attempts+1)
		ttl, _ := h.rdb.TTL(ctx, otpKey)
		if ttl > 0 {
			h.rdb.Set(ctx, otpKey, newVal, ttl)
		}
		logger.ErrorAsync("VerifyOTPHandler.Handle: invalid otp", slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Mã OTP không chính xác"}
	}

	// OTP is correct! Clear from Redis
	h.rdb.Delete(ctx, otpKey)

	// Check if patient exists
	patient, err := h.patientRepo.GetByPhoneHMAC(ctx, phoneHMAC)
	if err != nil {
		logger.ErrorAsync("VerifyOTPHandler.Handle: failed to get patient", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", err
	}

	if patient == nil {
		return &VerifyOTPResult{NeedsRegister: true}, "", nil
	}

	// Patient exists, issue tokens
	claims := auth.Claims{
		UserID: patient.ID,
		Roles:  []string{"patient"}, // Hardcode for now, or fetch from users table
	}
	accessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, "")
	if err != nil {
		logger.ErrorAsync("VerifyOTPHandler.Handle: failed to issue access token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", err
	}

	refreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		logger.ErrorAsync("VerifyOTPHandler.Handle: failed to issue refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", err
	}

	rtHash := auth.HashToken(refreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))
	// Save refresh token to Redis (No pubKeyHash for web)
	if err := h.rdb.Set(ctx, rtKey, patient.ID.String(), 7*24*time.Hour); err != nil {
		logger.ErrorAsync("VerifyOTPHandler.Handle: failed to save refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, "", err
	}

	return &VerifyOTPResult{
		NeedsRegister: false,
		AccessToken:   accessToken,
	}, refreshToken, nil
}
