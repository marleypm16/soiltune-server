package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"soiltune-consumer/api/handlers"
	"soiltune-consumer/api/middleware"
	"soiltune-consumer/api/repository"
	"soiltune-consumer/api/routes"
	"soiltune-consumer/api/services"
	"soiltune-consumer/consumer/mqttclient"
	"soiltune-consumer/internal/config"

	"github.com/gofiber/fiber/v3"
)

func main() {
	log.Printf("starting soiltune command API")

	apiConfig, err := config.LoadAPIConfig()
	if err != nil {
		log.Fatal(err)
	}

	mqttConfig, err := config.LoadMQTTConfig("soiltune-api", false)
	if err != nil {
		log.Fatal(err)
	}

	mqttClient, err := mqttclient.Connect(mqttConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer mqttClient.Disconnect(250)

	repo := repository.NewCommandRepository(mqttClient, mqttConfig.QoS)
	service := services.NewCommandService(repo)
	commandHandler := handlers.NewCommandHandler(service)

	app := fiber.New(fiber.Config{
		CaseSensitive: false,
		BodyLimit:     256 * 1024,
		AppName:       "soiltune-api",
	})

	routes.SetupRoutes(app, commandHandler, middleware.RequireAPIKey(apiConfig.Key), mqttClient.IsConnected)
	log.Printf("api listening on %s", apiConfig.Address)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(apiConfig.Address)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-shutdown:
		log.Printf("stopping soiltune command API")
		if err := app.Shutdown(); err != nil {
			log.Printf("error shutting down API server: %v", err)
		}
	}
}
