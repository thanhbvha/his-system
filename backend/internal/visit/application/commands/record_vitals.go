package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/visit/domain"
)

type RecordVitalsCommand struct {
	VisitID     uuid.UUID
	RecordedBy  uuid.UUID
	BpSystolic  *int
	BpDiastolic *int
	HeartRate   *int
	Temperature *float64
	SpO2        *int
	WeightKg    *float64
	HeightCm    *float64
}

func HandleRecordVitals(ctx context.Context, cmd RecordVitalsCommand, repo domain.VisitRepository) (*domain.VisitVital, error) {
	vital := &domain.VisitVital{
		VisitID:     cmd.VisitID,
		RecordedBy:  cmd.RecordedBy,
		BpSystolic:  cmd.BpSystolic,
		BpDiastolic: cmd.BpDiastolic,
		HeartRate:   cmd.HeartRate,
		Temperature: cmd.Temperature,
		SpO2:        cmd.SpO2,
		WeightKg:    cmd.WeightKg,
		HeightCm:    cmd.HeightCm,
		RecordedAt:  time.Now(),
	}

	if err := repo.SaveVital(ctx, vital); err != nil {
		return nil, err
	}

	return vital, nil
}
