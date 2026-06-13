package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/appointment/domain"
)

type ListAppointmentsQuery struct {
	Date      *time.Time
	DoctorID  *uuid.UUID
	PatientID *uuid.UUID
	Status    *domain.AppointmentStatus
	Page      int
	Limit     int
}

type AppointmentListItem struct {
	ID          string `json:"id"`
	PatientID   string `json:"patient_id"`
	DoctorID    string `json:"doctor_id"`
	ScheduledAt string `json:"scheduled_at"`
	Status      string `json:"status"`
}

type ListAppointmentsResult struct {
	Items []*AppointmentListItem `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

type ListAppointmentsHandler struct {
	apptRepo domain.AppointmentRepository
}

func NewListAppointmentsHandler(apptRepo domain.AppointmentRepository) *ListAppointmentsHandler {
	return &ListAppointmentsHandler{apptRepo: apptRepo}
}

func (h *ListAppointmentsHandler) Handle(ctx context.Context, q ListAppointmentsQuery) (*ListAppointmentsResult, error) {
	appts, total, err := h.apptRepo.List(ctx, domain.ListFilter{
		Date:      q.Date,
		DoctorID:  q.DoctorID,
		PatientID: q.PatientID,
		Status:    q.Status,
		Page:      q.Page,
		Limit:     q.Limit,
	})
	if err != nil {
		return nil, err
	}

	var items []*AppointmentListItem
	for _, a := range appts {
		items = append(items, &AppointmentListItem{
			ID:          a.ID.String(),
			PatientID:   a.PatientID.String(),
			DoctorID:    a.DoctorID.String(),
			ScheduledAt: a.ScheduledAt.Format(time.RFC3339),
			Status:      string(a.Status),
		})
	}
	if items == nil {
		items = make([]*AppointmentListItem, 0)
	}

	return &ListAppointmentsResult{
		Items: items,
		Total: total,
		Page:  q.Page,
		Limit: q.Limit,
	}, nil
}
