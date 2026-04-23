package handlers

import (
	"encoding/json"
	"errors"

	"soiltune-consumer/api/services"
	"soiltune-consumer/internal/models"

	"github.com/gofiber/fiber/v3"
)

type CommandHandler struct {
	service *services.CommandService
}

func NewCommandHandler(service *services.CommandService) *CommandHandler {
	return &CommandHandler{service: service}
}

func (h *CommandHandler) Handle(c fiber.Ctx) error {
	sensorID := c.Params("sensorId")
	if sensorID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing sensorId")
	}

	comando := c.Body()
	if len(comando) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Missing command in request body")
	}

	var payload models.Command
	if err := json.Unmarshal(comando, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON body")
	}
	if payload.Command == nil {
		return c.Status(fiber.StatusBadRequest).SendString("Missing command field in request body")
	}

	if err := h.service.Execute(sensorID, payload); err != nil {
		if errors.Is(err, fiber.ErrServiceUnavailable) {
			return c.Status(fiber.StatusServiceUnavailable).SendString("MQTT client is not connected")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error occurred while processing command")
	}

	return c.SendStatus(fiber.StatusOK)
}
