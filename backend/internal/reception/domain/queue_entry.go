package domain

import (
	"time"

	"github.com/google/uuid"
)

type QueueStatus string

const (
	StatusWaiting    QueueStatus = "WAITING"
	StatusCalled     QueueStatus = "CALLED"
	StatusInProgress QueueStatus = "IN_PROGRESS"
	StatusDone       QueueStatus = "DONE"
	StatusSkipped    QueueStatus = "SKIPPED"
)

type QueueEntry struct {
	ID            uuid.UUID
	PatientID     uuid.UUID
	VisitID       *uuid.UUID
	AppointmentID *uuid.UUID
	ServiceType   string
	QueueNumber   string
	Status        QueueStatus
	CalledAt      *time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time
}

func (q *QueueEntry) Call() error {
	q.Status = StatusCalled
	now := time.Now()
	q.CalledAt = &now
	return nil
}

func (q *QueueEntry) Skip() error {
	q.Status = StatusSkipped
	return nil
}

func (q *QueueEntry) Complete() error {
	q.Status = StatusDone
	now := time.Now()
	q.CompletedAt = &now
	return nil
}
