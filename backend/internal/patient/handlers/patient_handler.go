package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"his-system/internal/patient/application/command"
	"his-system/internal/patient/application/query"
	"his-system/internal/patient/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

func handleError(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrPhoneExists, domain.ErrCCCDExists:
		e := *appErrors.ErrConflict
		e.Message = err.Error()
		return response.Fail(c, &e)
	case domain.ErrInvalidPhone, domain.ErrInvalidCCCD, domain.ErrInvalidBHYT:
		e := *appErrors.ErrValidation
		e.Message = err.Error()
		return response.Fail(c, &e)
	}

	if err.Error() == "patient not found" {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.Fail(c, appErrors.ErrInternal)
}

type PatientHandler struct {
	createCmd       *command.CreatePatientHandler
	updateCmd       *command.UpdatePatientHandler
	updateInsCmd    *command.UpdateInsuranceHandler
	searchQuery     *query.SearchPatientsHandler
	getByIDQuery    *query.GetPatientByIDHandler
	getHistoryQuery *query.GetPatientHistoryHandler
}

func NewPatientHandler(
	createCmd *command.CreatePatientHandler,
	updateCmd *command.UpdatePatientHandler,
	updateInsCmd *command.UpdateInsuranceHandler,
	searchQuery *query.SearchPatientsHandler,
	getByIDQuery *query.GetPatientByIDHandler,
	getHistoryQuery *query.GetPatientHistoryHandler,
) *PatientHandler {
	return &PatientHandler{
		createCmd:       createCmd,
		updateCmd:       updateCmd,
		updateInsCmd:    updateInsCmd,
		searchQuery:     searchQuery,
		getByIDQuery:    getByIDQuery,
		getHistoryQuery: getHistoryQuery,
	}
}

// List godoc
// @Summary List patients
// @Description Search and list patients by name, phone, or CCCD.
// @Tags Patient (Staff)
// @Accept json
// @Produce json
// @Param q query string false "Search by name"
// @Param phone query string false "Search by phone"
// @Param cccd query string false "Search by CCCD"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response
// @Failure 401,403 {object} response.Response
// @Router /patients [get]
// @Security BearerAuth
func (h *PatientHandler) List(c *fiber.Ctx) error {
	q := c.Query("q")
	phone := c.Query("phone")
	cccd := c.Query("cccd")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	result, err := h.searchQuery.Handle(c.Context(), query.SearchPatientsQuery{
		Phone: phone,
		CCCD:  cccd,
		Name:  q,
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, result)
}

type CreatePatientReq struct {
	FullName      string  `json:"full_name"`
	DOB           string  `json:"dob"` // YYYY-MM-DD
	Gender        string  `json:"gender"`
	Phone         string  `json:"phone"`
	CCCD          *string `json:"cccd"`
	Email         *string `json:"email"`
	AddressDetail *string `json:"address"`
}

// Create godoc
// @Summary Create patient
// @Description Create a new patient record.
// @Tags Patient (Staff)
// @Accept json
// @Produce json
// @Param request body CreatePatientReq true "Patient Creation Payload"
// @Success 200 {object} response.Response
// @Failure 400,401,403,409 {object} response.Response
// @Router /patients [post]
// @Security BearerAuth
func (h *PatientHandler) Create(c *fiber.Ctx) error {
	var req CreatePatientReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var dob *time.Time
	if req.DOB != "" {
		t, err := time.Parse("2006-01-02", req.DOB)
		if err == nil {
			dob = &t
		}
	}

	p, err := h.createCmd.Handle(c.Context(), command.CreatePatientCommand{
		FullName:      req.FullName,
		DOB:           dob,
		Gender:        req.Gender,
		Phone:         req.Phone,
		CCCD:          req.CCCD,
		Email:         req.Email,
		AddressDetail: req.AddressDetail,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{
		"id":           p.ID,
		"full_name":    p.FullName,
		"patient_code": "BN-" + p.ID.String()[:8],
	})
}

// GetByID godoc
// @Summary Get patient by ID
// @Description Retrieve detailed patient information by their ID.
// @Tags Patient (Staff)
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Success 200 {object} response.Response
// @Failure 401,403,404 {object} response.Response
// @Router /patients/{id} [get]
// @Security BearerAuth
func (h *PatientHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return handleError(c, err)
	}

	result, err := h.getByIDQuery.Handle(c.Context(), query.GetPatientByIDQuery{
		ID:      id,
		MaskPII: false, // Staff requested
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, result)
}

type UpdatePatientReq struct {
	FullName      string  `json:"full_name"`
	DOB           string  `json:"dob"`
	Gender        string  `json:"gender"`
	BloodType     *string `json:"blood_type"`
	Phone         string  `json:"phone"`
	CCCD          *string `json:"cccd"`
	Email         *string `json:"email"`
	AddressDetail *string `json:"address"`
	AvatarURL     *string `json:"avatar_url"`
	IsActive      bool    `json:"is_active"`
}

// Update godoc
// @Summary Update patient
// @Description Update existing patient information.
// @Tags Patient (Staff)
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Param request body UpdatePatientReq true "Patient Update Payload"
// @Success 200 {object} response.Response
// @Failure 400,401,403,404 {object} response.Response
// @Router /patients/{id} [put]
// @Security BearerAuth
func (h *PatientHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req UpdatePatientReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var dob *time.Time
	if req.DOB != "" {
		t, err := time.Parse("2006-01-02", req.DOB)
		if err == nil {
			dob = &t
		}
	}

	p, err := h.updateCmd.Handle(c.Context(), command.UpdatePatientCommand{
		ID:            id,
		FullName:      req.FullName,
		DOB:           dob,
		Gender:        req.Gender,
		BloodType:     req.BloodType,
		Phone:         req.Phone,
		CCCD:          req.CCCD,
		Email:         req.Email,
		AddressDetail: req.AddressDetail,
		AvatarURL:     req.AvatarURL,
		IsActive:      req.IsActive,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"id": p.ID})
}

// GetMyProfile godoc
// @Summary Get my profile
// @Description Retrieve the logged-in patient's profile.
// @Tags Patient (Portal)
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401,404 {object} response.Response
// @Router /patients/me [get]
// @Security BearerAuth
func (h *PatientHandler) GetMyProfile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(auth.Claims)

	result, err := h.getByIDQuery.Handle(c.Context(), query.GetPatientByIDQuery{
		ID:      claims.UserID,
		MaskPII: true, // Patient requested -> Mask
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, result)
}

// UpdateMyProfile godoc
// @Summary Update my profile
// @Description Update the logged-in patient's profile information.
// @Tags Patient (Portal)
// @Accept json
// @Produce json
// @Param request body UpdatePatientReq true "Profile Update Payload"
// @Success 200 {object} response.Response
// @Failure 400,401 {object} response.Response
// @Router /patients/me [put]
// @Security BearerAuth
func (h *PatientHandler) UpdateMyProfile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(auth.Claims)

	var req UpdatePatientReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var dob *time.Time
	if req.DOB != "" {
		t, err := time.Parse("2006-01-02", req.DOB)
		if err == nil {
			dob = &t
		}
	}

	p, err := h.updateCmd.Handle(c.Context(), command.UpdatePatientCommand{
		ID:            claims.UserID,
		FullName:      req.FullName,
		DOB:           dob,
		Gender:        req.Gender,
		BloodType:     req.BloodType,
		Phone:         req.Phone,
		CCCD:          req.CCCD,
		Email:         req.Email,
		AddressDetail: req.AddressDetail,
		AvatarURL:     req.AvatarURL,
		IsActive:      true,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"id": p.ID})
}

type UpdateInsuranceReq struct {
	BHYTNumber      string  `json:"bhyt_number"`
	ValidFrom       string  `json:"valid_from"`
	ValidTo         string  `json:"valid_to"`
	CoverageLevel   *string `json:"coverage_level"`
	IssuingProvince *string `json:"issuing_province"`
}

// UpdateMyInsurance godoc
// @Summary Update my insurance
// @Description Update the logged-in patient's health insurance details.
// @Tags Patient (Portal)
// @Accept json
// @Produce json
// @Param request body UpdateInsuranceReq true "Insurance Update Payload"
// @Success 200 {object} response.Response
// @Failure 400,401 {object} response.Response
// @Router /patients/me/insurance [put]
// @Security BearerAuth
func (h *PatientHandler) UpdateMyInsurance(c *fiber.Ctx) error {
	claims := c.Locals("claims").(auth.Claims)

	var req UpdateInsuranceReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var validFrom, validTo *time.Time
	if req.ValidFrom != "" {
		t, err := time.Parse("2006-01-02", req.ValidFrom)
		if err == nil {
			validFrom = &t
		}
	}
	if req.ValidTo != "" {
		t, err := time.Parse("2006-01-02", req.ValidTo)
		if err == nil {
			validTo = &t
		}
	}

	ins, err := h.updateInsCmd.Handle(c.Context(), command.UpdateInsuranceCommand{
		PatientID:       claims.UserID,
		BHYTNumber:      req.BHYTNumber,
		ValidFrom:       validFrom,
		ValidTo:         validTo,
		CoverageLevel:   req.CoverageLevel,
		IssuingProvince: req.IssuingProvince,
	})
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, fiber.Map{"id": ins.ID})
}
