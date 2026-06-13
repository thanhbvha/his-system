package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

// RequireRole checks if the authenticated user has at least one of the required roles.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		if !ok {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// Check if user has any of the allowed roles
		for _, requiredRole := range roles {
			for _, userRole := range claims.Roles {
				if userRole == requiredRole {
					return c.Next()
				}
			}
		}

		return response.Fail(c, appErrors.ErrForbidden)
	}
}

// RequirePermission checks if the authenticated user has the required permission
// by looking up the permissions of their roles from Redis cache or Database.
func RequirePermission(permission string, rdb *commonRedis.Client, roleRepo domain.RoleRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		if !ok {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		ctx := c.Context()

		// Allow if Admin
		for _, role := range claims.Roles {
			if role == "admin" {
				return c.Next() // Admins have all permissions implicitly
			}
		}

		for _, roleName := range claims.Roles {
			if checkRolePermission(ctx, roleName, permission, rdb, roleRepo) {
				return c.Next()
			}
		}

		return response.Fail(c, appErrors.ErrForbidden)
	}
}

func checkRolePermission(ctx context.Context, roleName, requiredPerm string, rdb *commonRedis.Client, roleRepo domain.RoleRepository) bool {
	cacheKey := fmt.Sprintf("perm:%s", roleName)
	if rdb != nil {
		cacheKey = rdb.BuildKey(cacheKey)
	}

	// 1. Check Redis Cache
	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var perms []string
			if err := json.Unmarshal([]byte(cached), &perms); err == nil {
				return hasPermission(perms, requiredPerm)
			}
		}
	}

	// 2. Cache Miss: Query DB
	role, err := roleRepo.GetByName(ctx, roleName)
	if err != nil || role == nil {
		return false
	}

	// Extract permissions
	perms := make([]string, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		perms = append(perms, fmt.Sprintf("%s:%s", p.Resource, p.Action))
	}

	// 3. Update Cache
	if rdb != nil {
		if b, err := json.Marshal(perms); err == nil {
			// TTL 5 minutes
			rdb.Set(ctx, cacheKey, string(b), 5*time.Minute)
		}
	}

	return hasPermission(perms, requiredPerm)
}

func hasPermission(perms []string, req string) bool {
	for _, p := range perms {
		if p == req || p == "*" {
			return true
		}
	}
	return false
}
