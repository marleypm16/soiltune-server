package routes

import (
	"soiltune-consumer/api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, commandHandler *handlers.CommandHandler, authenticate fiber.Handler, ready func() bool) {
	app.Get("/health/live", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/health/ready", func(c fiber.Ctx) error {
		if ready == nil || !ready() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})
	app.Post("/command/:sensorId", authenticate, commandHandler.Handle)
}
