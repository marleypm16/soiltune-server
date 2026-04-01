package mqttHandler

import (
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func MqttHandler(messageHandler mqtt.MessageHandler) {
	// MQTT broker configuration
	mqttBroker := os.Getenv("MQTTBROKER")
	mqttTopic := os.Getenv("MQTTTOPIC")

	// Create MQTT client options
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID("mqtt-influxdb-bridge")
	opts.SetDefaultPublishHandler(messageHandler)

	// Connect to MQTT broker
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Subscribe to the MQTT topic
	if token := client.Subscribe(mqttTopic, 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to MQTT topic: %v", token.Error())
	}

	// Keep the program running
	select {}
}
