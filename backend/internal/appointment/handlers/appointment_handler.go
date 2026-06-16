package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"his-system/internal/appointment/application/command"
	"his-system/internal/appointment/application/query"
	"his-system/internal/appointment/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

type AppointmentHandler struct {
	bookCmd      *command.BookAppointmentHandler
	cancelCmd    *command.CancelAppointmentHandler
	confirmCmd   *command.ConfirmAppointmentHandler
	slotsQuery   *query.GetAvailableSlotsHandler
	listQuery    *query.ListAppointmentsHandler
}

func NewAppointmentHandler(
	bookCmd *command.BookAppointmentHandler,
	cancelCmd *command.CancelAppointmentHandler,
	confirmCmd *command.ConfirmAppointmentHandler,
	slotsQuery *query.GetAvailableSlotsHandler,
	listQuery *query.ListAppointmentsHandler,
) *AppointmentHandler {
	return &AppointmentHandler{
		bookCmd:      bookCmd,
		cancelCmd:    cancelCmd,
		confirmCmd:   confirmCmd,
		slotsQuery:   slotsQuery,
		listQuery:    listQuery,
	}
}

func handleError(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrSlotAlreadyBooked:
		e := *appErrors.ErrConflict
		e.Message = err.Error()
		return response.Fail(c, &e)
	case domain.ErrCancelTooLate, domain.ErrInvalidStatus:
		e := *appErrors.ErrValidation
		e.Message = err.Error()
		return response.Fail(c, &e)
	}

	if err.Error() == "appointment not found" {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.Fail(c, appErrors.ErrInternal)
}

func (h *AppointmentHandler) GetAvailableSlots(c *fiber.Ctx) error {
	docStr := c.Query("doctor_id")
	dateStr := c.Query("date")

	if docStr == "" || dateStr == "" {
		return response.Fail(c, appErrors.ErrValidation)
	}

	docID, err := uuid.Parse(docStr)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.slotsQuery.Handle(c.Context(), query.GetAvailableSlotsQuery{
		DoctorID: docID,
		Date:     date,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, res)
}

func (h *AppointmentHandler) List(c *fiber.Ctx) error {
	var date *time.Time
	if d := c.Query("date"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			date = &t
		}
	}

	var docID *uuid.UUID
	if d := c.Query("doctor_id"); d != "" {
		id, err := uuid.Parse(d)
		if err == nil {
			docID = &id
		}
	}

	var patientID *uuid.UUID
	if p := c.Query("patient_id"); p != "" {
		id, err := uuid.Parse(p)
		if err == nil {
			patientID = &id
		}
	}

	claims := c.Locals("claims").(auth.Claims)
	isStaff := false
	for _, r := range claims.Roles {
		if r == "receptionist" || r == "doctor" || r == "admin" || r == "nurse" {
			isStaff = true
			break
		}
	}
	if !isStaff {
		if patientID != nil && *patientID != claims.UserID {
			return response.Fail(c, appErrors.ErrForbidden)
		}
		pid := claims.UserID
		patientID = &pid
	}

	var status *domain.AppointmentStatus
	if s := c.Query("status"); s != "" {
		st := domain.AppointmentStatus(s)
		status = &st
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	res, err := h.listQuery.Handle(c.Context(), query.ListAppointmentsQuery{
		Date:      date,
		DoctorID:  docID,
		PatientID: patientID,
		Status:    status,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, res)
}

type BookReq struct {
	DoctorID    string  `json:"doctor_id"`
	ServiceID   *string `json:"service_id"`
	SlotID      *string `json:"slot_id"`
	ScheduledAt string  `json:"scheduled_at"` // RFC3339
	Note        *string `json:"note"`
	PatientID   *string `json:"patient_id"`   // Used by staff
}

func (h *AppointmentHandler) Book(c *fiber.Ctx) error {
	claims := c.Locals("claims").(auth.Claims)

	var req BookReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	docID, err := uuid.Parse(req.DoctorID)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var slotID *uuid.UUID
	if req.SlotID != nil {
		id, err := uuid.Parse(*req.SlotID)
		if err == nil {
			slotID = &id
		}
	}

	var srvID *uuid.UUID
	if req.ServiceID != nil {
		id, err := uuid.Parse(*req.ServiceID)
		if err == nil {
			srvID = &id
		}
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	isStaff := false
	for _, r := range claims.Roles {
		if r == "receptionist" || r == "admin" {
			isStaff = true
			break
		}
	}

	var pid uuid.UUID
	var bookedBy *uuid.UUID
	if isStaff {
		bookedBy = &claims.UserID
		if req.PatientID == nil {
			return response.Fail(c, appErrors.ErrValidation)
		}
		p, err := uuid.Parse(*req.PatientID)
		if err != nil {
			return response.Fail(c, appErrors.ErrValidation)
		}
		pid = p
	} else {
		pid = claims.UserID
	}

	res, err := h.bookCmd.Handle(c.Context(), command.BookAppointmentCommand{
		PatientID:   pid,
		DoctorID:    docID,
		ServiceID:   srvID,
		SlotID:      slotID,
		ScheduledAt: scheduledAt,
		Note:        req.Note,
		BookedBy:    bookedBy,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"id": res.ID})
}

type CancelReq struct {
	Reason string `json:"reason"`
}

func (h *AppointmentHandler) Cancel(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req CancelReq
	_ = c.BodyParser(&req)

	err = h.cancelCmd.Handle(c.Context(), command.CancelAppointmentCommand{
		ID:           id,
		CancelReason: req.Reason,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"success": true})
}

func (h *AppointmentHandler) Confirm(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	err = h.confirmCmd.Handle(c.Context(), command.ConfirmAppointmentCommand{ID: id})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"success": true})
}
