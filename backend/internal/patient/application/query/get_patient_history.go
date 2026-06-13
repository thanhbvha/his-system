package query

import (
	"context"

	"github.com/google/uuid"
)

type GetPatientHistoryQuery struct {
	PatientID uuid.UUID
}

type VisitSummary struct {
    AppointmentID   string `json:"appointment_id"`
    Date            string `json:"date"`
    DoctorName      string `json:"doctor_name"`
    DepartmentName  string `json:"department_name"`
    Status          string `json:"status"`
}

type GetPatientHistoryHandler struct {
    // This handler will query appointments for a patient.
    // In Sprint 3, we can wire this up to an Appointment repository or service.
    // For now, returning an empty list.
}

func NewGetPatientHistoryHandler() *GetPatientHistoryHandler {
	return &GetPatientHistoryHandler{}
}

func (h *GetPatientHistoryHandler) Handle(ctx context.Context, q GetPatientHistoryQuery) ([]*VisitSummary, error) {
	// Implementation will be provided in Step 5 (Appointment module)
    // or we can just fetch from an appointment repository interface injected here.
	return make([]*VisitSummary, 0), nil
}
