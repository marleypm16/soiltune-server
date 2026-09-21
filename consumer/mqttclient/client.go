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
	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("mqtt credentials are required")
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("mqtt connected: client_id=%s", cfg.ClientID)
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	})
	opts.SetReconnectingHandler(func(c mqtt.Client, co *mqtt.ClientOptions) {
		log.Printf("mqtt reconnecting: client_id=%s", cfg.ClientID)
	})

	log.Printf("mqtt connecting: client_id=%s", cfg.ClientID)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
