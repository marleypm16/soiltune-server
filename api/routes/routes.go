package routes

import (
	"soiltune-consumer/api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, commandHandler *handlers.CommandHandler) {
	app.Post("/command/:sensorId", commandHandler.Handle)
}
