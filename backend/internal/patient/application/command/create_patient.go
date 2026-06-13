package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/patient/domain"
	"his-system/pkg/crypto"
	commonRedis "github.com/thanhbvha/go-common/redis"
)

type CreatePatientCommand struct {
	FullName      string
	DOB           *time.Time
	Gender        string
	BloodType     *string
	Phone         string // plaintext
	CCCD          *string // plaintext
	Email         *string // plaintext
	AddressDetail *string // plaintext
	AvatarURL     *string
}

type CreatePatientHandler struct {
	repo   domain.PatientRepository
	cipher *crypto.FieldCipher
	rdb    *commonRedis.Client
}

func NewCreatePatientHandler(repo domain.PatientRepository, cipher *crypto.FieldCipher, rdb *commonRedis.Client) *CreatePatientHandler {
	return &CreatePatientHandler{repo: repo, cipher: cipher, rdb: rdb}
}

func (h *CreatePatientHandler) Handle(ctx context.Context, cmd CreatePatientCommand) (*domain.Patient, error) {
	// 1. Phone
	phone, err := domain.NewPhoneNumber(cmd.Phone, h.cipher)
	if err != nil {
		return nil, err
	}

	// 2. Check unique Phone
	existing, err := h.repo.GetByPhoneHMAC(ctx, phone.HMAC())
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrPhoneExists
	}

	// 3. CCCD (optional)
	var cccdEnc, cccdHMAC *string
	if cmd.CCCD != nil && *cmd.CCCD != "" {
		cccd, err := domain.NewCCCD(*cmd.CCCD, h.cipher)
		if err != nil {
			return nil, err
		}
		e := cccd.Encrypted()
		hm := cccd.HMAC()
		cccdEnc = &e
		cccdHMAC = &hm

		// Check unique CCCD
		existingCccd, err := h.repo.GetByCCCDHMAC(ctx, hm)
		if err != nil {
			return nil, err
		}
		if existingCccd != nil {
			return nil, domain.ErrCCCDExists
		}
	}

	// 4. Email (optional)
	var emailEnc, emailHMAC *string
	if cmd.Email != nil && *cmd.Email != "" {
		emailEncStr, err := h.cipher.Encrypt(*cmd.Email)
		if err != nil {
			return nil, err
		}
		hmacStr := h.cipher.HMAC(*cmd.Email)
		emailEnc = &emailEncStr
		emailHMAC = &hmacStr
	}

	// 5. Address (optional)
	var addrEnc *string
	if cmd.AddressDetail != nil && *cmd.AddressDetail != "" {
		enc, err := h.cipher.Encrypt(*cmd.AddressDetail)
		if err != nil {
			return nil, err
		}
		addrEnc = &enc
	}

	patient := &domain.Patient{
		ID:                     uuid.New(),
		FullName:               cmd.FullName,
		DOB:                    cmd.DOB,
		Gender:                 cmd.Gender,
		BloodType:              cmd.BloodType,
		IsActive:               true,
		PhoneEncrypted:         phone.Encrypted(),
		PhoneHMAC:              phone.HMAC(),
		CCCDEncrypted:          cccdEnc,
		CCCDHMAC:               cccdHMAC,
		EmailEncrypted:         emailEnc,
		EmailHMAC:              emailHMAC,
		AddressDetailEncrypted: addrEnc,
		AvatarURL:              cmd.AvatarURL,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := h.repo.Create(ctx, patient); err != nil {
		return nil, err
	}

	// TODO: Publish Redis Stream event HIS.PATIENT.PatientRegistered

	return patient, nil
}
