package mqttHandler

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Subscribe(client mqtt.Client, topic string, handler mqtt.MessageHandler) error {
	if client == nil {
		return fmt.Errorf("mqtt client is required")
	}
	if topic == "" {
		return fmt.Errorf("mqtt topic is required")
	}
	if handler == nil {
		return fmt.Errorf("mqtt message handler is required")
	}

	log.Printf("subscribing to mqtt topic: topic=%s", topic)
	if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Printf("mqtt subscription active: topic=%s", topic)

	return nil
}
