package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/handlers/admin"
	"his-system/internal/identity/domain"
	"his-system/pkg/middleware"
)

type Router struct {
	userHandler *admin.AdminUserHandler
	roleHandler *admin.AdminRoleHandler
	deptHandler *admin.AdminDepartmentHandler
	roleRepo    domain.RoleRepository
	rdb         *commonRedis.Client
}

type RouterDeps struct {
	UserHandler *admin.AdminUserHandler
	RoleHandler *admin.AdminRoleHandler
	DeptHandler *admin.AdminDepartmentHandler
	RoleRepo    domain.RoleRepository
	Rdb         *commonRedis.Client
}

func NewRouter(deps RouterDeps) *Router {
	return &Router{
		userHandler: deps.UserHandler,
		roleHandler: deps.RoleHandler,
		deptHandler: deps.DeptHandler,
		roleRepo:    deps.RoleRepo,
		rdb:         deps.Rdb,
	}
}

func (r *Router) Register(root fiber.Router) {
	users := root.Group("/users")
	users.Get("/", r.userHandler.List)
	users.Post("/", r.userHandler.Create)
	users.Get("/:id", r.userHandler.GetByID)
	users.Put("/:id/deactivate", r.userHandler.Deactivate)
	users.Put("/:id/roles",
		middleware.RequirePermission("user:assign_role", r.rdb, r.roleRepo),
		r.userHandler.AssignRoles,
	)
	users.Put("/:id/profile", r.userHandler.UpdateProfile)

	roles := root.Group("/roles")
	roles.Get("/", r.roleHandler.List)
	roles.Get("/:id/permissions", r.roleHandler.GetPermissions)
	roles.Put("/:id/permissions", r.roleHandler.UpdatePermissions)

	perms := root.Group("/permissions")
	perms.Get("/", r.roleHandler.ListPermissions)

	depts := root.Group("/departments")
	depts.Get("/", r.deptHandler.List)
	depts.Post("/", r.deptHandler.Create)
}
