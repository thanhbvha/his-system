package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/patient/domain"
	"his-system/pkg/crypto"
)

type UpdateInsuranceCommand struct {
	PatientID       uuid.UUID
	BHYTNumber      string // plaintext
	ValidFrom       *time.Time
	ValidTo         *time.Time
	CoverageLevel   *string
	IssuingProvince *string
}

type UpdateInsuranceHandler struct {
	repo   domain.PatientRepository
	cipher *crypto.FieldCipher
}

func NewUpdateInsuranceHandler(repo domain.PatientRepository, cipher *crypto.FieldCipher) *UpdateInsuranceHandler {
	return &UpdateInsuranceHandler{repo: repo, cipher: cipher}
}

func (h *UpdateInsuranceHandler) Handle(ctx context.Context, cmd UpdateInsuranceCommand) (*domain.PatientInsurance, error) {
	bhyt, err := domain.NewBHYTNumber(cmd.BHYTNumber, h.cipher)
	if err != nil {
		return nil, err
	}

	ins := &domain.PatientInsurance{
		ID:                  uuid.New(),
		PatientID:           cmd.PatientID,
		BHYTNumberEncrypted: bhyt.Encrypted(),
		BHYTNumberHMAC:      bhyt.HMAC(),
		ValidFrom:           cmd.ValidFrom,
		ValidTo:             cmd.ValidTo,
		CoverageLevel:       cmd.CoverageLevel,
		IssuingProvince:     cmd.IssuingProvince,
	}

	existing, err := h.repo.GetInsurance(ctx, cmd.PatientID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		ins.ID = existing.ID // Update existing
	}

	if err := h.repo.UpsertInsurance(ctx, ins); err != nil {
		return nil, err
	}

	return ins, nil
}
