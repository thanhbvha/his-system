package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	commonLogger "github.com/thanhbvha/go-common/logger"
)

// RequestLogger là middleware ghi nhận thông tin request/response
// vào hệ thống logger chung.
func RequestLogger(logger *commonLogger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Bỏ qua log cho health check để tránh spam
		if c.Path() == "/health" {
			return c.Next()
		}

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Nếu có lỗi từ Fiber trả về (ví dụ: 404, 405)
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			}
		}

		if logger != nil {
			logMsg := "HTTP Request"
			args := []any{
				"request_id", c.Locals("requestid"),
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.IP(),
				"user_agent", string(c.Request().Header.UserAgent()),
			}

			// Ghi log mức Error nếu HTTP Status >= 500, ngược lại Info
			if status >= 500 {
				if err != nil {
					args = append(args, "error", err.Error())
				}
				logger.ErrorAsync(logMsg, args...)
			} else {
				logger.InfoAsync(logMsg, args...)
			}
		}

		return err
	}
}
