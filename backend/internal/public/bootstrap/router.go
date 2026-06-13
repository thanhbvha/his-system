package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"his-system/internal/public/handlers"
)

type Router struct {
	handler *handlers.PublicHandler
}

func NewRouter(handler *handlers.PublicHandler) *Router {
	return &Router{handler: handler}
}

func (r *Router) Register(root fiber.Router) {
	root.Get("/clinic-info", r.handler.GetClinicInfo)
	root.Get("/doctors", r.handler.ListDoctors)
	root.Get("/services", r.handler.ListServices)
}
