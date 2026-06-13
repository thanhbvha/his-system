package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"his-system/internal/patient/domain"
	"his-system/pkg/crypto"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type UpdatePatientCommand struct {
	ID            uuid.UUID
	FullName      string
	DOB           *time.Time
	Gender        string
	BloodType     *string
	Phone         string // plaintext
	CCCD          *string // plaintext
	Email         *string // plaintext
	AddressDetail *string // plaintext
	AvatarURL     *string
	IsActive      bool
}

type UpdatePatientHandler struct {
	repo   domain.PatientRepository
	cipher *crypto.FieldCipher
	rdb    *commonRedis.Client
}

func NewUpdatePatientHandler(repo domain.PatientRepository, cipher *crypto.FieldCipher, rdb *commonRedis.Client) *UpdatePatientHandler {
	return &UpdatePatientHandler{repo: repo, cipher: cipher, rdb: rdb}
}

func (h *UpdatePatientHandler) Handle(ctx context.Context, cmd UpdatePatientCommand) (*domain.Patient, error) {
	patient, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, errors.New("patient not found")
	}

	phone, err := domain.NewPhoneNumber(cmd.Phone, h.cipher)
	if err != nil {
		return nil, err
	}
	if phone.HMAC() != patient.PhoneHMAC {
		existing, err := h.repo.GetByPhoneHMAC(ctx, phone.HMAC())
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != patient.ID {
			return nil, domain.ErrPhoneExists
		}
		patient.PhoneEncrypted = phone.Encrypted()
		patient.PhoneHMAC = phone.HMAC()
	}

	if cmd.CCCD != nil && *cmd.CCCD != "" {
		cccd, err := domain.NewCCCD(*cmd.CCCD, h.cipher)
		if err != nil {
			return nil, err
		}
		hm := cccd.HMAC()
		if patient.CCCDHMAC == nil || hm != *patient.CCCDHMAC {
			existingCccd, err := h.repo.GetByCCCDHMAC(ctx, hm)
			if err != nil {
				return nil, err
			}
			if existingCccd != nil && existingCccd.ID != patient.ID {
				return nil, domain.ErrCCCDExists
			}
			e := cccd.Encrypted()
			patient.CCCDEncrypted = &e
			patient.CCCDHMAC = &hm
		}
	} else {
		patient.CCCDEncrypted = nil
		patient.CCCDHMAC = nil
	}

	if cmd.Email != nil && *cmd.Email != "" {
		emailHMAC := h.cipher.HMAC(*cmd.Email)
		if patient.EmailHMAC == nil || emailHMAC != *patient.EmailHMAC {
			emailEncStr, err := h.cipher.Encrypt(*cmd.Email)
			if err != nil {
				return nil, err
			}
			patient.EmailEncrypted = &emailEncStr
			patient.EmailHMAC = &emailHMAC
		}
	} else {
		patient.EmailEncrypted = nil
		patient.EmailHMAC = nil
	}

	if cmd.AddressDetail != nil && *cmd.AddressDetail != "" {
		enc, err := h.cipher.Encrypt(*cmd.AddressDetail)
		if err != nil {
			return nil, err
		}
		patient.AddressDetailEncrypted = &enc
	} else {
		patient.AddressDetailEncrypted = nil
	}

	patient.FullName = cmd.FullName
	patient.DOB = cmd.DOB
	patient.Gender = cmd.Gender
	patient.BloodType = cmd.BloodType
	patient.IsActive = cmd.IsActive
	patient.AvatarURL = cmd.AvatarURL
	patient.UpdatedAt = time.Now()

	if err := h.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	// TODO: Publish Redis Stream event HIS.PATIENT.PatientUpdated

	return patient, nil
}
