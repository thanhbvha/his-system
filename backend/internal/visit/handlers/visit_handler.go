package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	commonQueue "github.com/thanhbvha/go-common/queue"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"

	"his-system/internal/visit/application/commands"
	"his-system/internal/visit/application/queries"
	"his-system/internal/visit/domain"
)

type VisitHandler struct {
	repo domain.VisitRepository
	q    *commonQueue.Queue
}

func NewVisitHandler(repo domain.VisitRepository, q *commonQueue.Queue) *VisitHandler {
	return &VisitHandler{repo: repo, q: q}
}

type CreateVisitReq struct {
	PatientID      uuid.UUID  `json:"patient_id"`
	DoctorID       uuid.UUID  `json:"doctor_id"`
	QueueEntryID   *uuid.UUID `json:"queue_entry_id,omitempty"`
	ChiefComplaint *string    `json:"chief_complaint,omitempty"`
}

// CreateVisit godoc
// @Summary Create visit
// @Description Create a new clinical visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param request body CreateVisitReq true "Visit Creation Payload"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits [post]
// @Security BearerAuth
func (h *VisitHandler) CreateVisit(c *fiber.Ctx) error {
	var req CreateVisitReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request body"})
	}
	cmd := commands.CreateVisitCommand{
		PatientID:      req.PatientID,
		DoctorID:       req.DoctorID,
		QueueEntryID:   req.QueueEntryID,
		ChiefComplaint: req.ChiefComplaint,
	}
	v, err := commands.HandleCreateVisit(c.Context(), cmd, h.repo, h.q)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, v)
}

// GetWorklist godoc
// @Summary Get doctor worklist
// @Description Retrieve the visit worklist for a specific doctor.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param doctor_id query string true "Doctor ID"
// @Param date query string false "Date (YYYY-MM-DD)"
// @Param status query string false "Filter by Status"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits [get]
// @Security BearerAuth
func (h *VisitHandler) GetWorklist(c *fiber.Ctx) error {
	doctorIDStr := c.Query("doctor_id")
	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid doctor_id"})
	}

	dateStr := c.Query("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid date format, use YYYY-MM-DD"})
	}

	status := domain.VisitStatus(c.Query("status"))

	q := queries.GetDoctorWorklistQuery{DoctorID: doctorID, Date: date, Status: status}
	worklist, err := queries.HandleGetDoctorWorklist(c.Context(), q, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, worklist)
}

// GetVisitDetail godoc
// @Summary Get visit detail
// @Description Retrieve details of a specific visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Success 200 {object} response.Response
// @Failure 400,404 {object} response.Response
// @Router /visits/{id} [get]
// @Security BearerAuth
func (h *VisitHandler) GetVisitDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}

	detail, err := queries.HandleGetVisitDetail(c.Context(), queries.GetVisitDetailQuery{VisitID: id}, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "NOT_FOUND", Status: 404, Message: err.Error()})
	}
	return response.OK(c, detail)
}

type UpdateVisitStatusReq struct {
	Status domain.VisitStatus `json:"status"`
}

// UpdateVisitStatus godoc
// @Summary Update visit status
// @Description Change the status of a visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Param request body UpdateVisitStatusReq true "Status Update Payload"
// @Success 200 {object} response.Response
// @Failure 400,422 {object} response.Response
// @Router /visits/{id}/status [put]
// @Security BearerAuth
func (h *VisitHandler) UpdateVisitStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	var req UpdateVisitStatusReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request body"})
	}
	cmd := commands.UpdateVisitStatusCommand{VisitID: id, NewStatus: req.Status}
	if err := commands.HandleUpdateVisitStatus(c.Context(), cmd, h.repo); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "UNPROCESSABLE_ENTITY", Status: 422, Message: err.Error()})
	}
	return response.OK(c, fiber.Map{"message": "Status updated"})
}

type RecordVitalsReq struct {
	RecordedBy  uuid.UUID `json:"recorded_by"`
	BpSystolic  *int      `json:"bp_systolic,omitempty"`
	BpDiastolic *int      `json:"bp_diastolic,omitempty"`
	HeartRate   *int      `json:"heart_rate,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	SpO2        *int      `json:"spo2,omitempty"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	HeightCm    *float64  `json:"height_cm,omitempty"`
}

// RecordVitals godoc
// @Summary Record vitals
// @Description Record patient vitals for a visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Param request body RecordVitalsReq true "Vitals Payload"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits/{id}/vitals [post]
// @Security BearerAuth
func (h *VisitHandler) RecordVitals(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	var req RecordVitalsReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request body"})
	}
	claims := c.Locals("claims").(auth.Claims)

	cmd := commands.RecordVitalsCommand{
		VisitID:     id,
		RecordedBy:  claims.UserID,
		BpSystolic:  req.BpSystolic,
		BpDiastolic: req.BpDiastolic,
		HeartRate:   req.HeartRate,
		Temperature: req.Temperature,
		SpO2:        req.SpO2,
		WeightKg:    req.WeightKg,
		HeightCm:    req.HeightCm,
	}
	vital, err := commands.HandleRecordVitals(c.Context(), cmd, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, vital)
}

// GetVitals godoc
// @Summary Get vitals
// @Description Retrieve vitals recorded for a visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits/{id}/vitals [get]
// @Security BearerAuth
func (h *VisitHandler) GetVitals(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	vitals, err := h.repo.FindVitalsByVisitID(c.Context(), id)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, vitals)
}

type CreateOrderReq struct {
	OrderType domain.OrderType `json:"order_type"`
	Details   string           `json:"details"`
}

// CreateOrder godoc
// @Summary Create order
// @Description Create an order (e.g., lab test, prescription) for a visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Param request body CreateOrderReq true "Order Creation Payload"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits/{id}/orders [post]
// @Security BearerAuth
func (h *VisitHandler) CreateOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	var req CreateOrderReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request body"})
	}
	cmd := commands.CreateVisitOrderCommand{VisitID: id, OrderType: req.OrderType, Details: req.Details}
	order, err := commands.HandleCreateVisitOrder(c.Context(), cmd, h.repo, h.q)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, order)
}

// GetOrders godoc
// @Summary Get orders
// @Description Retrieve all orders created during a visit.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /visits/{id}/orders [get]
// @Security BearerAuth
func (h *VisitHandler) GetOrders(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	orders, err := h.repo.FindOrdersByVisitID(c.Context(), id)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, orders)
}

// CloseVisit godoc
// @Summary Close visit
// @Description Mark a visit as closed.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param id path string true "Visit ID"
// @Success 200 {object} response.Response
// @Failure 400,422 {object} response.Response
// @Router /visits/{id}/close [post]
// @Security BearerAuth
func (h *VisitHandler) CloseVisit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid visit ID"})
	}
	cmd := commands.CloseVisitCommand{VisitID: id}
	if err := commands.HandleCloseVisit(c.Context(), cmd, h.repo, h.q); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "UNPROCESSABLE_ENTITY", Status: 422, Message: err.Error()})
	}
	return response.OK(c, fiber.Map{"message": "Visit closed"})
}

// SearchICD10 godoc
// @Summary Search ICD-10 codes
// @Description Search for ICD-10 diagnosis codes by description or code.
// @Tags Visit (Doctor)
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Max results" default(20)
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /icd10 [get]
// @Security BearerAuth
func (h *VisitHandler) SearchICD10(c *fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Query parameter 'q' is required"})
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	results, err := queries.HandleSearchICD10(c.Context(), queries.SearchICD10Query{Query: q, Limit: limit}, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, results)
}
