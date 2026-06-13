package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	commonLogger "github.com/thanhbvha/go-common/logger"
	"his-system/pkg/errors"
	"his-system/pkg/response"
)

// Recover là middleware để bắt các panic trong quá trình xử lý request,
// đảm bảo server không bị crash và trả về HTTP 500 chuẩn.
func Recover(logger *commonLogger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				errStr := fmt.Sprintf("%v", r)

				if logger != nil {
					logger.ErrorAsync("Panic recovered",
						"error", errStr,
						"path", c.Path(),
						"method", c.Method(),
						"ip", c.IP(),
						"stack", stack,
						"dispatch_time", time.Now().Format(time.RFC3339Nano),
					)
				}

				// Trả về lỗi 500 theo chuẩn hệ thống
				_ = response.Fail(c, errors.ErrInternal)
			}
		}()

		return c.Next()
	}
}
