package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/handlers/auth"
	"his-system/internal/identity/domain"
	"his-system/pkg/middleware"
)

type Router struct {
	desktopHandler *auth.DesktopAuthHandler
	webHandler     *auth.WebAuthHandler
	deviceRepo     domain.DeviceRepository
	signKey        []byte
	encKey         []byte
	rdb            *commonRedis.Client
}

type RouterDeps struct {
	DesktopHandler *auth.DesktopAuthHandler
	WebHandler     *auth.WebAuthHandler
	DeviceRepo     domain.DeviceRepository
	SignKey        []byte
	EncKey         []byte
	Rdb            *commonRedis.Client
}

func NewRouter(deps RouterDeps) *Router {
	return &Router{
		desktopHandler: deps.DesktopHandler,
		webHandler:     deps.WebHandler,
		deviceRepo:     deps.DeviceRepo,
		signKey:        deps.SignKey,
		encKey:         deps.EncKey,
		rdb:            deps.Rdb,
	}
}

func (r *Router) Register(root fiber.Router) {
	root.Use(middleware.AuthRateLimit(r.rdb))

	root.Post("/login/init", r.desktopHandler.InitLogin)
	root.Post("/login/complete", r.desktopHandler.CompleteLogin)
	root.Post("/mfa/verify", r.desktopHandler.VerifyMFA)
	root.Post("/refresh", r.desktopHandler.RefreshToken)
	root.Post("/logout", r.desktopHandler.Logout)

	root.Post("/otp/send", r.webHandler.SendOTP)
	root.Post("/otp/verify", r.webHandler.VerifyOTP)
	root.Post("/register", r.webHandler.Register)

	web := root.Group("/web")
	web.Post("/refresh", r.webHandler.RefreshWeb)
	web.Post("/logout", r.webHandler.LogoutWeb)

	jwtAuth := middleware.JWTAuth(r.signKey, r.encKey)
	root.Post("/mfa/setup",
		jwtAuth,
		middleware.RequestSignature(r.deviceRepo),
		r.desktopHandler.SetupMFA,
	)
	root.Put("/me/language",
		jwtAuth,
		middleware.RequestSignature(r.deviceRepo),
		r.desktopHandler.UpdateLanguage,
	)
}
