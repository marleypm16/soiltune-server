package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

func ensureEnvLoaded() {
	loadEnvOnce.Do(func() {
		_ = godotenv.Overload(".env")
	})
}

type MQTTConfig struct {
	Broker   string
	Topic    string
	ClientID string
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

	if broker == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTBROKER is required")
	}
	if requireTopic && topic == "" {
		return MQTTConfig{}, fmt.Errorf("MQTTTOPIC is required")
	}
	if clientID == "" {
		clientID = "soiltune-client"
	}

	return MQTTConfig{Broker: broker, Topic: topic, ClientID: clientID}, nil
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
		return InfluxConfig{}, fmt.Errorf("DBINFLUXTOKEN is required")
	}
	if org == "" {
		return InfluxConfig{}, fmt.Errorf("DBINFLUXORG is required")
	}
	if bucket == "" {
		return InfluxConfig{}, fmt.Errorf("DBINFLUXBUCKET is required")
	}

	return InfluxConfig{URL: url, Token: token, Org: org, Bucket: bucket}, nil
}
