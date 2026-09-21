package handlers

import (
	"encoding/json"
	"errors"
	"regexp"

	"soiltune-consumer/api/services"
	"soiltune-consumer/internal/models"

	"github.com/gofiber/fiber/v3"
)

var sensorIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type CommandHandler struct {
	service *services.CommandService
}

func NewCommandHandler(service *services.CommandService) *CommandHandler {
	return &CommandHandler{service: service}
}

func (h *CommandHandler) Handle(c fiber.Ctx) error {
	sensorID := c.Params("sensorId")
	if !sensorIDPattern.MatchString(sensorID) {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid sensorId")
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
	if !payload.IsValid() {
		return c.Status(fiber.StatusBadRequest).SendString("Command must be 0 (off) or 1 (on)")
	}

	if err := h.service.Execute(sensorID, payload); err != nil {
		if errors.Is(err, fiber.ErrServiceUnavailable) {
			return c.Status(fiber.StatusServiceUnavailable).SendString("MQTT client is not connected")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error occurred while processing command")
	}

	return c.SendStatus(fiber.StatusAccepted)
}
