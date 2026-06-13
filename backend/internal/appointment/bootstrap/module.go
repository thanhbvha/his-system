package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/appointment/handlers"
	apptCmd "his-system/internal/appointment/application/command"
	apptQuery "his-system/internal/appointment/application/query"
	apptInfra "his-system/internal/appointment/infrastructure"
	identityInfra "his-system/internal/identity/infrastructure"
)

type ModuleDeps struct {
	PgPool *pgxpool.Pool
	Rdb    *commonRedis.Client
}

type Module struct {
	router *Router
}

func NewModule(deps ModuleDeps) *Module {
	apptDomainRepo := apptInfra.NewAppointmentRepositoryPG(deps.PgPool)
	slotDomainRepo := apptInfra.NewSlotRepositoryPG(deps.PgPool)
	deviceRepo     := identityInfra.NewDeviceRepositoryPG(deps.PgPool)

	bookApptCmd    := apptCmd.NewBookAppointmentHandler(apptDomainRepo, slotDomainRepo, deps.Rdb)
	cancelApptCmd  := apptCmd.NewCancelAppointmentHandler(apptDomainRepo, slotDomainRepo, deps.Rdb)
	confirmApptCmd := apptCmd.NewConfirmAppointmentHandler(apptDomainRepo, deps.Rdb)

	slotsQ     := apptQuery.NewGetAvailableSlotsHandler(slotDomainRepo)
	listApptsQ := apptQuery.NewListAppointmentsHandler(apptDomainRepo)

	handler := handlers.NewAppointmentHandler(bookApptCmd, cancelApptCmd, confirmApptCmd, slotsQ, listApptsQ)

	router := NewRouter(RouterDeps{
		Handler:    handler,
		DeviceRepo: deviceRepo,
	})

	return &Module{router: router}
}

func (m *Module) Register(root fiber.Router) {
	m.router.Register(root)
}
