package mqttclient

import (
	"fmt"
	"log"

	"soiltune-consumer/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Connect(cfg config.MQTTConfig) (mqtt.Client, error) {
	if cfg.Broker == "" {
		return nil, fmt.Errorf("mqtt broker is required")
	}

	opts := mqtt.NewClientOptions().AddBroker(cfg.Broker).SetClientID(cfg.ClientID)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("mqtt connected: broker=%s client_id=%s", cfg.Broker, cfg.ClientID)
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	})
	opts.SetReconnectingHandler(func(c mqtt.Client, co *mqtt.ClientOptions) {
		log.Printf("mqtt reconnecting: broker=%s client_id=%s", cfg.Broker, cfg.ClientID)
	})

	log.Printf("mqtt connecting: broker=%s client_id=%s", cfg.Broker, cfg.ClientID)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
