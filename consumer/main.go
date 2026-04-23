package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"soiltune-consumer/consumer/influxdb"
	"soiltune-consumer/consumer/mqttHandler"
	"soiltune-consumer/consumer/mqttclient"
	"soiltune-consumer/internal/config"
)

func main() {
	log.Printf("starting soiltune ingestion service")

	influxConfig, err := config.LoadInfluxConfig()
	if err != nil {
		log.Fatal(err)
	}

	mqttConfig, err := config.LoadMQTTConfig("mqtt-influxdb-bridge", true)
	if err != nil {
		log.Fatal(err)
	}

	influxService, err := influxdb.NewService(influxConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer influxService.Close()

	mqttClient, err := mqttclient.Connect(mqttConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer mqttClient.Disconnect(250)

	if err := mqttHandler.Subscribe(mqttClient, mqttConfig.Topic, influxService.Handler()); err != nil {
		log.Fatal(err)
	}

	log.Printf("ingestion service ready: mqtt_topic=%s influx_org=%s influx_bucket=%s", mqttConfig.Topic, influxConfig.Org, influxConfig.Bucket)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	log.Printf("stopping soiltune ingestion service")
}
