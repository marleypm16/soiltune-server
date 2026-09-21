package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"soiltune-consumer/internal/apperrors"
	"soiltune-consumer/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CommandRepository struct {
	client mqttPublisher
	qos    byte
}

type mqttPublisher interface {
	IsConnected() bool
	Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token
}

func NewCommandRepository(client mqttPublisher, qos byte) *CommandRepository {
	return &CommandRepository{client: client, qos: qos}
}

func (r *CommandRepository) Publish(sensorID string, command models.Command) error {
	if r == nil || r.client == nil || !r.client.IsConnected() {
		log.Printf("mqtt publish skipped: client unavailable for sensor_id=%s", sensorID)
		return apperrors.ErrUnavailable
	}

	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	topic := models.CommandTopic(sensorID)
	log.Printf("publishing command: topic=%s", topic)
	token := r.client.Publish(topic, r.qos, false, payload)
	if !token.WaitTimeout(5 * time.Second) {
		log.Printf("mqtt publish timeout: topic=%s", topic)
		return apperrors.ErrUnavailable
	}

	if err := token.Error(); err != nil {
		log.Printf("mqtt publish failed: topic=%s error=%v", topic, err)
		return fmt.Errorf("%w: %v", apperrors.ErrUnavailable, err)
	}

	log.Printf("mqtt publish success: topic=%s", topic)
	return nil
}
