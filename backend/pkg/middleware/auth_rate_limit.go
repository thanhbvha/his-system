package middleware

import (
	"context"
	"fmt"
	"time"

	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"

	"github.com/gofiber/fiber/v2"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

// AuthRateLimit restricts login attempts to 5 per minute per IP address.
func AuthRateLimit(rdb *commonRedis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := rdb.BuildKey(fmt.Sprintf("rl:login:%s", ip))

		ctx := context.Background()

		count, err := rdb.Incr(ctx, key, 1*time.Minute)
		if err != nil {
			// Fail open on Redis errors to prevent locking out all users
			return c.Next()
		}

		if count > 5 {
			return response.Fail(c, &appErrors.AppError{
				Code:    "TOO_MANY_REQUESTS",
				Status:  429,
				Message: "Quá nhiều yêu cầu đăng nhập. Vui lòng thử lại sau 1 phút.",
			})
		}

		return c.Next()
	}
}
