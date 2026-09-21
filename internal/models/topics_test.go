package models

import "testing"

func TestTopicContract(t *testing.T) {
	if got := TelemetryTopic("sensor-01"); got != "soiltune/telemetry/sensor-01" {
		t.Fatalf("TelemetryTopic() = %q", got)
	}
	if got := CommandTopic("sensor-01"); got != "soiltune/commands/sensor-01" {
		t.Fatalf("CommandTopic() = %q", got)
	}
	if got, ok := SensorIDFromTelemetryTopic("soiltune/telemetry/sensor-01"); !ok || got != "sensor-01" {
		t.Fatalf("SensorIDFromTelemetryTopic() = %q, %v", got, ok)
	}
	if _, ok := SensorIDFromTelemetryTopic("other/sensor-01"); ok {
		t.Fatal("unexpected valid topic")
	}
}
