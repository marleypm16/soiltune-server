package influxdb

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"soiltune-consumer/internal/config"
	"soiltune-consumer/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Service struct {
	client influxdb2.Client
	org    string
	bucket string
}

func NewService(cfg config.InfluxConfig) (*Service, error) {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &Service{client: client, org: cfg.Org, bucket: cfg.Bucket}, nil
}

func (s *Service) Close() {
	if s != nil && s.client != nil {
		s.client.Close()
	}
}

// Query executes a Flux query against InfluxDB and returns each record's values as a map.
func (s *Service) Query(ctx context.Context, flux string) ([]map[string]interface{}, error) {
	queryAPI := s.client.QueryAPI(s.org)
	result, err := queryAPI.Query(ctx, flux)
	if err != nil {
		return nil, err
	}

	var rows []map[string]interface{}
	for result.Next() {
		rec := result.Record()
		if rec == nil {
			continue
		}
		rows = append(rows, rec.Values())
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return rows, nil
}

func (s *Service) Handler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		var data models.SensorData
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			log.Printf("error parsing MQTT payload: %v", err)
			return
		}

		if data.SensorID == "" {
			log.Printf("ignoring sensor payload without sensor_id")
			return
		}

		point := influxdb2.NewPointWithMeasurement("sensor_data").
			AddTag("sensor_id", data.SensorID).
			AddField("temperatura", data.Temperature).
			AddField("umidade", data.Humidity).
			AddField("peso", data.Weight).
			AddField("estado", data.State).
			SetTime(time.Now().UTC())

		if err := s.client.WriteAPIBlocking(s.org, s.bucket).WritePoint(context.Background(), point); err != nil {
			log.Printf("error writing point to InfluxDB: %v", err)
			return
		}

		log.Printf("sensor data written to InfluxDB: sensor_id=%s", data.SensorID)
	}
}
