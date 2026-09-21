package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"soiltune-consumer/internal/config"
	"soiltune-consumer/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type pointWriter interface {
	WritePoint(ctx context.Context, point ...*write.Point) error
}

type Service struct {
	client        influxdb2.Client
	writer        pointWriter
	writeTimeout  time.Duration
	writeAttempts int
	retryDelay    func(attempt int) time.Duration
}

func NewService(cfg config.InfluxConfig) (*Service, error) {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &Service{
		client:        client,
		writer:        client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
		writeTimeout:  cfg.WriteTimeout,
		writeAttempts: cfg.WriteAttempts,
		retryDelay: func(attempt int) time.Duration {
			return time.Duration(1<<(attempt-1)) * 100 * time.Millisecond
		},
	}, nil
}

func (s *Service) Close() {
	if s != nil && s.client != nil {
		s.client.Close()
	}
}

func (s *Service) Handler() mqtt.MessageHandler {
	return func(_ mqtt.Client, msg mqtt.Message) {
		if err := s.ProcessMessage(msg.Topic(), msg.Payload()); err != nil {
			log.Printf("telemetry rejected: topic=%s error=%v", msg.Topic(), err)
			return
		}

		log.Printf("sensor data written to InfluxDB: topic=%s", msg.Topic())
	}
}

func (s *Service) ProcessMessage(topic string, payload []byte) error {
	data, err := decodeSensorData(payload)
	if err != nil {
		return fmt.Errorf("decode telemetry: %w", err)
	}

	topicSensorID, ok := models.SensorIDFromTelemetryTopic(topic)
	if !ok {
		return fmt.Errorf("topic must match %q", models.TelemetryTopicFilter)
	}
	if data.SensorID != topicSensorID {
		return fmt.Errorf("sensor_id %q does not match topic device %q", data.SensorID, topicSensorID)
	}

	return s.writePoint(newSensorPoint(data))
}

func decodeSensorData(payload []byte) (models.SensorData, error) {
	var data models.SensorData
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		return models.SensorData{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return models.SensorData{}, fmt.Errorf("payload must contain one JSON object")
	}
	if err := data.Validate(); err != nil {
		return models.SensorData{}, err
	}

	return data, nil
}

func newSensorPoint(data models.SensorData) *write.Point {
	point := influxdb2.NewPointWithMeasurement("sensor_data").
		AddTag("sensor_id", data.SensorID).
		AddTag("schema_version", fmt.Sprintf("v%d", data.Version)).
		AddField("temperature", *data.Temperature).
		AddField("humidity", *data.Humidity).
		AddField("weight", *data.Weight).
		SetTime(data.RecordedAt.UTC())

	if data.State != nil {
		point.AddField("state", strings.TrimSpace(*data.State))
	}
	return point
}

func (s *Service) writePoint(point *write.Point) error {
	if s == nil || s.writer == nil {
		return fmt.Errorf("influx writer is required")
	}

	attempts := s.writeAttempts
	if attempts < 1 {
		attempts = 1
	}
	timeout := s.writeTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		lastErr = s.writer.WritePoint(ctx, point)
		cancel()
		if lastErr == nil {
			return nil
		}
		if attempt < attempts {
			delay := 100 * time.Millisecond
			if s.retryDelay != nil {
				delay = s.retryDelay(attempt)
			}
			log.Printf("influx write failed; retrying: attempt=%d/%d delay=%s error=%v", attempt, attempts, delay, lastErr)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("write telemetry after %d attempts: %w", attempts, lastErr)
}
