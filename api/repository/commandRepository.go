package repository

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

func InitMQTT(c mqtt.Client) {
	client = c
}

func CommandRepository(sensorId string, comando []byte) error {
	topic := "/commands/" + sensorId
	token := client.Publish(topic, 0, false, comando)
	token.Wait()
	return token.Error()
}
