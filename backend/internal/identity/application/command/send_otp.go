package command

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/thanhbvha/go-common/queue"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/pkg/crypto"
	appErrors "his-system/pkg/errors"
)

var rawPhoneRegex = regexp.MustCompile(`^0\d{9}$`)

type SendOTPCommand struct {
	Phone    string // raw phone, e.g., "0912345678"
	ClientIP string
}

type SendOTPHandler struct {
	rdb    *commonRedis.Client
	q      *queue.Queue
	cipher *crypto.FieldCipher
}

func NewSendOTPHandler(rdb *commonRedis.Client, q *queue.Queue, cipher *crypto.FieldCipher) *SendOTPHandler {
	return &SendOTPHandler{rdb: rdb, q: q, cipher: cipher}
}

func (h *SendOTPHandler) Handle(ctx context.Context, cmd SendOTPCommand) error {
	// 1. Validate phone format
	if !rawPhoneRegex.MatchString(cmd.Phone) {
		return &appErrors.AppError{Code: "INVALID_PHONE", Status: 422, Message: "Số điện thoại không hợp lệ"}
	}

	// 2. Compute phone HMAC for rate limiting & OTP lookup
	phoneHMAC := h.cipher.HMAC(cmd.Phone)

	// 3. Rate Limit check: max 3 requests per hour
	rlKey := h.rdb.BuildKey(fmt.Sprintf("rl:otp:%s", phoneHMAC))
	count, err := h.rdb.Incr(ctx, rlKey, 1*time.Hour)
	if err != nil {
		return err
	}
	if count > 3 {
		return &appErrors.AppError{Code: "TOO_MANY_REQUESTS", Status: 429, Message: "Bạn đã yêu cầu OTP quá nhiều lần. Vui lòng thử lại sau."}
	}

	// 4. Generate 6-digit OTP
	otp, err := generateOTP()
	if err != nil {
		return err
	}

	// 5. Store OTP in Redis (TTL 5 minutes)
	otpKey := h.rdb.BuildKey(fmt.Sprintf("otp:%s", phoneHMAC))
	// Store as "otp:attempts"
	val := fmt.Sprintf("%s:0", otp)
	if err := h.rdb.Set(ctx, otpKey, val, 5*time.Minute); err != nil {
		return err
	}

	// 6. Enqueue job
	if h.q != nil {
		payload := map[string]interface{}{
			"phone": cmd.Phone,
			"otp":   otp,
		}
		if err := h.q.Enqueue(ctx, "send_otp", payload); err != nil {
			return err
		}
	}

	return nil
}

func generateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
