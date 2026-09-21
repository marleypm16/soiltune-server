package routes_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"soiltune-consumer/api/handlers"
	"soiltune-consumer/api/middleware"
	"soiltune-consumer/api/repository"
	"soiltune-consumer/api/routes"
	"soiltune-consumer/api/services"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v3"
)

const integrationAPIKey = "12345678901234567890123456789012"

type fakeMQTTPublisher struct {
	connected bool
	topic     string
	qos       byte
	retained  bool
	payload   interface{}
}

func (fake *fakeMQTTPublisher) IsConnected() bool {
	return fake.connected
}

func (fake *fakeMQTTPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	fake.topic = topic
	fake.qos = qos
	fake.retained = retained
	fake.payload = payload
	return successfulToken{}
}

type successfulToken struct{}

func (successfulToken) Wait() bool                     { return true }
func (successfulToken) WaitTimeout(time.Duration) bool { return true }
func (successfulToken) Error() error                   { return nil }
func (successfulToken) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

func TestAuthenticatedHTTPCommandReachesMQTTPublisher(t *testing.T) {
	publisher := &fakeMQTTPublisher{connected: true}
	repo := repository.NewCommandRepository(publisher, 1)
	service := services.NewCommandService(repo)
	handler := handlers.NewCommandHandler(service)
	app := fiber.New()
	routes.SetupRoutes(app, handler, middleware.RequireAPIKey(integrationAPIKey), func() bool { return true })

	unauthorized := httptest.NewRequest(http.MethodPost, "/command/sensor-01", strings.NewReader(`{"command":1}`))
	unauthorizedResponse, err := app.Test(unauthorized)
	if err != nil {
		t.Fatalf("unauthorized request failed: %v", err)
	}
	defer unauthorizedResponse.Body.Close()
	if unauthorizedResponse.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorizedResponse.StatusCode, fiber.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodPost, "/command/sensor-01", strings.NewReader(`{"command":1}`))
	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+integrationAPIKey)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("authorized request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusAccepted {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want %d; body=%s", response.StatusCode, fiber.StatusAccepted, body)
	}
	if publisher.topic != "soiltune/commands/sensor-01" {
		t.Fatalf("topic = %q", publisher.topic)
	}
	if publisher.qos != 1 || publisher.retained {
		t.Fatalf("publish options: qos=%d retained=%v", publisher.qos, publisher.retained)
	}
	if string(publisher.payload.([]byte)) != `{"command":1}` {
		t.Fatalf("payload = %s", publisher.payload)
	}
}

func TestReadinessReflectsMQTTConnection(t *testing.T) {
	publisher := &fakeMQTTPublisher{connected: false}
	repo := repository.NewCommandRepository(publisher, 1)
	app := fiber.New()
	routes.SetupRoutes(
		app,
		handlers.NewCommandHandler(services.NewCommandService(repo)),
		middleware.RequireAPIKey(integrationAPIKey),
		publisher.IsConnected,
	)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if err != nil {
		t.Fatalf("readiness request failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusServiceUnavailable)
	}
}
