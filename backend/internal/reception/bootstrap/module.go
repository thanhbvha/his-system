package bootstrap

import (
	"his-system/internal/reception/handlers"
	"his-system/internal/reception/infrastructure"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceptionModule struct {
	QueueHandler *handlers.QueueHandler
}

func NewReceptionModule(db *pgxpool.Pool) *ReceptionModule {
	repo := infrastructure.NewQueueRepositoryPG(db)
	queueHandler := handlers.NewQueueHandler(repo)

	return &ReceptionModule{
		QueueHandler: queueHandler,
	}
}

func (m *ReceptionModule) RegisterRoutes(app fiber.Router) {
	// API routes
	api := app.Group("/")

	// Assuming these are protected by standard RBAC middleware further up in main.go
	api.Get("/", m.QueueHandler.GetCurrentQueue)
	api.Post("/checkin", m.QueueHandler.CheckIn)
	api.Post("/call/:id", m.QueueHandler.CallQueue)
	api.Post("/skip/:id", m.QueueHandler.SkipQueue)
	api.Post("/complete/:id", m.QueueHandler.CompleteQueue)
	api.Get("/stats", m.QueueHandler.GetQueueStats)
}
