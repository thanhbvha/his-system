package query

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"his-system/internal/patient/domain"
	"his-system/pkg/crypto"
)

type GetPatientByIDQuery struct {
	ID      uuid.UUID
	MaskPII bool // true if requested by PATIENT, false if requested by STAFF
}

type PatientDetail struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	DOB           string `json:"dob"`
	Gender        string `json:"gender"`
	BloodType     string `json:"blood_type"`
	Phone         string `json:"phone"` // Could be masked or plain
	CCCD          string `json:"cccd"`  // Could be masked or plain
	Email         string `json:"email"`
	AddressDetail string `json:"address_detail"`
	AvatarURL     string `json:"avatar_url"`
	IsActive      bool   `json:"is_active"`
	PatientCode   string `json:"patient_code"`
}

type GetPatientByIDHandler struct {
	repo   domain.PatientRepository
	cipher *crypto.FieldCipher
}

func NewGetPatientByIDHandler(repo domain.PatientRepository, cipher *crypto.FieldCipher) *GetPatientByIDHandler {
	return &GetPatientByIDHandler{repo: repo, cipher: cipher}
}

func maskCCCD(plain string) string {
	if len(plain) < 6 {
		return "***"
	}
	return plain[:3] + "******" + plain[len(plain)-3:]
}

func (h *GetPatientByIDHandler) Handle(ctx context.Context, q GetPatientByIDQuery) (*PatientDetail, error) {
	p, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("patient not found")
	}

	phonePlain := ""
	if p.PhoneEncrypted != "" {
		pt, _ := h.cipher.Decrypt(p.PhoneEncrypted)
		if q.MaskPII {
			phonePlain = MaskPhone(pt)
		} else {
			phonePlain = pt
		}
	}

	cccdPlain := ""
	if p.CCCDEncrypted != nil && *p.CCCDEncrypted != "" {
		pt, _ := h.cipher.Decrypt(*p.CCCDEncrypted)
		if q.MaskPII {
			cccdPlain = maskCCCD(pt)
		} else {
			cccdPlain = pt
		}
	}

	emailPlain := ""
	if p.EmailEncrypted != nil && *p.EmailEncrypted != "" {
		pt, _ := h.cipher.Decrypt(*p.EmailEncrypted)
		emailPlain = pt
	}

	addrPlain := ""
	if p.AddressDetailEncrypted != nil && *p.AddressDetailEncrypted != "" {
		pt, _ := h.cipher.Decrypt(*p.AddressDetailEncrypted)
		addrPlain = pt
	}

	dob := ""
	if p.DOB != nil {
		dob = p.DOB.Format("2006-01-02")
	}

	bloodType := ""
	if p.BloodType != nil {
		bloodType = *p.BloodType
	}

	avatarUrl := ""
	if p.AvatarURL != nil {
		avatarUrl = *p.AvatarURL
	}

	return &PatientDetail{
		ID:            p.ID.String(),
		FullName:      p.FullName,
		DOB:           dob,
		Gender:        p.Gender,
		BloodType:     bloodType,
		Phone:         phonePlain,
		CCCD:          cccdPlain,
		Email:         emailPlain,
		AddressDetail: addrPlain,
		AvatarURL:     avatarUrl,
		IsActive:      p.IsActive,
		PatientCode:   "BN-" + p.ID.String()[:8],
	}, nil
}
