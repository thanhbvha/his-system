package command

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"his-system/internal/appointment/domain"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type CancelAppointmentCommand struct {
	ID           uuid.UUID
	CancelReason string
}

type CancelAppointmentHandler struct {
	apptRepo domain.AppointmentRepository
	slotRepo domain.SlotRepository
	rdb      *commonRedis.Client
}

func NewCancelAppointmentHandler(apptRepo domain.AppointmentRepository, slotRepo domain.SlotRepository, rdb *commonRedis.Client) *CancelAppointmentHandler {
	return &CancelAppointmentHandler{apptRepo: apptRepo, slotRepo: slotRepo, rdb: rdb}
}

func (h *CancelAppointmentHandler) Handle(ctx context.Context, cmd CancelAppointmentCommand) error {
	appt, err := h.apptRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if appt == nil {
		return errors.New("appointment not found")
	}

	if err := appt.Cancel(cmd.CancelReason); err != nil {
		return err
	}

	if err := h.apptRepo.Update(ctx, appt); err != nil {
		return err
	}

	if appt.SlotID != nil {
		_ = h.slotRepo.ReleaseSlot(ctx, *appt.SlotID)
	}

	return nil
}
