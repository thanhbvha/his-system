package bootstrap

import (
	"github.com/gofiber/fiber/v2"

	"his-system/internal/appointment/handlers"
	"his-system/internal/identity/domain"
	"his-system/pkg/middleware"
)

type Router struct {
	handler    *handlers.AppointmentHandler
	deviceRepo domain.DeviceRepository
}

type RouterDeps struct {
	Handler    *handlers.AppointmentHandler
	DeviceRepo domain.DeviceRepository
}

func NewRouter(deps RouterDeps) *Router {
	return &Router{
		handler:    deps.Handler,
		deviceRepo: deps.DeviceRepo,
	}
}

func (r *Router) Register(root fiber.Router) {
	sig := middleware.RequestSignature(r.deviceRepo)

	root.Get("/slots", r.handler.GetAvailableSlots)
	root.Get("/", r.handler.List)
	root.Post("/",
		middleware.RequireRole("patient", "receptionist", "admin"),
		r.handler.Book,
	)
	root.Put("/:id/confirm",
		middleware.RequireRole("receptionist", "admin"),
		sig,
		r.handler.Confirm,
	)
	root.Delete("/:id",
		middleware.RequireRole("patient", "receptionist", "admin"),
		r.handler.Cancel,
	)
}
