package query

import (
	"context"

	"his-system/internal/patient/domain"
	"his-system/pkg/crypto"
)

type SearchPatientsQuery struct {
	Phone string // plaintext query
	CCCD  string // plaintext query
	Name  string // plaintext query
	Page  int
	Limit int
}

type PatientListItem struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	DOB         string `json:"dob"`
	Gender      string `json:"gender"`
	PhoneMasked string `json:"phone_masked"`
	PatientCode string `json:"patient_code"`
}

type SearchPatientsResult struct {
	Items []*PatientListItem `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

type SearchPatientsHandler struct {
	repo   domain.PatientRepository
	cipher *crypto.FieldCipher
}

func NewSearchPatientsHandler(repo domain.PatientRepository, cipher *crypto.FieldCipher) *SearchPatientsHandler {
	return &SearchPatientsHandler{repo: repo, cipher: cipher}
}

func MaskPhone(plain string) string {
	if len(plain) < 7 {
		return "***"
	}
	return plain[:3] + "***" + plain[len(plain)-3:]
}

func (h *SearchPatientsHandler) Handle(ctx context.Context, q SearchPatientsQuery) (*SearchPatientsResult, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}

	var patients []*domain.Patient
	var total int64
	var err error

	if q.Phone != "" {
		phoneHMAC := h.cipher.HMAC(q.Phone)
		p, err := h.repo.GetByPhoneHMAC(ctx, phoneHMAC)
		if err != nil {
			return nil, err
		}
		if p != nil {
			patients = append(patients, p)
			total = 1
		}
	} else if q.CCCD != "" {
		cccdHMAC := h.cipher.HMAC(q.CCCD)
		p, err := h.repo.GetByCCCDHMAC(ctx, cccdHMAC)
		if err != nil {
			return nil, err
		}
		if p != nil {
			patients = append(patients, p)
			total = 1
		}
	} else if q.Name != "" {
		patients, total, err = h.repo.SearchByName(ctx, q.Name, q.Page, q.Limit)
		if err != nil {
			return nil, err
		}
	} else {
		patients, total, err = h.repo.List(ctx, q.Page, q.Limit)
		if err != nil {
			return nil, err
		}
	}

	var items []*PatientListItem
	for _, p := range patients {
		phonePlain := ""
		if p.PhoneEncrypted != "" {
			pt, _ := h.cipher.Decrypt(p.PhoneEncrypted)
			phonePlain = MaskPhone(pt)
		}

		dob := ""
		if p.DOB != nil {
			dob = p.DOB.Format("2006-01-02")
		}

		items = append(items, &PatientListItem{
			ID:          p.ID.String(),
			FullName:    p.FullName,
			DOB:         dob,
			Gender:      p.Gender,
			PhoneMasked: phonePlain,
			PatientCode: "BN-" + p.ID.String()[:8],
		})
	}

	if items == nil {
		items = make([]*PatientListItem, 0)
	}

	return &SearchPatientsResult{
		Items: items,
		Total: total,
		Page:  q.Page,
		Limit: q.Limit,
	}, nil
}
