package notify

import (
	"context"
	"time"

	"github.com/thanhbvha/go-common/logger"
)

type ZaloClient interface {
	SendOTP(ctx context.Context, phone, otp string) error
}

type mockZaloClient struct{}

func NewZaloClient() ZaloClient {
	return &mockZaloClient{}
}

func (c *mockZaloClient) SendOTP(ctx context.Context, phone, otp string) error {
	logger.InfoAsync("[MOCK ZALO ZNS] Sending OTP",
		"phone", phone,
		"otp", otp,
		"dispatch_time", time.Now().Format(time.RFC3339Nano),
	)
	return nil // Simulate success
}
