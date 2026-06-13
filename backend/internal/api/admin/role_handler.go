package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"his-system/internal/identity/application/command"
	"his-system/internal/identity/application/query"
	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

type AdminRoleHandler struct {
	roleRepo           domain.RoleRepository
	getRolePermsCmd    query.GetRolePermissionsHandler
	updateRolePermsCmd command.UpdateRolePermissionsHandler
	listPermsCmd       query.ListPermissionsHandler
}

func NewAdminRoleHandler(
	roleRepo domain.RoleRepository,
	getRolePermsCmd *query.GetRolePermissionsHandler,
	updateRolePermsCmd *command.UpdateRolePermissionsHandler,
	listPermsCmd *query.ListPermissionsHandler,
) *AdminRoleHandler {
	return &AdminRoleHandler{
		roleRepo:           roleRepo,
		getRolePermsCmd:    *getRolePermsCmd,
		updateRolePermsCmd: *updateRolePermsCmd,
		listPermsCmd:       *listPermsCmd,
	}
}

func (h *AdminRoleHandler) List(c *fiber.Ctx) error {
	roles, err := h.roleRepo.List(c.Context())
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, roles)
}

func (h *AdminRoleHandler) ListPermissions(c *fiber.Ctx) error {
	res, err := h.listPermsCmd.Handle(c.Context(), query.ListPermissionsQuery{})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, res.Permissions)
}

func (h *AdminRoleHandler) GetPermissions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.getRolePermsCmd.Handle(c.Context(), query.GetRolePermissionsQuery{RoleID: id})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}
	if res.Role == nil {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.OK(c, res.Role.Permissions)
}

type updatePermissionsReq struct {
	Permissions []domain.Permission `json:"permissions"`
}

func (h *AdminRoleHandler) UpdatePermissions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req updatePermissionsReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.updateRolePermsCmd.Handle(c.Context(), command.UpdateRolePermissionsCommand{
		RoleID:      id,
		Permissions: req.Permissions,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}
	if res == nil {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.OK(c, nil)
}
