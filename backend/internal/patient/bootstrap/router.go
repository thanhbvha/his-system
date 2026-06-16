package bootstrap

import (
	"github.com/gofiber/fiber/v2"

	"his-system/internal/patient/handlers"
	"his-system/internal/identity/domain"
	"his-system/pkg/middleware"
)

type Router struct {
	handler    *handlers.PatientHandler
	deviceRepo domain.DeviceRepository
	signKey    []byte
	encKey     []byte
}

type RouterDeps struct {
	Handler    *handlers.PatientHandler
	DeviceRepo domain.DeviceRepository
	SignKey    []byte
	EncKey     []byte
}

func NewRouter(deps RouterDeps) *Router {
	return &Router{
		handler:    deps.Handler,
		deviceRepo: deps.DeviceRepo,
		signKey:    deps.SignKey,
		encKey:     deps.EncKey,
	}
}

func (r *Router) Register(root fiber.Router) {
	sig := middleware.RequestSignature(r.deviceRepo)

	root.Get("/me",
		middleware.RequireRole("patient"),
		r.handler.GetMyProfile,
	)
	root.Put("/me",
		middleware.RequireRole("patient"),
		r.handler.UpdateMyProfile,
	)
	root.Put("/me/insurance",
		middleware.RequireRole("patient"),
		r.handler.UpdateMyInsurance,
	)

	root.Get("/",
		middleware.RequireRole("admin", "receptionist", "doctor", "nurse", "lab_tech", "pharmacist"),
		r.handler.List,
	)
	root.Post("/",
		middleware.RequireRole("receptionist", "admin"),
		sig,
		r.handler.Create,
	)
	root.Get("/:id",
		middleware.RequireRole("admin", "receptionist", "doctor", "nurse"),
		r.handler.GetByID,
	)
	root.Put("/:id",
		middleware.RequireRole("receptionist", "admin"),
		sig,
		r.handler.Update,
	)
}
