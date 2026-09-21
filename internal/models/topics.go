package models

import "strings"

const (
	TelemetryTopicPrefix = "soiltune/telemetry/"
	TelemetryTopicFilter = TelemetryTopicPrefix + "#"
	CommandTopicPrefix   = "soiltune/commands/"
)

func TelemetryTopic(sensorID string) string {
	return TelemetryTopicPrefix + sensorID
}

func CommandTopic(sensorID string) string {
	return CommandTopicPrefix + sensorID
}

func SensorIDFromTelemetryTopic(topic string) (string, bool) {
	if !strings.HasPrefix(topic, TelemetryTopicPrefix) {
		return "", false
	}

	sensorID := strings.TrimPrefix(topic, TelemetryTopicPrefix)
	return sensorID, IsValidSensorID(sensorID)
}
