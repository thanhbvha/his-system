package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

// JWTAuth extracts and verifies JWT from the Authorization header.
func JWTAuth(signingKey, encKey []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		tokenStr := parts[1]
		claims, err := auth.VerifyAccessToken(tokenStr, signingKey, encKey)
		if err != nil {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// Store claims in context locals
		c.Locals("claims", claims)

		return c.Next()
	}
}

// GetClaims retrieves auth.Claims from Fiber context.
func GetClaims(c *fiber.Ctx) (auth.Claims, bool) {
	claims, ok := c.Locals("claims").(auth.Claims)
	return claims, ok
}
