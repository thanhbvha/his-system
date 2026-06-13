package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/appointment/domain"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type BookAppointmentCommand struct {
	PatientID   uuid.UUID
	DoctorID    uuid.UUID
	ServiceID   *uuid.UUID
	SlotID      *uuid.UUID
	ScheduledAt time.Time
	Note        *string
	BookedBy    *uuid.UUID
}

type BookAppointmentHandler struct {
	apptRepo domain.AppointmentRepository
	slotRepo domain.SlotRepository
	rdb      *commonRedis.Client
}

func NewBookAppointmentHandler(apptRepo domain.AppointmentRepository, slotRepo domain.SlotRepository, rdb *commonRedis.Client) *BookAppointmentHandler {
	return &BookAppointmentHandler{apptRepo: apptRepo, slotRepo: slotRepo, rdb: rdb}
}

func (h *BookAppointmentHandler) Handle(ctx context.Context, cmd BookAppointmentCommand) (*domain.Appointment, error) {
	if cmd.SlotID != nil {
		err := h.slotRepo.BookSlot(ctx, *cmd.SlotID)
		if err != nil {
			return nil, err
		}
	}

	appt := &domain.Appointment{
		ID:          uuid.New(),
		PatientID:   cmd.PatientID,
		DoctorID:    cmd.DoctorID,
		ServiceID:   cmd.ServiceID,
		SlotID:      cmd.SlotID,
		ScheduledAt: cmd.ScheduledAt,
		Status:      domain.StatusPending,
		Note:        cmd.Note,
		BookedBy:    cmd.BookedBy,
		BookedAt:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := h.apptRepo.Create(ctx, appt)
	if err != nil {
		if cmd.SlotID != nil {
			_ = h.slotRepo.ReleaseSlot(context.Background(), *cmd.SlotID)
		}
		return nil, err
	}

	return appt, nil
}
