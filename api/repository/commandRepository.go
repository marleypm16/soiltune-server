package repository

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v3"
)

var client mqtt.Client

func InitMQTT(c mqtt.Client) {
	client = c
}

func CommandRepository(sensorId string, comando []byte) error {
	topic := "/comandos/" + sensorId
	if client != nil && client.IsConnected() {
		token := client.Publish(topic, 0, false, comando)
		token.Wait()
		return token.Error()
	}
	return fiber.ErrServiceUnavailable
}
