package routes

import (
	"soiltune-consumer/api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, commandHandler *handlers.CommandHandler, authenticate fiber.Handler) {
	app.Post("/command/:sensorId", authenticate, commandHandler.Handle)
}
