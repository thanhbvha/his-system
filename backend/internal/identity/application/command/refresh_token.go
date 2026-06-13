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
	"strings"
	"time"

	"github.com/google/uuid"
	commonRedis "github.com/thanhbvha/go-common/redis"
	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
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
	rdb      *commonRedis.Client
	signKey  []byte
	encKey   []byte
}

func NewRefreshTokenHandler(userRepo domain.UserRepository, rdb *commonRedis.Client, signKey, encKey []byte) *RefreshTokenHandler {
	return &RefreshTokenHandler{userRepo: userRepo, rdb: rdb, signKey: signKey, encKey: encKey}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	// 1. Verify Signature
	if err := h.verifySignature(cmd.RefreshToken, cmd.Signature, cmd.PublicKeyPEM); err != nil {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Chữ ký không hợp lệ"}
	}

	// 2. Lookup in Redis
	rtHash := auth.HashToken(cmd.RefreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))

	val, err := h.rdb.Get(ctx, rtKey)
	if err != nil || val == "" {
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
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Public key không khớp với token gốc"}
	}

	// 4. Issue new tokens
	userID, _ := uuid.Parse(userIDStr)
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Người dùng không tồn tại hoặc bị khóa"}
	}

	claims := auth.Claims{
		UserID:    userID,
		Username:  user.Username,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}

	newAccessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, pubKeyHash)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		return nil, err
	}

	// Rotate in Redis: delete old, set new
	h.rdb.Delete(ctx, rtKey)

	newRtHash := auth.HashToken(newRefreshToken)
	newRtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", newRtHash))
	if err := h.rdb.Set(ctx, newRtKey, val, 7*24*time.Hour); err != nil {
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
