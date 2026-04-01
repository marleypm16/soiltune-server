package main

import (
	"soiltune-consumer/influxdb"
	"soiltune-consumer/mqttHandler"
)

func main() {
	influxdb.InfluxDBHandler()
	messageHandler := influxdb.InfluxDBHandler()
	mqttHandler.MqttHandler(messageHandler)
}
