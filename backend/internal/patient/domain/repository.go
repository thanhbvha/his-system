package domain

import (
	"context"

	"github.com/google/uuid"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *Patient) error
	GetByID(ctx context.Context, id uuid.UUID) (*Patient, error)
	GetByPhoneHMAC(ctx context.Context, phoneHMAC string) (*Patient, error)
	GetByCCCDHMAC(ctx context.Context, cccdHMAC string) (*Patient, error)
	SearchByName(ctx context.Context, query string, page, limit int) ([]*Patient, int64, error)
	Update(ctx context.Context, patient *Patient) error
	List(ctx context.Context, page, limit int) ([]*Patient, int64, error)

	// Insurance
	UpsertInsurance(ctx context.Context, ins *PatientInsurance) error
	GetInsurance(ctx context.Context, patientID uuid.UUID) (*PatientInsurance, error)
}
