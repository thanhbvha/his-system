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

// List godoc
// @Summary List departments
// @Description Retrieve a list of departments.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param include_inactive query boolean false "Include inactive departments" default(false)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/departments [get]
// @Security BearerAuth
func (h *AdminDepartmentHandler) List(c *fiber.Ctx) error {
	includeInactive := c.Query("include_inactive") == "true"
	depts, err := h.deptRepo.List(c.Context(), includeInactive)
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, depts)
}

type CreateDeptReq struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Create godoc
// @Summary Create department
// @Description Create a new department.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param request body CreateDeptReq true "Department Creation Payload"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /admin/departments [post]
// @Security BearerAuth
func (h *AdminDepartmentHandler) Create(c *fiber.Ctx) error {
	var req CreateDeptReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.createDeptCmd.Handle(c.Context(), command.CreateDepartmentCommand{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, res)
}
