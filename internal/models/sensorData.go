package models

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

const SensorPayloadVersion = 1

var sensorIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type SensorData struct {
	Version     int       `json:"version"`
	SensorID    string    `json:"sensor_id"`
	RecordedAt  time.Time `json:"recorded_at"`
	Temperature *float64  `json:"temperature"`
	Humidity    *float64  `json:"humidity"`
	Weight      *float64  `json:"weight"`
	State       *string   `json:"state,omitempty"`
}

func IsValidSensorID(sensorID string) bool {
	return sensorIDPattern.MatchString(sensorID)
}

func (data SensorData) Validate() error {
	if data.Version != SensorPayloadVersion {
		return fmt.Errorf("version must be %d", SensorPayloadVersion)
	}
	if !IsValidSensorID(data.SensorID) {
		return fmt.Errorf("sensor_id is invalid")
	}
	if data.RecordedAt.IsZero() {
		return fmt.Errorf("recorded_at is required")
	}
	if data.Temperature == nil {
		return fmt.Errorf("temperature is required")
	}
	if math.IsNaN(*data.Temperature) || math.IsInf(*data.Temperature, 0) || *data.Temperature < -273.15 {
		return fmt.Errorf("temperature is invalid")
	}
	if data.Humidity == nil {
		return fmt.Errorf("humidity is required")
	}
	if math.IsNaN(*data.Humidity) || math.IsInf(*data.Humidity, 0) || *data.Humidity < 0 || *data.Humidity > 100 {
		return fmt.Errorf("humidity must be between 0 and 100")
	}
	if data.Weight == nil {
		return fmt.Errorf("weight is required")
	}
	if math.IsNaN(*data.Weight) || math.IsInf(*data.Weight, 0) || *data.Weight < 0 {
		return fmt.Errorf("weight must be greater than or equal to 0")
	}
	if data.State != nil {
		state := strings.TrimSpace(*data.State)
		if state == "" || len(state) > 32 {
			return fmt.Errorf("state must contain between 1 and 32 characters")
		}
	}
	return nil
}
