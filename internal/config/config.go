package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
}

type APIConfig struct {
	Key string
}

type InfluxConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

func LoadMQTTConfig(clientID string, requireTopic bool) (MQTTConfig, error) {
	ensureEnvLoaded()

	broker := os.Getenv("MQTTBROKER")
	topic := os.Getenv("MQTTTOPIC")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if broker == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTBROKER is required")
	}
	if requireTopic && topic == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTTOPIC is required")
	}
	if username == "" {
		return MQTTConfig{}, fmt.Errorf("MQTT_USERNAME is required")
	}
	if password == "" {
		return MQTTConfig{}, fmt.Errorf("MQTT_PASSWORD is required")
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
	}, nil
}

func LoadAPIConfig() (APIConfig, error) {
	ensureEnvLoaded()

	key := strings.TrimSpace(os.Getenv("API_KEY"))
	if len(key) < 32 {
		return APIConfig{}, fmt.Errorf("API_KEY must contain at least 32 characters")
	}

	return APIConfig{Key: key}, nil
}

func LoadInfluxConfig() (InfluxConfig, error) {
	ensureEnvLoaded()

	url := os.Getenv("DBINFLUX")
	token := os.Getenv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN")
	org := os.Getenv("DOCKER_INFLUXDB_INIT_ORG")
	bucket := os.Getenv("DOCKER_INFLUXDB_INIT_BUCKET")

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

	return InfluxConfig{URL: url, Token: token, Org: org, Bucket: bucket}, nil
}
