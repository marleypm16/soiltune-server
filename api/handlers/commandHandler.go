package handlers

import (
	"soiltune-consumer/api/services"

	"github.com/gofiber/fiber/v3"
)

func CommandHandler(c fiber.Ctx) error {
	sensorID := c.Params("sensorId")
	comando := c.Body()
	service := services.CommandService(sensorID, comando)
	if service != nil {
		return c.Status(500).SendString("Error occurred while processing command")
	}
	return c.SendStatus(200)
}
