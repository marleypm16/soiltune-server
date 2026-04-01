package main

import (
	"soiltune-consumer/api/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New(fiber.Config{
		CaseSensitive: false,
		BodyLimit:     256 * 1024, // 256 KB,
		AppName:       "soiltune-api",
	})

	routes.SetupRoutes(app)

	app.Listen(":8000")
}
