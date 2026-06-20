package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	commonQueue "github.com/thanhbvha/go-common/queue"

	"his-system/internal/visit/handlers"
	"his-system/internal/visit/infrastructure"
)

type VisitModule struct {
	Handler *handlers.VisitHandler
}

func NewVisitModule(db *pgxpool.Pool, q *commonQueue.Queue) *VisitModule {
	repo := infrastructure.NewVisitRepositoryPG(db)
	handler := handlers.NewVisitHandler(repo, q)
	return &VisitModule{Handler: handler}
}

func (m *VisitModule) RegisterRoutes(router fiber.Router) {
	// Visit CRUD
	router.Get("/", m.Handler.GetWorklist)
	router.Post("/", m.Handler.CreateVisit)
	router.Get("/:id", m.Handler.GetVisitDetail)
	router.Put("/:id/status", m.Handler.UpdateVisitStatus)

	// Vitals
	router.Post("/:id/vitals", m.Handler.RecordVitals)
	router.Get("/:id/vitals", m.Handler.GetVitals)

	// Orders
	router.Post("/:id/orders", m.Handler.CreateOrder)
	router.Get("/:id/orders", m.Handler.GetOrders)

	// Close visit
	router.Post("/:id/close", m.Handler.CloseVisit)
}
