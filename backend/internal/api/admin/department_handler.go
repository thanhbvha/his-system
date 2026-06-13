package admin

import (
	"github.com/gofiber/fiber/v2"
	"his-system/internal/identity/application/command"
	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

type AdminDepartmentHandler struct {
	deptRepo      domain.DepartmentRepository
	createDeptCmd command.CreateDepartmentHandler
}

func NewAdminDepartmentHandler(
	deptRepo domain.DepartmentRepository,
	createDeptCmd *command.CreateDepartmentHandler,
) *AdminDepartmentHandler {
	return &AdminDepartmentHandler{
		deptRepo:      deptRepo,
		createDeptCmd: *createDeptCmd,
	}
}

func (h *AdminDepartmentHandler) List(c *fiber.Ctx) error {
	includeInactive := c.Query("include_inactive") == "true"
	depts, err := h.deptRepo.List(c.Context(), includeInactive)
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, depts)
}

type createDeptReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *AdminDepartmentHandler) Create(c *fiber.Ctx) error {
	var req createDeptReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.createDeptCmd.Handle(c.Context(), command.CreateDepartmentCommand{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, res)
}
