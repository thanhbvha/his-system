package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"his-system/internal/identity/application/command"
	"his-system/internal/identity/application/query"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

type AdminUserHandler struct {
	getUserByIDCmd query.GetUserByIDHandler
	listUsersCmd   query.ListUsersHandler
	createStaffCmd command.CreateStaffHandler
	deactivateCmd  command.DeactivateUserHandler
	assignRolesCmd command.AssignUserRolesHandler
}

func NewAdminUserHandler(
	getUserByIDCmd *query.GetUserByIDHandler,
	listUsersCmd *query.ListUsersHandler,
	createStaffCmd *command.CreateStaffHandler,
	deactivateCmd *command.DeactivateUserHandler,
	assignRolesCmd *command.AssignUserRolesHandler,
) *AdminUserHandler {
	return &AdminUserHandler{
		getUserByIDCmd: *getUserByIDCmd,
		listUsersCmd:   *listUsersCmd,
		createStaffCmd: *createStaffCmd,
		deactivateCmd:  *deactivateCmd,
		assignRolesCmd: *assignRolesCmd,
	}
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	res, err := h.listUsersCmd.Handle(c.Context(), query.ListUsersQuery{Page: page, Limit: limit})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: int(res.Total),
	}

	return response.OKWithMeta(c, res.Users, meta)
}

func (h *AdminUserHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.getUserByIDCmd.Handle(c.Context(), query.GetUserByIDQuery{ID: id})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}
	if res.User == nil {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.OK(c, res.User)
}

type createStaffReq struct {
	Username     string      `json:"username"`
	Password     string      `json:"password"`
	Email        string      `json:"email"`
	RoleIDs      []uuid.UUID `json:"role_ids"`
	DepartmentID uuid.UUID   `json:"department_id"`
}

func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	var req createStaffReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.createStaffCmd.Handle(c.Context(), command.CreateStaffCommand{
		Username:     req.Username,
		Password:     req.Password,
		Email:        req.Email,
		RoleIDs:      req.RoleIDs,
		DepartmentID: req.DepartmentID,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, res)
}

func (h *AdminUserHandler) Deactivate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.deactivateCmd.Handle(c.Context(), command.DeactivateUserCommand{UserID: id})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}
	if res == nil {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.OK(c, nil)
}

type assignRolesReq struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
}

func (h *AdminUserHandler) AssignRoles(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req assignRolesReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.assignRolesCmd.Handle(c.Context(), command.AssignUserRolesCommand{
		UserID:  id,
		RoleIDs: req.RoleIDs,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}
	if res == nil {
		return response.Fail(c, appErrors.ErrNotFound)
	}

	return response.OK(c, nil)
}
