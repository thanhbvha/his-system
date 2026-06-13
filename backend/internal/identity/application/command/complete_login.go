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
	"time"

	"github.com/google/uuid"
	commonRedis "github.com/thanhbvha/go-common/redis"
	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CompleteLoginHandler struct {
	userRepo   domain.UserRepository
	deviceRepo domain.DeviceRepository
	rdb        *commonRedis.Client
	signKey    []byte
	encKey     []byte
}

func NewCompleteLoginHandler(userRepo domain.UserRepository, deviceRepo domain.DeviceRepository, rdb *commonRedis.Client, signKey, encKey []byte) *CompleteLoginHandler {
	return &CompleteLoginHandler{userRepo: userRepo, deviceRepo: deviceRepo, rdb: rdb, signKey: signKey, encKey: encKey}
}

func (h *CompleteLoginHandler) Handle(ctx context.Context, cmd CompleteLoginCommand) (*CompleteLoginResult, error) {
	challengeKey := h.rdb.BuildKey(fmt.Sprintf("challenge:%s", cmd.ChallengeString))
	userIDStr, err := h.rdb.Get(ctx, challengeKey)
	if err != nil || userIDStr == "" {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Challenge hết hạn hoặc không hợp lệ"}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Không tìm thấy người dùng"}
	}

	// Verify ECDSA Signature
	if err := h.verifySignature(cmd.ChallengeString, cmd.Signature, cmd.PublicKeyPEM); err != nil {
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
		h.rdb.Delete(ctx, mfaKey)
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

	claims := auth.Claims{
		UserID:      userID,
		Username:    user.Username,
		Roles:       []string{}, // TODO: Populate roles
		Permissions: []string{}, // TODO: Populate permissions
		IssuedAt:    time.Now().Unix(),
		ExpiresAt:   time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, pubKeyHash)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		return nil, err
	}

	// Save Refresh Token to Redis
	rtHash := auth.HashToken(refreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))
	rtValue := fmt.Sprintf("%s:%s", userIDStr, pubKeyHash)
	if err := h.rdb.Set(ctx, rtKey, rtValue, 7*24*time.Hour); err != nil {
		return nil, err
	}

	// Cleanup challenge
	h.rdb.Delete(ctx, challengeKey)

	return &CompleteLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *CompleteLoginHandler) verifySignature(payload string, sigBase64 string, pubKeyPEM string) error {
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return err
	}

	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not ECDSA public key")
	}

	hash := sha256.Sum256([]byte(payload))
	valid := ecdsa.VerifyASN1(ecdsaPub, hash[:], sigBytes)
	if !valid {
		return errors.New("invalid signature")
	}
	return nil
}
