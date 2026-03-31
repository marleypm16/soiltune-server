package routes

import (
	"soiltune-consumer/api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	app.Post("/command/:sensorId", handlers.CommandHandler)
}
