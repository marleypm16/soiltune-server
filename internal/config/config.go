package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"soiltune-consumer/internal/models"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

func ensureEnvLoaded() {
	loadEnvOnce.Do(func() {
		_ = godotenv.Load(".env")
	})
}

type MQTTConfig struct {
	Broker   string
	Topic    string
	ClientID string
	Username string
	Password string
	QoS      byte
}

type APIConfig struct {
	Key     string
	Address string
}

type InfluxConfig struct {
	URL           string
	Token         string
	Org           string
	Bucket        string
	WriteTimeout  time.Duration
	WriteAttempts int
}

func LoadMQTTConfig(clientID string, requireTopic bool) (MQTTConfig, error) {
	ensureEnvLoaded()

	broker := os.Getenv("MQTTBROKER")
	topic := os.Getenv("MQTTTOPIC")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	qosValue := envOrDefault("MQTT_QOS", "1")

	if broker == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTBROKER is required")
	}
	if requireTopic && topic == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTTOPIC is required")
	}
	if requireTopic && topic != models.TelemetryTopicFilter {
		return MQTTConfig{}, fmt.Errorf("MQTTTOPIC must be %q", models.TelemetryTopicFilter)
	}
	if username == "" {
		return MQTTConfig{}, fmt.Errorf("MQTT_USERNAME is required")
	}
	if password == "" {
		return MQTTConfig{}, fmt.Errorf("MQTT_PASSWORD is required")
	}
	qos, err := strconv.ParseUint(qosValue, 10, 8)
	if err != nil || qos > 2 {
		return MQTTConfig{}, fmt.Errorf("MQTT_QOS must be 0, 1 or 2")
	}
	if clientID == "" {
		clientID = "soiltune-client"
	}

	return MQTTConfig{
		Broker:   broker,
		Topic:    topic,
		ClientID: clientID,
		Username: username,
		Password: password,
		QoS:      byte(qos),
	}, nil
}

func LoadAPIConfig() (APIConfig, error) {
	ensureEnvLoaded()

	key := strings.TrimSpace(os.Getenv("API_KEY"))
	if len(key) < 32 {
		return APIConfig{}, fmt.Errorf("API_KEY must contain at least 32 characters")
	}

	portValue := envOrDefault("API_PORT", "8000")
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return APIConfig{}, fmt.Errorf("API_PORT must be between 1 and 65535")
	}

	return APIConfig{Key: key, Address: fmt.Sprintf(":%d", port)}, nil
}

func LoadInfluxConfig() (InfluxConfig, error) {
	ensureEnvLoaded()

	url := os.Getenv("DBINFLUX")
	token := os.Getenv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN")
	org := os.Getenv("DOCKER_INFLUXDB_INIT_ORG")
	bucket := os.Getenv("DOCKER_INFLUXDB_INIT_BUCKET")
	writeTimeoutValue := envOrDefault("INFLUX_WRITE_TIMEOUT", "5s")
	writeAttemptsValue := envOrDefault("INFLUX_WRITE_ATTEMPTS", "3")

	if url == "" {
		return InfluxConfig{}, fmt.Errorf("DBINFLUX is required")
	}
	if token == "" {
		return InfluxConfig{}, fmt.Errorf("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN is required")
	}
	if org == "" {
		return InfluxConfig{}, fmt.Errorf("DOCKER_INFLUXDB_INIT_ORG is required")
	}
	if bucket == "" {
		return InfluxConfig{}, fmt.Errorf("DOCKER_INFLUXDB_INIT_BUCKET is required")
	}
	writeTimeout, err := time.ParseDuration(writeTimeoutValue)
	if err != nil || writeTimeout <= 0 {
		return InfluxConfig{}, fmt.Errorf("INFLUX_WRITE_TIMEOUT must be a positive duration")
	}
	writeAttempts, err := strconv.Atoi(writeAttemptsValue)
	if err != nil || writeAttempts < 1 || writeAttempts > 10 {
		return InfluxConfig{}, fmt.Errorf("INFLUX_WRITE_ATTEMPTS must be between 1 and 10")
	}

	return InfluxConfig{
		URL:           url,
		Token:         token,
		Org:           org,
		Bucket:        bucket,
		WriteTimeout:  writeTimeout,
		WriteAttempts: writeAttempts,
	}, nil
}

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}
