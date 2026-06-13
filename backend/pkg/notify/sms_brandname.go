package notify

import (
	"context"
	"time"

	"github.com/thanhbvha/go-common/logger"
)

type SMSClient interface {
	SendBrandname(ctx context.Context, phone, message string) error
}

type mockSMSClient struct{}

func NewSMSClient() SMSClient {
	return &mockSMSClient{}
}

func (c *mockSMSClient) SendBrandname(ctx context.Context, phone, message string) error {
	logger.InfoAsync("[MOCK SMS BRANDNAME] Sending SMS",
		"phone", phone,
		"message", message,
		"dispatch_time", time.Now().Format(time.RFC3339Nano),
	)
	return nil // Simulate success
}
