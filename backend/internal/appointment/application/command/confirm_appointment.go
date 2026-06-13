package command

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"his-system/internal/appointment/domain"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type ConfirmAppointmentCommand struct {
	ID uuid.UUID
}

type ConfirmAppointmentHandler struct {
	apptRepo domain.AppointmentRepository
	rdb      *commonRedis.Client
}

func NewConfirmAppointmentHandler(apptRepo domain.AppointmentRepository, rdb *commonRedis.Client) *ConfirmAppointmentHandler {
	return &ConfirmAppointmentHandler{apptRepo: apptRepo, rdb: rdb}
}

func (h *ConfirmAppointmentHandler) Handle(ctx context.Context, cmd ConfirmAppointmentCommand) error {
	appt, err := h.apptRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if appt == nil {
		return errors.New("appointment not found")
	}

	if err := appt.Confirm(); err != nil {
		return err
	}

	if err := h.apptRepo.Update(ctx, appt); err != nil {
		return err
	}

	return nil
}
