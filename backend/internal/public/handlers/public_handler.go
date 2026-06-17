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
	serviceID := c.Query("service_id")
	
	doctors := []fiber.Map{
		{"id": "doc-1", "full_name": "Dr. John Doe", "specialty": "Cardiology", "service_id": "srv-1"},
		{"id": "doc-2", "full_name": "Dr. Jane Smith", "specialty": "Neurology", "service_id": "srv-2"},
		{"id": "doc-3", "full_name": "Dr. Emily Chen", "specialty": "Pediatrics", "service_id": "srv-3"},
		{"id": "doc-4", "full_name": "Dr. Michael Brown", "specialty": "General Practice", "service_id": "srv-1"},
	}

	if serviceID != "" {
		filtered := []fiber.Map{}
		for _, doc := range doctors {
			if doc["service_id"] == serviceID {
				filtered = append(filtered, doc)
			}
		}
		return response.OK(c, filtered)
	}

	return response.OK(c, doctors)
}

func (h *PublicHandler) ListServices(c *fiber.Ctx) error {
	services := []fiber.Map{
		{"id": "srv-1", "name": "General Checkup", "price": 500000, "duration_minutes": 30},
		{"id": "srv-2", "name": "Neurology Consultation", "price": 800000, "duration_minutes": 45},
		{"id": "srv-3", "name": "Pediatric Checkup", "price": 400000, "duration_minutes": 30},
	}
	return response.OK(c, services)
}
