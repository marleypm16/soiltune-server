package handlers

import "testing"

func TestSensorIDPattern(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{id: "sensor-01", want: true},
		{id: "field_sensor_2", want: true},
		{id: "", want: false},
		{id: "-sensor", want: false},
		{id: "sensor/other", want: false},
		{id: "sensor+#", want: false},
		{id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", want: false},
	}

	for _, tt := range tests {
		if got := sensorIDPattern.MatchString(tt.id); got != tt.want {
			t.Errorf("sensorIDPattern.MatchString(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}
