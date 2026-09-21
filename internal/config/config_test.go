package config

import (
	"testing"
	"time"
)

func TestLoadMQTTConfigDefaultsToQoSOne(t *testing.T) {
	t.Setenv("MQTTBROKER", "tcp://localhost:1883")
	t.Setenv("MQTTTOPIC", "soiltune/telemetry/#")
	t.Setenv("MQTT_USERNAME", "server")
	t.Setenv("MQTT_PASSWORD", "secret")
	t.Setenv("MQTT_QOS", "")

	cfg, err := LoadMQTTConfig("test-client", true)
	if err != nil {
		t.Fatalf("LoadMQTTConfig() error = %v", err)
	}
	if cfg.QoS != 1 {
		t.Fatalf("QoS = %d, want 1", cfg.QoS)
	}
}

func TestLoadMQTTConfigRejectsInvalidQoS(t *testing.T) {
	t.Setenv("MQTTBROKER", "tcp://localhost:1883")
	t.Setenv("MQTTTOPIC", "soiltune/telemetry/#")
	t.Setenv("MQTT_USERNAME", "server")
	t.Setenv("MQTT_PASSWORD", "secret")
	t.Setenv("MQTT_QOS", "3")

	if _, err := LoadMQTTConfig("test-client", true); err == nil {
		t.Fatal("LoadMQTTConfig() error = nil")
	}
}

func TestLoadAPIConfig(t *testing.T) {
	t.Setenv("API_KEY", "12345678901234567890123456789012")
	t.Setenv("API_PORT", "9000")

	cfg, err := LoadAPIConfig()
	if err != nil {
		t.Fatalf("LoadAPIConfig() error = %v", err)
	}
	if cfg.Address != ":9000" {
		t.Fatalf("Address = %q, want :9000", cfg.Address)
	}
}

func TestLoadInfluxConfigWritePolicy(t *testing.T) {
	t.Setenv("DBINFLUX", "http://localhost:8086")
	t.Setenv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN", "token")
	t.Setenv("DOCKER_INFLUXDB_INIT_ORG", "org")
	t.Setenv("DOCKER_INFLUXDB_INIT_BUCKET", "bucket")
	t.Setenv("INFLUX_WRITE_TIMEOUT", "3s")
	t.Setenv("INFLUX_WRITE_ATTEMPTS", "4")

	cfg, err := LoadInfluxConfig()
	if err != nil {
		t.Fatalf("LoadInfluxConfig() error = %v", err)
	}
	if cfg.WriteTimeout != 3*time.Second || cfg.WriteAttempts != 4 {
		t.Fatalf("write policy = %s/%d", cfg.WriteTimeout, cfg.WriteAttempts)
	}
}
