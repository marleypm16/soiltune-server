package routes

import (
	"soiltune-consumer/api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, commandHandler *handlers.CommandHandler, sensorHandler *handlers.SensorHandler) {
	app.Post("/command/:sensorId", commandHandler.Handle)
	app.Get("/sensorids", sensorHandler.GetSensorIDs)
	app.Get("/sensor/:id", sensorHandler.GetSensorData)
}
