package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "PENDING"
	StatusConfirmed AppointmentStatus = "CONFIRMED"
	StatusCheckedIn AppointmentStatus = "CHECKED_IN"
	StatusCompleted AppointmentStatus = "COMPLETED"
	StatusCancelled AppointmentStatus = "CANCELLED"
)

var (
	ErrSlotAlreadyBooked = errors.New("slot này vừa được đặt bởi người khác")
	ErrCancelTooLate     = errors.New("không thể hủy lịch trong vòng 24h")
	ErrInvalidStatus     = errors.New("trạng thái không hợp lệ để thực hiện thao tác")
)

type Appointment struct {
	ID           uuid.UUID
	PatientID    uuid.UUID
	DoctorID     uuid.UUID
	ServiceID    *uuid.UUID
	SlotID       *uuid.UUID
	ScheduledAt  time.Time
	Status       AppointmentStatus
	Note         *string
	CancelReason *string
	BookedBy     *uuid.UUID
	BookedAt     time.Time
	ConfirmedAt  *time.Time
	CheckedInAt  *time.Time
	CompletedAt  *time.Time
	CancelledAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (a *Appointment) Cancel(reason string) error {
	if a.Status != StatusPending && a.Status != StatusConfirmed {
		return ErrInvalidStatus
	}
	if time.Until(a.ScheduledAt) < 24*time.Hour {
		return ErrCancelTooLate
	}
	a.Status = StatusCancelled
    if reason != "" {
        a.CancelReason = &reason
    }
    now := time.Now()
    a.CancelledAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *Appointment) Confirm() error {
	if a.Status != StatusPending {
		return ErrInvalidStatus
	}
	a.Status = StatusConfirmed
    now := time.Now()
    a.ConfirmedAt = &now
	a.UpdatedAt = now
	return nil
}

type Slot struct {
	ID         uuid.UUID
	DoctorID   uuid.UUID
	ScheduleID *uuid.UUID
	Date       time.Time
	StartTime  time.Time
	EndTime    time.Time
	IsBooked   bool
	CreatedAt  time.Time
}
