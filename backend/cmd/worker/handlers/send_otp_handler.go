package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/queue"
	"his-system/pkg/notify"
)

type SendOTPHandler struct {
	zaloClient notify.ZaloClient
	smsClient  notify.SMSClient
}

func NewSendOTPHandler(zaloClient notify.ZaloClient, smsClient notify.SMSClient) *SendOTPHandler {
	return &SendOTPHandler{
		zaloClient: zaloClient,
		smsClient:  smsClient,
	}
}

func (h *SendOTPHandler) Handle(job queue.Job) error {
	ctx := context.Background() // Or pass tracing context if queue supports it

	data, ok := job.Data.(map[string]interface{})
	if !ok {
		logger.ErrorAsync("Invalid job data format", "dispatch_time", time.Now().Format(time.RFC3339Nano))
		return nil
	}

	phone, ok := data["phone"].(string)
	if !ok {
		logger.ErrorAsync("Invalid job payload: missing phone", "dispatch_time", time.Now().Format(time.RFC3339Nano))
		return nil // Return nil so it won't retry a bad payload
	}

	otp, ok := data["otp"].(string)
	if !ok {
		logger.ErrorAsync("Invalid job payload: missing otp", "dispatch_time", time.Now().Format(time.RFC3339Nano))
		return nil
	}

	// 1. Try Zalo ZNS first
	err := h.zaloClient.SendOTP(ctx, phone, otp)
	if err == nil {
		return nil // Success
	}

	logger.WarnAsync("Failed to send OTP via Zalo, falling back to SMS",
		"error", err.Error(),
		"phone", phone,
		"dispatch_time", time.Now().Format(time.RFC3339Nano),
	)

	// 2. Fallback to SMS Brandname
	message := fmt.Sprintf("Ma OTP HIS cua ban la: %s. Co hieu luc trong 5 phut.", otp)
	err = h.smsClient.SendBrandname(ctx, phone, message)
	if err != nil {
		logger.ErrorAsync("Failed to send OTP via SMS",
			"error", err.Error(),
			"phone", phone,
			"dispatch_time", time.Now().Format(time.RFC3339Nano),
		)
		return err // Return error to trigger queue retry
	}

	return nil
}
