package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VisitRepository interface {
	Save(ctx context.Context, v *Visit) error
	FindByID(ctx context.Context, id uuid.UUID) (*Visit, error)
	FindWorklist(ctx context.Context, doctorID uuid.UUID, date time.Time, status VisitStatus) ([]*VisitWithPatient, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status VisitStatus, v *Visit) error

	SaveVital(ctx context.Context, vital *VisitVital) error
	FindVitalsByVisitID(ctx context.Context, visitID uuid.UUID) ([]*VisitVital, error)

	SaveOrder(ctx context.Context, order *VisitOrder) error
	FindOrdersByVisitID(ctx context.Context, visitID uuid.UUID) ([]*VisitOrder, error)

	SearchICD10(ctx context.Context, query string, limit int) ([]*ICD10Code, error)
}

// VisitWithPatient is a projection used by the worklist query to avoid an extra round-trip.
type VisitWithPatient struct {
	Visit
	PatientFullName string `json:"patient_full_name"`
	PatientPhone    string `json:"patient_phone"`
}
