package models

import (
	"strings"
	"testing"
	"time"
)

func TestSensorDataValidate(t *testing.T) {
	now := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	valid := SensorData{
		Version:     SensorPayloadVersion,
		SensorID:    "sensor-01",
		RecordedAt:  now,
		Temperature: floatPointer(24.5),
		Humidity:    floatPointer(60),
		Weight:      floatPointer(125.3),
	}

	tests := []struct {
		name    string
		mutate  func(*SensorData)
		wantErr string
	}{
		{name: "valid"},
		{name: "unsupported version", mutate: func(data *SensorData) { data.Version = 2 }, wantErr: "version"},
		{name: "invalid sensor id", mutate: func(data *SensorData) { data.SensorID = "sensor/other" }, wantErr: "sensor_id"},
		{name: "missing timestamp", mutate: func(data *SensorData) { data.RecordedAt = time.Time{} }, wantErr: "recorded_at"},
		{name: "missing temperature", mutate: func(data *SensorData) { data.Temperature = nil }, wantErr: "temperature"},
		{name: "invalid humidity", mutate: func(data *SensorData) { data.Humidity = floatPointer(101) }, wantErr: "humidity"},
		{name: "negative weight", mutate: func(data *SensorData) { data.Weight = floatPointer(-1) }, wantErr: "weight"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := valid
			if tt.mutate != nil {
				tt.mutate(&data)
			}
			err := data.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidSensorID(t *testing.T) {
	for _, sensorID := range []string{"sensor-01", "field_sensor_2"} {
		if !IsValidSensorID(sensorID) {
			t.Errorf("expected %q to be valid", sensorID)
		}
	}
	for _, sensorID := range []string{"", "-sensor", "sensor/other", "sensor+#", strings.Repeat("a", 65)} {
		if IsValidSensorID(sensorID) {
			t.Errorf("expected %q to be invalid", sensorID)
		}
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
