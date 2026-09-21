package influxdb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type fakePointWriter struct {
	calls     int
	failUntil int
	point     *write.Point
}

func (fake *fakePointWriter) WritePoint(_ context.Context, points ...*write.Point) error {
	fake.calls++
	if len(points) > 0 {
		fake.point = points[0]
	}
	if fake.calls <= fake.failUntil {
		return errors.New("temporary write failure")
	}
	return nil
}

func TestProcessMessageMapsValidatedTelemetry(t *testing.T) {
	writer := &fakePointWriter{}
	service := testService(writer)
	payload := []byte(`{
		"version": 1,
		"sensor_id": "sensor-01",
		"recorded_at": "2026-09-21T12:30:00Z",
		"temperature": 24.5,
		"humidity": 60.2,
		"weight": 150.0,
		"state": "on"
	}`)

	if err := service.ProcessMessage("soiltune/telemetry/sensor-01", payload); err != nil {
		t.Fatalf("ProcessMessage() error = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("writer calls = %d, want 1", writer.calls)
	}
	if writer.point == nil || writer.point.Name() != "sensor_data" {
		t.Fatalf("unexpected point: %#v", writer.point)
	}
	if got := writer.point.Time(); !got.Equal(time.Date(2026, time.September, 21, 12, 30, 0, 0, time.UTC)) {
		t.Fatalf("point time = %s", got)
	}

	fields := make(map[string]interface{})
	for _, field := range writer.point.FieldList() {
		fields[field.Key] = field.Value
	}
	if fields["temperature"] != 24.5 || fields["humidity"] != 60.2 || fields["weight"] != 150.0 || fields["state"] != "on" {
		t.Fatalf("unexpected fields: %#v", fields)
	}

	tags := make(map[string]string)
	for _, tag := range writer.point.TagList() {
		tags[tag.Key] = tag.Value
	}
	if tags["sensor_id"] != "sensor-01" || tags["schema_version"] != "v1" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}

func TestProcessMessageRejectsInvalidContract(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		payload string
	}{
		{
			name:  "unknown field",
			topic: "soiltune/telemetry/sensor-01",
			payload: `{"version":1,"sensor_id":"sensor-01","recorded_at":"2026-09-21T12:30:00Z",
				"temperature":24.5,"humidity":60,"weight":100,"temperatura":24.5}`,
		},
		{
			name:  "topic sensor mismatch",
			topic: "soiltune/telemetry/sensor-02",
			payload: `{"version":1,"sensor_id":"sensor-01","recorded_at":"2026-09-21T12:30:00Z",
				"temperature":24.5,"humidity":60,"weight":100}`,
		},
		{
			name:  "missing required value",
			topic: "soiltune/telemetry/sensor-01",
			payload: `{"version":1,"sensor_id":"sensor-01","recorded_at":"2026-09-21T12:30:00Z",
				"humidity":60,"weight":100}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &fakePointWriter{}
			service := testService(writer)
			if err := service.ProcessMessage(tt.topic, []byte(tt.payload)); err == nil {
				t.Fatal("ProcessMessage() error = nil")
			}
			if writer.calls != 0 {
				t.Fatalf("writer calls = %d, want 0", writer.calls)
			}
		})
	}
}

func TestProcessMessageRetriesInfluxWrite(t *testing.T) {
	writer := &fakePointWriter{failUntil: 2}
	service := testService(writer)
	service.writeAttempts = 3
	payload := []byte(`{"version":1,"sensor_id":"sensor-01","recorded_at":"2026-09-21T12:30:00Z",
		"temperature":24.5,"humidity":60,"weight":100}`)

	if err := service.ProcessMessage("soiltune/telemetry/sensor-01", payload); err != nil {
		t.Fatalf("ProcessMessage() error = %v", err)
	}
	if writer.calls != 3 {
		t.Fatalf("writer calls = %d, want 3", writer.calls)
	}
}

func testService(writer pointWriter) *Service {
	return &Service{
		writer:        writer,
		writeTimeout:  time.Second,
		writeAttempts: 1,
		retryDelay:    func(int) time.Duration { return 0 },
	}
}
