//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"soiltune-consumer/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
)

func TestComposeFlows(t *testing.T) {
	_ = godotenv.Load("../.env")

	deviceID := requiredEnv(t, "MQTT_DEVICE_ID")
	devicePassword := requiredEnv(t, "MQTT_DEVICE_PASSWORD")
	apiKey := requiredEnv(t, "API_KEY")
	influxToken := requiredEnv(t, "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN")
	influxOrg := requiredEnv(t, "DOCKER_INFLUXDB_INIT_ORG")
	influxBucket := requiredEnv(t, "DOCKER_INFLUXDB_INIT_BUCKET")

	broker := envOrDefault("INTEGRATION_MQTT_BROKER", "tcp://localhost:1883")
	apiURL := envOrDefault("INTEGRATION_API_URL", "http://localhost:8080")
	influxURL := envOrDefault("INTEGRATION_INFLUX_URL", "http://localhost:8086")

	device := connectDevice(t, broker, deviceID, devicePassword)
	defer device.Disconnect(250)

	t.Run("HTTP command reaches device", func(t *testing.T) {
		commands := make(chan []byte, 1)
		topic := models.CommandTopic(deviceID)
		token := device.Subscribe(topic, 1, func(_ mqtt.Client, message mqtt.Message) {
			commands <- append([]byte(nil), message.Payload()...)
		})
		if !token.WaitTimeout(5*time.Second) || token.Error() != nil {
			t.Fatalf("subscribe to %s: %v", topic, token.Error())
		}

		request, err := http.NewRequest(
			http.MethodPost,
			apiURL+"/command/"+deviceID,
			bytes.NewBufferString(`{"command":1}`),
		)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+apiKey)
		request.Header.Set("Content-Type", "application/json")

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("command request: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusAccepted {
			body, _ := io.ReadAll(response.Body)
			t.Fatalf("command status = %d, body=%s", response.StatusCode, body)
		}

		select {
		case payload := <-commands:
			if string(payload) != `{"command":1}` {
				t.Fatalf("command payload = %s", payload)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("device did not receive command")
		}
	})

	t.Run("MQTT telemetry reaches InfluxDB", func(t *testing.T) {
		recordedAt := time.Now().UTC().Truncate(time.Millisecond)
		payload, err := json.Marshal(map[string]interface{}{
			"version":     1,
			"sensor_id":   deviceID,
			"recorded_at": recordedAt,
			"temperature": 24.75,
			"humidity":    61.5,
			"weight":      120.0,
			"state":       "on",
		})
		if err != nil {
			t.Fatal(err)
		}

		token := device.Publish(models.TelemetryTopic(deviceID), 1, false, payload)
		if !token.WaitTimeout(5*time.Second) || token.Error() != nil {
			t.Fatalf("publish telemetry: %v", token.Error())
		}

		client := influxdb2.NewClient(influxURL, influxToken)
		defer client.Close()
		queryAPI := client.QueryAPI(influxOrg)
		start := strconv.Quote(recordedAt.Add(-time.Second).Format(time.RFC3339Nano))
		query := fmt.Sprintf(
			`from(bucket: %s) |> range(start: time(v: %s)) |> filter(fn: (r) => r._measurement == "sensor_data" and r.sensor_id == %s and r._field == "temperature")`,
			strconv.Quote(influxBucket),
			start,
			strconv.Quote(deviceID),
		)

		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			result, queryErr := queryAPI.Query(context.Background(), query)
			if queryErr == nil {
				found := result.Next()
				result.Close()
				if found {
					return
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		t.Fatal("telemetry point was not found in InfluxDB")
	})
}

func connectDevice(t *testing.T, broker, username, password string) mqtt.Client {
	t.Helper()
	options := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(username + "-integration").
		SetUsername(username).
		SetPassword(password)
	client := mqtt.NewClient(options)
	token := client.Connect()
	if !token.WaitTimeout(10*time.Second) || token.Error() != nil {
		t.Fatalf("connect device to MQTT: %v", token.Error())
	}
	return client
}

func requiredEnv(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("%s is required", name)
	}
	return value
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
