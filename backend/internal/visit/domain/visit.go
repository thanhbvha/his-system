package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type VisitStatus string

const (
	VisitRegistered VisitStatus = "REGISTERED"
	VisitWaiting    VisitStatus = "WAITING"
	VisitInProgress VisitStatus = "IN_PROGRESS"
	VisitOrdered    VisitStatus = "ORDERED"
	VisitCompleted  VisitStatus = "COMPLETED"
	VisitCancelled  VisitStatus = "CANCELLED"
)

type Visit struct {
	ID             uuid.UUID
	PatientID      uuid.UUID
	DoctorID       uuid.UUID
	QueueEntryID   *uuid.UUID
	Status         VisitStatus
	ChiefComplaint *string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (v *Visit) Start() error {
	if v.Status != VisitRegistered && v.Status != VisitWaiting {
		return errors.New("visit can only be started from REGISTERED or WAITING status")
	}
	v.Status = VisitInProgress
	now := time.Now()
	v.StartedAt = &now
	v.UpdatedAt = now
	return nil
}

func (v *Visit) Complete() error {
	if v.Status != VisitInProgress && v.Status != VisitOrdered {
		return errors.New("visit can only be completed from IN_PROGRESS or ORDERED status")
	}
	v.Status = VisitCompleted
	now := time.Now()
	v.CompletedAt = &now
	v.UpdatedAt = now
	return nil
}

func (v *Visit) Cancel() error {
	if v.Status == VisitCompleted {
		return errors.New("cannot cancel a completed visit")
	}
	v.Status = VisitCancelled
	v.UpdatedAt = time.Now()
	return nil
}

func (v *Visit) SetOrdered() error {
	if v.Status != VisitInProgress {
		return errors.New("visit must be IN_PROGRESS before ordering")
	}
	v.Status = VisitOrdered
	v.UpdatedAt = time.Now()
	return nil
}
