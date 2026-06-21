package command

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"

	"github.com/google/uuid"
	"github.com/thanhbvha/go-common/logger"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type CompleteLoginCommand struct {
	ChallengeString   string
	Signature         string // base64 DER-encoded ECDSA signature
	PublicKeyPEM      string // ECDSA P-256 Public Key PEM
	MFAToken          string // optional
	DeviceFingerprint string
	ClientIP          string
}

type CompleteLoginResult struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserPayload `json:"user"`
}

type UserPayload struct {
	ID                string   `json:"id"`
	Username          string   `json:"username"`
	RoleIDs           []string `json:"role_ids"`
	Roles             []string `json:"roles"`
	MFAEnabled        bool     `json:"mfa_enabled"`
	PreferredLanguage string   `json:"preferred_language"`
}

type CompleteLoginHandler struct {
	userRepo   domain.UserRepository
	deviceRepo domain.DeviceRepository
	roleRepo   domain.RoleRepository
	rdb        *commonRedis.Client
	signKey    []byte
	encKey     []byte
}

func NewCompleteLoginHandler(userRepo domain.UserRepository, deviceRepo domain.DeviceRepository, roleRepo domain.RoleRepository, rdb *commonRedis.Client, signKey, encKey []byte) *CompleteLoginHandler {
	return &CompleteLoginHandler{userRepo: userRepo, deviceRepo: deviceRepo, roleRepo: roleRepo, rdb: rdb, signKey: signKey, encKey: encKey}
}

func (h *CompleteLoginHandler) Handle(ctx context.Context, cmd CompleteLoginCommand) (*CompleteLoginResult, error) {
	challengeKey := h.rdb.BuildKey(fmt.Sprintf("challenge:%s", cmd.ChallengeString))
	userIDStr, err := h.rdb.Get(ctx, challengeKey)
	if err != nil || userIDStr == "" {
		logger.ErrorAsync("CompleteLoginHandler.Handle: challenge expired or invalid", slog.String("error", fmt.Sprintf("%v", err)), slog.String("challenge", cmd.ChallengeString), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Challenge hết hạn hoặc không hợp lệ"}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: user not found", slog.String("error", fmt.Sprintf("%v", err)), slog.String("user_id", userID.String()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Không tìm thấy người dùng"}
	}

	// Verify ECDSA Signature
	if err := h.verifySignature(cmd.ChallengeString, cmd.Signature, cmd.PublicKeyPEM); err != nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: invalid device signature", slog.String("error", err.Error()), slog.String("user_id", userID.String()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Chữ ký thiết bị không hợp lệ"}
	}

	// Verify MFA if required
	if user.MFAEnabled {
		if cmd.MFAToken == "" {
			return nil, &appErrors.AppError{Code: "FORBIDDEN", Status: 403, Message: "Yêu cầu mã xác thực MFA"}
		}
		mfaKey := h.rdb.BuildKey(fmt.Sprintf("mfa:%s", cmd.MFAToken))
		mfaUserID, err := h.rdb.Get(ctx, mfaKey)
		if err != nil || mfaUserID != userIDStr {
			return nil, &appErrors.AppError{Code: "FORBIDDEN", Status: 403, Message: "Mã MFA hết hạn hoặc không hợp lệ"}
		}
		// MFA verified, delete token
		h.rdb.Native().Del(ctx, mfaKey)
	}

	// Hash public key for hardware binding
	hSha256 := sha256.New()
	hSha256.Write([]byte(cmd.PublicKeyPEM))
	pubKeyHash := hex.EncodeToString(hSha256.Sum(nil))

	// Upsert Device
	device := &domain.Device{
		ID:                uuid.New(),
		UserID:            userID,
		DeviceFingerprint: cmd.DeviceFingerprint,
		PublicKeyPEM:      cmd.PublicKeyPEM,
		PublicKeyHash:     pubKeyHash,
		RegisteredAt:      time.Now(),
		IsActive:          true,
	}
	if err := h.deviceRepo.Upsert(ctx, device); err != nil {
		return nil, err
	}

	var roleNames []string
	for _, roleID := range user.RoleIDs {
		role, err := h.roleRepo.GetByID(ctx, roleID)
		if err == nil && role != nil {
			roleNames = append(roleNames, role.Name)
		}
	}

	claims := auth.Claims{
		UserID:      userID,
		Username:    user.Username,
		Roles:       roleNames,
		Permissions: []string{}, // Middleware will fetch permissions based on Roles

		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, pubKeyHash)
	if err != nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: failed to issue access token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	refreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: failed to issue refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	// Save Refresh Token to Redis
	rtHash := auth.HashToken(refreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))
	rtValue := fmt.Sprintf("%s:%s", userIDStr, pubKeyHash)
	if err := h.rdb.Set(ctx, rtKey, rtValue, 7*24*time.Hour); err != nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: failed to save refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	// Cleanup challenge
	h.rdb.Native().Del(ctx, challengeKey)

	var roleIDs []string
	for _, id := range user.RoleIDs {
		roleIDs = append(roleIDs, id.String())
	}

	return &CompleteLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserPayload{
			ID:                user.ID.String(),
			Username:          user.Username,
			RoleIDs:           roleIDs,
			Roles:             roleNames,
			MFAEnabled:        user.MFAEnabled,
			PreferredLanguage: user.PreferredLanguage,
		},
	}, nil
}

func (h *CompleteLoginHandler) verifySignature(payload string, sigBase64 string, pubKeyPEM string) error {
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return err
	}

	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: failed to parse PEM block containing the public key", slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.ErrorAsync("CompleteLoginHandler.Handle: failed to parse PKIX public key", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return err
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		logger.ErrorAsync("CompleteLoginHandler.Handle: not ECDSA public key", slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return errors.New("not ECDSA public key")
	}

	hash := sha256.Sum256([]byte(payload))
	valid := ecdsa.VerifyASN1(ecdsaPub, hash[:], sigBytes)
	if !valid {
		logger.ErrorAsync("CompleteLoginHandler.Handle: invalid signature", slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return errors.New("invalid signature")
	}
	return nil
}
