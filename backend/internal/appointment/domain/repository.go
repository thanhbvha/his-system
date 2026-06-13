package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	Date      *time.Time
	DoctorID  *uuid.UUID
	PatientID *uuid.UUID
	Status    *AppointmentStatus
	Page      int
	Limit     int
}

type AppointmentRepository interface {
	Create(ctx context.Context, appt *Appointment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Appointment, error)
	List(ctx context.Context, filter ListFilter) ([]*Appointment, int64, error)
	Update(ctx context.Context, appt *Appointment) error
}

type SlotRepository interface {
	GetAvailable(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*Slot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Slot, error)
	BulkCreate(ctx context.Context, slots []*Slot) error
	
	// Anti-Double-Booking
	BookSlot(ctx context.Context, slotID uuid.UUID) error
	ReleaseSlot(ctx context.Context, slotID uuid.UUID) error
}
