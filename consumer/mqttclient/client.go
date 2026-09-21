package mqttclient

import (
	"fmt"
	"log"
	"time"

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
		SetPassword(cfg.Password).
		SetCleanSession(false).
		SetResumeSubs(true).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second).
		SetOrderMatters(false)
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

	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		token := client.Connect()
		if !token.WaitTimeout(10 * time.Second) {
			lastErr = fmt.Errorf("mqtt connection timed out")
		} else if token.Error() == nil {
			return client, nil
		} else {
			lastErr = token.Error()
		}

		if attempt < 10 {
			delay := time.Duration(attempt) * time.Second
			log.Printf("mqtt connection failed; retrying: attempt=%d/10 delay=%s error=%v", attempt, delay, lastErr)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("mqtt connection failed after 10 attempts: %w", lastErr)
}
