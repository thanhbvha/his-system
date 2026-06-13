package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"his-system/internal/public/handlers"
)

type ModuleDeps struct{}

type Module struct {
	router *Router
}

func NewModule(deps ModuleDeps) *Module {
	handler := handlers.NewPublicHandler()
	router := NewRouter(handler)
	return &Module{router: router}
}

func (m *Module) Register(root fiber.Router) {
	m.router.Register(root)
}
