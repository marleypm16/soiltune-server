package repository

import (
	"encoding/json"
	"log"
	"time"

	"soiltune-consumer/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v3"
)

type CommandRepository struct {
	client mqtt.Client
}

func NewCommandRepository(client mqtt.Client) *CommandRepository {
	return &CommandRepository{client: client}
}

func (r *CommandRepository) Publish(sensorID string, command models.Command) error {
	if r == nil || r.client == nil || !r.client.IsConnected() {
		log.Printf("mqtt publish skipped: client unavailable for sensor_id=%s", sensorID)
		return fiber.ErrServiceUnavailable
	}

	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	topic := "/comandos/" + sensorID
	log.Printf("publishing command: topic=%s", topic)
	token := r.client.Publish(topic, 0, false, payload)
	if !token.WaitTimeout(5 * time.Second) {
		log.Printf("mqtt publish timeout: topic=%s", topic)
		return fiber.ErrServiceUnavailable
	}

	if err := token.Error(); err != nil {
		log.Printf("mqtt publish failed: topic=%s error=%v", topic, err)
		return err
	}

	log.Printf("mqtt publish success: topic=%s", topic)
	return nil
}
