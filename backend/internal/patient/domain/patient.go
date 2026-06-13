package domain

import (
	"time"

	"github.com/google/uuid"
)

type Patient struct {
	ID                     uuid.UUID
	FullName               string
	DOB                    *time.Time
	Gender                 string // "MALE" | "FEMALE" | "OTHER"
	BloodType              *string
	IsActive               bool

	// PII — chứa dạng mã hóa, không bao giờ plaintext
	PhoneEncrypted         string
	PhoneHMAC              string
	CCCDEncrypted          *string
	CCCDHMAC               *string
	EmailEncrypted         *string
	EmailHMAC              *string
	AddressDetailEncrypted *string

	AvatarURL              *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type PatientInsurance struct {
	ID                  uuid.UUID
	PatientID           uuid.UUID
	BHYTNumberEncrypted string
	BHYTNumberHMAC      string
	ValidFrom           *time.Time
	ValidTo             *time.Time
	CoverageLevel       *string
	IssuingProvince     *string
}

type PatientContact struct {
	ID             uuid.UUID
	PatientID      uuid.UUID
	Name           string
	Relationship   string
	PhoneEncrypted string
	PhoneHMAC      string
	IsPrimary      bool
}
