package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/handlers/admin"
	"his-system/internal/identity/application/command"
	identityQuery "his-system/internal/identity/application/query"
	identityInfra "his-system/internal/identity/infrastructure"
	"his-system/pkg/crypto"
	"his-system/pkg/middleware"
)

type ModuleDeps struct {
	PgPool *pgxpool.Pool
	Rdb    *commonRedis.Client
	Cipher *crypto.FieldCipher
}

type Module struct {
	router     *Router
	deviceRepo identityInfra.DeviceRepositoryPG
}

func NewModule(deps ModuleDeps) *Module {
	userRepo   := identityInfra.NewUserRepositoryPG(deps.PgPool)
	roleRepo   := identityInfra.NewRoleRepositoryPG(deps.PgPool)
	deviceRepo := identityInfra.NewDeviceRepositoryPG(deps.PgPool)
	deptRepo   := identityInfra.NewDepartmentRepositoryPG(deps.PgPool)

	getUserByIDQuery  := identityQuery.NewGetUserByIDHandler(userRepo, roleRepo, deps.Cipher)
	listUsersQuery    := identityQuery.NewListUsersHandler(userRepo, roleRepo, deps.Cipher, deps.PgPool)
	getRolePermsQuery := identityQuery.NewGetRolePermissionsHandler(roleRepo)
	listPermsQuery    := identityQuery.NewListPermissionsHandler(roleRepo)

	createStaffCmd     := command.NewCreateStaffHandler(userRepo, deps.Cipher, deps.PgPool)
	deactivateCmd      := command.NewDeactivateUserHandler(userRepo, deviceRepo)
	assignRolesCmd     := command.NewAssignUserRolesHandler(userRepo)
	updateProfileCmd   := command.NewUpdateStaffProfileHandler(deps.PgPool)
	updateEmailCmd     := command.NewUpdateUserEmailHandler(userRepo, deps.Cipher)
	updateRolePermsCmd := command.NewUpdateRolePermissionsHandler(roleRepo, deps.Rdb)
	createDeptCmd      := command.NewCreateDepartmentHandler(deptRepo)

	userHandler := admin.NewAdminUserHandler(getUserByIDQuery, listUsersQuery, createStaffCmd, deactivateCmd, assignRolesCmd, updateProfileCmd, updateEmailCmd)
	roleHandler := admin.NewAdminRoleHandler(roleRepo, getRolePermsQuery, updateRolePermsCmd, listPermsQuery)
	deptHandler := admin.NewAdminDepartmentHandler(deptRepo, createDeptCmd)

	router := NewRouter(RouterDeps{
		UserHandler: userHandler,
		RoleHandler: roleHandler,
		DeptHandler: deptHandler,
		RoleRepo:    roleRepo,
		Rdb:         deps.Rdb,
	})

	return &Module{
		router:     router,
		deviceRepo: *deviceRepo,
	}
}

func (m *Module) Register(root fiber.Router) {
	m.router.Register(root)
}

func DeviceMiddleware(pgPool *pgxpool.Pool) fiber.Handler {
	deviceRepo := identityInfra.NewDeviceRepositoryPG(pgPool)
	return middleware.RequestSignature(deviceRepo)
}
