package bootstrap

import (
	"his-system/internal/reception/handlers"
	"his-system/internal/reception/infrastructure"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/pkg/middleware"
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

	// Common read access
	api.Get("/", middleware.RequireRole("admin", "receptionist", "doctor", "nurse"), m.QueueHandler.GetCurrentQueue)
	api.Get("/stats", middleware.RequireRole("admin", "receptionist", "doctor", "nurse"), m.QueueHandler.GetQueueStats)

	// Receptionist and Admin can check in
	api.Post("/checkin", middleware.RequireRole("admin", "receptionist"), m.QueueHandler.CheckIn)

	// Doctors, Nurses, and Admins can manage queue
	api.Post("/call/:id", middleware.RequireRole("admin", "doctor", "nurse"), m.QueueHandler.CallQueue)
	api.Post("/skip/:id", middleware.RequireRole("admin", "doctor", "nurse"), m.QueueHandler.SkipQueue)
	api.Post("/complete/:id", middleware.RequireRole("admin", "doctor", "nurse"), m.QueueHandler.CompleteQueue)
}
