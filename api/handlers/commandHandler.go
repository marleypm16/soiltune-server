package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"soiltune-consumer/internal/apperrors"
	"soiltune-consumer/internal/models"

	"github.com/gofiber/fiber/v3"
)

type CommandHandler struct {
	service commandExecutor
}

type commandExecutor interface {
	Execute(sensorID string, command models.Command) error
}

func NewCommandHandler(service commandExecutor) *CommandHandler {
	return &CommandHandler{service: service}
}

func (h *CommandHandler) Handle(c fiber.Ctx) error {
	sensorID := c.Params("sensorId")
	if !models.IsValidSensorID(sensorID) {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid sensorId")
	}

	comando := c.Body()
	if len(comando) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Missing command in request body")
	}

	var payload models.Command
	decoder := json.NewDecoder(bytes.NewReader(comando))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return c.Status(fiber.StatusBadRequest).SendString("Request body must contain one JSON object")
	}
	if payload.Command == nil {
		return c.Status(fiber.StatusBadRequest).SendString("Missing command field in request body")
	}
	if !payload.IsValid() {
		return c.Status(fiber.StatusBadRequest).SendString("Command must be 0 (off) or 1 (on)")
	}

	if err := h.service.Execute(sensorID, payload); err != nil {
		if errors.Is(err, apperrors.ErrUnavailable) {
			return c.Status(fiber.StatusServiceUnavailable).SendString("MQTT client is not connected")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error occurred while processing command")
	}

	return c.SendStatus(fiber.StatusAccepted)
}
