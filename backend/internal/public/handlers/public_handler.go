package handlers

import (
	"github.com/gofiber/fiber/v2"
	"his-system/pkg/response"
)

type PublicHandler struct {
	// Inject repositories or queries later
}

func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

func (h *PublicHandler) GetClinicInfo(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"name":    "HIS International Clinic",
		"address": "123 Healthcare Blvd, Tech City",
		"phone":   "1900 1000",
	})
}

func (h *PublicHandler) ListDoctors(c *fiber.Ctx) error {
	// TODO: implement ListDoctors query
	return response.OK(c, []fiber.Map{})
}

func (h *PublicHandler) ListServices(c *fiber.Ctx) error {
	// TODO: implement ListServices query
	return response.OK(c, []fiber.Map{})
}
