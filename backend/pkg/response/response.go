package response

import (
	"his-system/pkg/errors"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Response{
		Success: true,
		Data:    data,
	})
}

func OKWithMeta(c *fiber.Ctx, data any, meta *Meta) error {
	return c.JSON(Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Fail(c *fiber.Ctx, err *errors.AppError) error {
	return c.Status(err.Status).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}
