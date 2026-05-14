package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"soiltune-consumer/api/handlers"
	"soiltune-consumer/api/repository"
	"soiltune-consumer/api/routes"
	"soiltune-consumer/api/services"
	"soiltune-consumer/consumer/mqttclient"
	"soiltune-consumer/internal/config"

	"github.com/gofiber/fiber/v3"
)

func main() {
	log.Printf("starting soiltune command API")

	mqttConfig, err := config.LoadMQTTConfig("soiltune-api", false)
	if err != nil {
		log.Fatal(err)
	}

	mqttClient, err := mqttclient.Connect(mqttConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer mqttClient.Disconnect(250)

	repo := repository.NewCommandRepository(mqttClient)
	service := services.NewCommandService(repo)
	commandHandler := handlers.NewCommandHandler(service)

	// InfluxDB service and sensor handler
	influxCfg, err := config.LoadInfluxConfig()
	if err != nil {
		log.Fatalf("loading influx config: %v", err)
	}
	influxSvc, err := services.NewInfluxService(influxCfg)
	if err != nil {
		log.Fatalf("creating influx service: %v", err)
	}
	defer influxSvc.Close()

	sensorHandler := handlers.NewSensorHandler(influxSvc)

	app := fiber.New(fiber.Config{
		CaseSensitive: false,
		BodyLimit:     256 * 1024,
		AppName:       "soiltune-api",
	})

	routes.SetupRoutes(app, commandHandler, sensorHandler)
	log.Printf("api listening on :8000")

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(":8000")
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
