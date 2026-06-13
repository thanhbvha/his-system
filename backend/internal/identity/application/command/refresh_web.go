package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
)

type RefreshWebCommand struct {
	RefreshToken string
}

type RefreshWebResult struct {
	AccessToken string `json:"access_token"`
}

type RefreshWebHandler struct {
	patientRepo domain.PatientRepository
	rdb         *commonRedis.Client
	signKey     []byte
	encKey      []byte
}

func NewRefreshWebHandler(patientRepo domain.PatientRepository, rdb *commonRedis.Client, signKey, encKey []byte) *RefreshWebHandler {
	return &RefreshWebHandler{
		patientRepo: patientRepo,
		rdb:         rdb,
		signKey:     signKey,
		encKey:      encKey,
	}
}

// Handle returns (result, newRefreshToken, error)
func (h *RefreshWebHandler) Handle(ctx context.Context, cmd RefreshWebCommand) (*RefreshWebResult, string, error) {
	if cmd.RefreshToken == "" {
		return nil, "", &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Yêu cầu refresh token"}
	}

	rtHash := auth.HashToken(cmd.RefreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))

	val, err := h.rdb.Get(ctx, rtKey)
	if err != nil || val == "" {
		return nil, "", &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Refresh token hết hạn hoặc không tồn tại"}
	}

	userIDStr := val // For web, we only stored userID (no pubKeyHash)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, "", err
	}

	patient, err := h.patientRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if patient == nil {
		return nil, "", &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Tài khoản không tồn tại"}
	}

	// Rotate token
	h.rdb.Delete(ctx, rtKey)

	newRefreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		return nil, "", err
	}

	newRtHash := auth.HashToken(newRefreshToken)
	newRtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", newRtHash))
	if err := h.rdb.Set(ctx, newRtKey, userIDStr, 7*24*time.Hour); err != nil {
		return nil, "", err
	}

	// Issue new access token
	claims := auth.Claims{
		UserID: patient.ID,
		Roles:  []string{"patient"},
	}
	accessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, "")
	if err != nil {
		return nil, "", err
	}

	return &RefreshWebResult{
		AccessToken: accessToken,
	}, newRefreshToken, nil
}
