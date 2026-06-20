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

type QueuePatient struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"full_name"`
	PatientCode string    `json:"patient_code"`
}


type QueueEntry struct {
	ID            uuid.UUID     `json:"id"`
	PatientID     uuid.UUID     `json:"patient_id"`
	Patient       *QueuePatient `json:"patient,omitempty"`
	VisitID       *uuid.UUID  `json:"visit_id,omitempty"`
	AppointmentID *uuid.UUID  `json:"appointment_id,omitempty"`
	ServiceType   string      `json:"service_type"`
	QueueNumber   string      `json:"queue_number"`
	Status        QueueStatus `json:"status"`
	CalledAt      *time.Time  `json:"called_at,omitempty"`
	CompletedAt   *time.Time  `json:"completed_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
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
