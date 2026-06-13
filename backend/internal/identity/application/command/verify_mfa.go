package command

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	commonRedis "github.com/thanhbvha/go-common/redis"
	"golang.org/x/crypto/bcrypt"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
	appErrors "his-system/pkg/errors"
)

type VerifyMFACommand struct {
	Username string
	Code     string
}

type VerifyMFAResult struct {
	MFAToken string `json:"mfa_token"`
}

type VerifyMFAHandler struct {
	mfaRepo  domain.MFARepository
	userRepo domain.UserRepository
	rdb      *commonRedis.Client
	encKey   []byte
}

func NewVerifyMFAHandler(mfaRepo domain.MFARepository, userRepo domain.UserRepository, rdb *commonRedis.Client, encKey []byte) *VerifyMFAHandler {
	return &VerifyMFAHandler{mfaRepo: mfaRepo, userRepo: userRepo, rdb: rdb, encKey: encKey}
}

func (h *VerifyMFAHandler) Handle(ctx context.Context, cmd VerifyMFACommand) (*VerifyMFAResult, error) {
	user, err := h.userRepo.GetByUsername(ctx, cmd.Username)
	if err != nil || user == nil {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Tài khoản không tồn tại"}
	}

	encSecret, backupCodes, err := h.mfaRepo.GetSecret(ctx, user.ID)
	if err != nil || encSecret == "" {
		return nil, &appErrors.AppError{Code: "FORBIDDEN", Status: 403, Message: "Người dùng chưa thiết lập MFA"}
	}

	// Try backup codes first if length > 6
	if len(cmd.Code) > 6 {
		validBackup := false
		var remainingCodes []string
		for i, hc := range backupCodes {
			if !validBackup && bcrypt.CompareHashAndPassword([]byte(hc), []byte(cmd.Code)) == nil {
				validBackup = true
				remainingCodes = append(backupCodes[:i], backupCodes[i+1:]...)
			}
		}

		if validBackup {
			// Update backup codes
			h.mfaRepo.SaveSecret(ctx, user.ID, encSecret, remainingCodes)
			return h.issueMFAToken(ctx, user.ID)
		}
	}

	// Decrypt TOTP Secret
	secretBytes, err := crypto.DecryptAESGCM([]byte(encSecret), h.encKey, []byte(user.ID.String()))
	if err != nil {
		return nil, &appErrors.AppError{Code: "INTERNAL_ERROR", Status: 500, Message: "Lỗi giải mã MFA secret"}
	}

	// Validate TOTP
	valid := totp.Validate(cmd.Code, string(secretBytes))
	if !valid {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Mã MFA không chính xác"}
	}

	// Enable MFA for user if not enabled yet
	if !user.MFAEnabled {
		user.MFAEnabled = true
		h.userRepo.Update(ctx, user)
	}

	return h.issueMFAToken(ctx, user.ID)
}

func (h *VerifyMFAHandler) issueMFAToken(ctx context.Context, userID uuid.UUID) (*VerifyMFAResult, error) {
	b := make([]byte, 32)
	rand.Read(b)
	mfaToken := base64.RawURLEncoding.EncodeToString(b)

	key := h.rdb.BuildKey(fmt.Sprintf("mfa:%s", mfaToken))
	if err := h.rdb.Set(ctx, key, userID.String(), 5*time.Minute); err != nil {
		return nil, err
	}

	return &VerifyMFAResult{MFAToken: mfaToken}, nil
}
