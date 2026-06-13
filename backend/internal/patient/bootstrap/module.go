package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/patient/handlers"
	patientCmd "his-system/internal/patient/application/command"
	patientQuery "his-system/internal/patient/application/query"
	patientInfra "his-system/internal/patient/infrastructure"
	identityInfra "his-system/internal/identity/infrastructure"
	"his-system/pkg/crypto"
)

type ModuleDeps struct {
	PgPool  *pgxpool.Pool
	Rdb     *commonRedis.Client
	Cipher  *crypto.FieldCipher
	SignKey []byte
	EncKey  []byte
}

type Module struct {
	router *Router
}

func NewModule(deps ModuleDeps) *Module {
	patientDomainRepo := patientInfra.NewPatientRepositoryPG(deps.PgPool)
	deviceRepo        := identityInfra.NewDeviceRepositoryPG(deps.PgPool)

	createPatientCmd := patientCmd.NewCreatePatientHandler(patientDomainRepo, deps.Cipher, deps.Rdb)
	updatePatientCmd := patientCmd.NewUpdatePatientHandler(patientDomainRepo, deps.Cipher, deps.Rdb)
	updateInsCmd     := patientCmd.NewUpdateInsuranceHandler(patientDomainRepo, deps.Cipher)

	searchPatientsQ    := patientQuery.NewSearchPatientsHandler(patientDomainRepo, deps.Cipher)
	getPatientByIDQ    := patientQuery.NewGetPatientByIDHandler(patientDomainRepo, deps.Cipher)
	getPatientHistoryQ := patientQuery.NewGetPatientHistoryHandler()

	handler := handlers.NewPatientHandler(createPatientCmd, updatePatientCmd, updateInsCmd, searchPatientsQ, getPatientByIDQ, getPatientHistoryQ)

	router := NewRouter(RouterDeps{
		Handler:    handler,
		DeviceRepo: deviceRepo,
		SignKey:    deps.SignKey,
		EncKey:     deps.EncKey,
	})

	return &Module{router: router}
}

func (m *Module) Register(root fiber.Router) {
	m.router.Register(root)
}
