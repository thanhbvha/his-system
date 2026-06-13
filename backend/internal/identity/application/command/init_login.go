package command

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/thanhbvha/go-common/logger"
	commonRedis "github.com/thanhbvha/go-common/redis"
	"golang.org/x/crypto/bcrypt"

	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
)

type InitLoginCommand struct {
	Username string
	Password string
	ClientIP string
}

type InitLoginResult struct {
	ChallengeString string `json:"challenge_string"`
	MFARequired     bool   `json:"mfa_required"`
}

type InitLoginHandler struct {
	userRepo domain.UserRepository
	rdb      *commonRedis.Client
}

func NewInitLoginHandler(userRepo domain.UserRepository, rdb *commonRedis.Client) *InitLoginHandler {
	return &InitLoginHandler{userRepo: userRepo, rdb: rdb}
}

func (h *InitLoginHandler) Handle(ctx context.Context, cmd InitLoginCommand) (*InitLoginResult, error) {
	// Check brute force attempts
	attemptsKey := h.rdb.BuildKey(fmt.Sprintf("login_attempts:%s", cmd.Username))
	attemptsStr, _ := h.rdb.Get(ctx, attemptsKey)
	attempts, _ := strconv.Atoi(attemptsStr)
	if attempts >= 5 {
		return nil, &appErrors.AppError{
			Code:    "TOO_MANY_REQUESTS",
			Status:  429,
			Message: "Tài khoản tạm thời bị khóa do sai mật khẩu quá nhiều lần. Vui lòng thử lại sau 15 phút.",
		}
	}

	user, err := h.userRepo.GetByUsername(ctx, cmd.Username)
	if err != nil || user == nil {
		h.incrementAttempts(ctx, attemptsKey)

		logger.ErrorAsync("InitLoginHandler.Handle: user not found", slog.String("error", err.Error()), slog.String("username", cmd.Username), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))

		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Tên đăng nhập hoặc mật khẩu không chính xác"}
	}

	if !user.IsActive {
		return nil, &appErrors.AppError{Code: "FORBIDDEN", Status: 403, Message: "Tài khoản đã bị vô hiệu hóa"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		h.incrementAttempts(ctx, attemptsKey)
		logger.ErrorAsync("InitLoginHandler.Handle: invalid password", slog.String("error", err.Error()), slog.String("username", cmd.Username), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, &appErrors.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Tên đăng nhập hoặc mật khẩu không chính xác"}
	}

	// Password correct, clear attempts
	h.rdb.Native().Del(ctx, attemptsKey)

	// Generate Challenge String (32 bytes)
	b := make([]byte, 32)
	rand.Read(b)
	challengeStr := base64.RawURLEncoding.EncodeToString(b)

	// Store challenge in Redis
	// Key: challenge:{challengeStr} -> Value: {userID}
	challengeKey := h.rdb.BuildKey(fmt.Sprintf("challenge:%s", challengeStr))
	if err := h.rdb.Set(ctx, challengeKey, user.ID.String(), 5*time.Minute); err != nil {
		logger.ErrorAsync("InitLoginHandler.Handle: failed to save challenge", slog.String("error", err.Error()), slog.String("username", cmd.Username), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	return &InitLoginResult{
		ChallengeString: challengeStr,
		MFARequired:     user.MFAEnabled,
	}, nil
}

func (h *InitLoginHandler) incrementAttempts(ctx context.Context, key string) {
	h.rdb.Incr(ctx, key, 15*time.Minute)
}
