package handlers

import (
	"encoding/json"
	"errors"
	"soiltune-consumer/api/services"
	"soiltune-consumer/models"

	"github.com/gofiber/fiber/v3"
)

func CommandHandler(c fiber.Ctx) error {
	sensorID := c.Params("sensorId")
	if sensorID == "" {
		return c.Status(400).SendString("Missing sensorId")
	}

	comando := c.Body()
	if len(comando) == 0 {
		return c.Status(400).SendString("Missing command in request body")
	}

	var payload models.Command
	if err := json.Unmarshal(comando, &payload); err != nil {
		return c.Status(400).SendString("Invalid JSON body")
	}
	if payload.Command == nil {
		return c.Status(400).SendString("Missing command field in request body")
	}

	service := services.CommandService(sensorID, comando)
	if service != nil {
		if errors.Is(service, fiber.ErrServiceUnavailable) {
			return c.Status(503).SendString("MQTT client is not connected")
		}
		return c.Status(500).SendString("Error occurred while processing command")
	}
	return c.SendStatus(200)
}
