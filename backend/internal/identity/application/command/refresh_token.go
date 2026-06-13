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
	"strings"
	"time"

	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"

	"github.com/google/uuid"
	"github.com/thanhbvha/go-common/logger"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type RefreshTokenCommand struct {
	RefreshToken string
	Signature    string // signature of RefreshToken string using the device's private key
	PublicKeyPEM string
}

type RefreshTokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenHandler struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
	rdb      *commonRedis.Client
	signKey  []byte
	encKey   []byte
}

func NewRefreshTokenHandler(userRepo domain.UserRepository, roleRepo domain.RoleRepository, rdb *commonRedis.Client, signKey, encKey []byte) *RefreshTokenHandler {
	return &RefreshTokenHandler{userRepo: userRepo, roleRepo: roleRepo, rdb: rdb, signKey: signKey, encKey: encKey}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	// 1. Verify Signature
	if err := h.verifySignature(cmd.RefreshToken, cmd.Signature, cmd.PublicKeyPEM); err != nil {
		logger.ErrorAsync("RefreshTokenHandler.Handle: invalid signature", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Chữ ký không hợp lệ"}
	}

	// 2. Lookup in Redis
	rtHash := auth.HashToken(cmd.RefreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))

	val, err := h.rdb.Get(ctx, rtKey)
	if err != nil || val == "" {
		logger.ErrorAsync("RefreshTokenHandler.Handle: refresh token not found or expired", slog.String("error", fmt.Sprintf("%v", err)), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Refresh token hết hạn hoặc không tồn tại"}
	}

	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid refresh token value in redis")
	}
	userIDStr := parts[0]
	storedPubKeyHash := parts[1]

	// 3. Verify Public Key Match
	hSha256 := sha256.New()
	hSha256.Write([]byte(cmd.PublicKeyPEM))
	pubKeyHash := hex.EncodeToString(hSha256.Sum(nil))

	if storedPubKeyHash != pubKeyHash {
		logger.ErrorAsync("RefreshTokenHandler.Handle: public key mismatch", slog.String("stored", storedPubKeyHash), slog.String("provided", pubKeyHash), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Public key không khớp với token gốc"}
	}

	// 4. Issue new tokens
	userID, _ := uuid.Parse(userIDStr)
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		logger.ErrorAsync("RefreshTokenHandler.Handle: user not found or inactive", slog.String("error", fmt.Sprintf("%v", err)), slog.String("user_id", userIDStr), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Người dùng không tồn tại hoặc bị khóa"}
	}

	var roleNames []string
	for _, roleID := range user.RoleIDs {
		role, err := h.roleRepo.GetByID(ctx, roleID)
		if err == nil && role != nil {
			roleNames = append(roleNames, role.Name)
		}
	}

	claims := auth.Claims{
		UserID:    userID,
		Username:  user.Username,
		Roles:     roleNames,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}

	newAccessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, pubKeyHash)
	if err != nil {
		logger.ErrorAsync("RefreshTokenHandler.Handle: failed to issue access token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	newRefreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		logger.ErrorAsync("RefreshTokenHandler.Handle: failed to issue refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	// Rotate in Redis: delete old, set new
	h.rdb.Native().Del(ctx, rtKey)

	newRtHash := auth.HashToken(newRefreshToken)
	newRtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", newRtHash))
	if err := h.rdb.Set(ctx, newRtKey, val, 7*24*time.Hour); err != nil {
		logger.ErrorAsync("RefreshTokenHandler.Handle: failed to save refresh token", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	return &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (h *RefreshTokenHandler) verifySignature(payload string, sigBase64 string, pubKeyPEM string) error {
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return err
	}

	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return errors.New("failed to parse PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not ECDSA")
	}

	hash := sha256.Sum256([]byte(payload))
	if !ecdsa.VerifyASN1(ecdsaPub, hash[:], sigBytes) {
		return errors.New("invalid signature")
	}
	return nil
}
