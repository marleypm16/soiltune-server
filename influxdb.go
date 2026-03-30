package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func influxDBHandler() mqtt.MessageHandler {
	// Load InfluxDB configuration from environment variables
	influxDBURL := os.Getenv("DBINFLUX")
	influxDBToken := os.Getenv("DBINFLUXTOKEN")
	influxDBOrg := os.Getenv("DBINFLUXORG")
	influxDBBucket := os.Getenv("DBINFLUXBUCKET")

	// Create InfluxDB client
	influxClient := influxdb2.NewClient(influxDBURL, influxDBToken)
	defer influxClient.Close()

	// MQTT message handler
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		var data SensorData
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return
		}

		// Write data to InfluxDB
		writeAPI := influxClient.WriteAPIBlocking(influxDBOrg, influxDBBucket)
		p := influxdb2.NewPointWithMeasurement("sensor_data").
			AddTag("sensor_id", data.SensorID).
			AddField("temperature", data.Temperature).
			AddField("humidity", data.Humidity).
			AddField("weight", data.Weight).
			SetTime(time.Now())
		err = writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.Printf("Error writing to InfluxDB: %v", err)
		} else {
			fmt.Printf("Data written to InfluxDB: %v\n", data)
		}
	}
	return messageHandler
}
