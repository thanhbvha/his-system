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
	updateProfileCmd command.UpdateStaffProfileHandler
	updateEmailCmd command.UpdateUserEmailHandler
}

func NewAdminUserHandler(
	getUserByIDCmd *query.GetUserByIDHandler,
	listUsersCmd *query.ListUsersHandler,
	createStaffCmd *command.CreateStaffHandler,
	deactivateCmd *command.DeactivateUserHandler,
	assignRolesCmd *command.AssignUserRolesHandler,
	updateProfileCmd *command.UpdateStaffProfileHandler,
	updateEmailCmd *command.UpdateUserEmailHandler,
) *AdminUserHandler {
	return &AdminUserHandler{
		getUserByIDCmd: *getUserByIDCmd,
		listUsersCmd:   *listUsersCmd,
		createStaffCmd: *createStaffCmd,
		deactivateCmd:  *deactivateCmd,
		assignRolesCmd: *assignRolesCmd,
		updateProfileCmd: *updateProfileCmd,
		updateEmailCmd: *updateEmailCmd,
	}
}

// List godoc
// @Summary List users
// @Description Retrieve a list of staff users.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response
// @Failure 401,403,500 {object} response.Response
// @Router /admin/users [get]
// @Security BearerAuth
func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	res, err := h.listUsersCmd.Handle(c.Context(), query.ListUsersQuery{Page: page, Limit: limit, Search: search})
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

// GetByID godoc
// @Summary Get user by ID
// @Description Retrieve details of a specific staff user.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403,404,500 {object} response.Response
// @Router /admin/users/{id} [get]
// @Security BearerAuth
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

type CreateStaffReq struct {
	Username     string      `json:"username"`
	Password     string      `json:"password"`
	Email        string      `json:"email"`
	RoleIDs      []uuid.UUID `json:"role_ids"`
	DepartmentID uuid.UUID   `json:"department_id"`
	FullName     string      `json:"full_name"`
}

// Create godoc
// @Summary Create staff user
// @Description Create a new staff user.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param request body CreateStaffReq true "Staff Creation Payload"
// @Success 200 {object} response.Response
// @Failure 400,401,403,500 {object} response.Response
// @Router /admin/users [post]
// @Security BearerAuth
func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	var req CreateStaffReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	res, err := h.createStaffCmd.Handle(c.Context(), command.CreateStaffCommand{
		Username:     req.Username,
		Password:     req.Password,
		Email:        req.Email,
		RoleIDs:      req.RoleIDs,
		DepartmentID: req.DepartmentID,
		FullName:     req.FullName,
	})

	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	return response.OK(c, res)
}

// Deactivate godoc
// @Summary Deactivate user
// @Description Deactivate a staff user.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403,404,500 {object} response.Response
// @Router /admin/users/{id}/deactivate [put]
// @Security BearerAuth
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

type AssignRolesReq struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
}

// AssignRoles godoc
// @Summary Assign roles
// @Description Assign roles to a staff user.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body AssignRolesReq true "Role Assignment Payload"
// @Success 200 {object} response.Response
// @Failure 400,401,403,404,500 {object} response.Response
// @Router /admin/users/{id}/roles [put]
// @Security BearerAuth
func (h *AdminUserHandler) AssignRoles(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req AssignRolesReq
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

type UpdateProfileReq struct {
	FullName     string      `json:"full_name"`
	DepartmentID uuid.UUID   `json:"department_id"`
	RoleIDs      []uuid.UUID `json:"role_ids"`
	Email        string      `json:"email"`
}

// UpdateProfile godoc
// @Summary Update profile
// @Description Update staff profile details and roles.
// @Tags Admin (User/Role)
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateProfileReq true "Profile Update Payload"
// @Success 200 {object} response.Response
// @Failure 400,401,403,404,500 {object} response.Response
// @Router /admin/users/{id}/profile [put]
// @Security BearerAuth
func (h *AdminUserHandler) UpdateProfile(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	var req UpdateProfileReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, appErrors.ErrValidation)
	}

	err = h.updateProfileCmd.Handle(c.Context(), command.UpdateStaffProfileCommand{
		UserID:       id,
		FullName:     req.FullName,
		DepartmentID: req.DepartmentID,
	})
	if err != nil {
		return response.Fail(c, appErrors.ErrInternal)
	}

	if len(req.RoleIDs) > 0 {
		_, err = h.assignRolesCmd.Handle(c.Context(), command.AssignUserRolesCommand{
			UserID:  id,
			RoleIDs: req.RoleIDs,
		})
		if err != nil {
			return response.Fail(c, appErrors.ErrInternal)
		}
	}

	if req.Email != "" {
		err = h.updateEmailCmd.Handle(c.Context(), command.UpdateUserEmailCommand{
			UserID: id,
			Email:  req.Email,
		})
		if err != nil {
			return response.Fail(c, appErrors.ErrInternal)
		}
	}

	return response.OK(c, nil)
}
