package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	commonQueue "github.com/thanhbvha/go-common/queue"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/handlers/auth"
	"his-system/internal/identity/application/command"
	identityInfra "his-system/internal/identity/infrastructure"
	"his-system/pkg/crypto"
	"his-system/pkg/middleware"
)

type ModuleDeps struct {
	PgPool  *pgxpool.Pool
	Rdb     *commonRedis.Client
	Queue   *commonQueue.Queue
	Cipher  *crypto.FieldCipher
	SignKey []byte
	EncKey  []byte
}

type Module struct {
	router *Router
}

func NewModule(deps ModuleDeps) *Module {
	userRepo      := identityInfra.NewUserRepositoryPG(deps.PgPool)
	deviceRepo    := identityInfra.NewDeviceRepositoryPG(deps.PgPool)
	roleRepo      := identityInfra.NewRoleRepositoryPG(deps.PgPool)
	mfaRepo       := identityInfra.NewMFARepositoryPG(deps.PgPool)
	patientAuthRepo := identityInfra.NewPatientRepositoryPG(deps.PgPool)

	initLoginCmd     := command.NewInitLoginHandler(userRepo, deps.Rdb)
	completeLoginCmd := command.NewCompleteLoginHandler(userRepo, deviceRepo, roleRepo, deps.Rdb, deps.SignKey, deps.EncKey)
	refreshTokenCmd  := command.NewRefreshTokenHandler(userRepo, roleRepo, deps.Rdb, deps.SignKey, deps.EncKey)
	logoutCmd        := command.NewLogoutHandler(deps.Rdb)
	setupMFACmd      := command.NewSetupMFAHandler(mfaRepo, deps.EncKey)
	verifyMFACmd     := command.NewVerifyMFAHandler(mfaRepo, userRepo, deps.Rdb, deps.EncKey)

	sendOTPCmd         := command.NewSendOTPHandler(deps.Rdb, deps.Queue, deps.Cipher)
	verifyOTPCmd       := command.NewVerifyOTPHandler(patientAuthRepo, deps.Rdb, deps.Cipher, deps.SignKey, deps.EncKey)
	registerPatientCmd := command.NewRegisterPatientHandler(userRepo, patientAuthRepo, deps.Rdb, deps.Cipher, deps.SignKey, deps.EncKey)
	refreshWebCmd      := command.NewRefreshWebHandler(patientAuthRepo, deps.Rdb, deps.SignKey, deps.EncKey)
	logoutWebCmd       := command.NewLogoutWebHandler(deps.Rdb)

	desktopHandler := auth.NewDesktopAuthHandler(initLoginCmd, completeLoginCmd, refreshTokenCmd, logoutCmd, setupMFACmd, verifyMFACmd)
	webHandler     := auth.NewWebAuthHandler(sendOTPCmd, verifyOTPCmd, registerPatientCmd, refreshWebCmd, logoutWebCmd)

	router := NewRouter(RouterDeps{
		DesktopHandler: desktopHandler,
		WebHandler:     webHandler,
		DeviceRepo:     deviceRepo,
		SignKey:        deps.SignKey,
		EncKey:         deps.EncKey,
		Rdb:            deps.Rdb,
	})

	return &Module{router: router}
}

func (m *Module) Register(root fiber.Router) {
	m.router.Register(root)
}

func JWTAuth(signKey, encKey []byte) fiber.Handler {
	return middleware.JWTAuth(signKey, encKey)
}
