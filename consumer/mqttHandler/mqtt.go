package mqttHandler

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Subscribe(client mqtt.Client, topic string, qos byte, handler mqtt.MessageHandler) error {
	if client == nil {
		return fmt.Errorf("mqtt client is required")
	}
	if topic == "" {
		return fmt.Errorf("mqtt topic is required")
	}
	if handler == nil {
		return fmt.Errorf("mqtt message handler is required")
	}
	if qos > 2 {
		return fmt.Errorf("mqtt qos must be 0, 1 or 2")
	}

	log.Printf("subscribing to mqtt topic: topic=%s qos=%d", topic, qos)
	if token := client.Subscribe(topic, qos, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Printf("mqtt subscription active: topic=%s", topic)

	return nil
}
