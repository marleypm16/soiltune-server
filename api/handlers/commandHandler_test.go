package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"soiltune-consumer/internal/apperrors"
	"soiltune-consumer/internal/models"

	"github.com/gofiber/fiber/v3"
)

type fakeCommandExecutor struct {
	err      error
	sensorID string
	command  models.Command
}

func (fake *fakeCommandExecutor) Execute(sensorID string, command models.Command) error {
	fake.sensorID = sensorID
	fake.command = command
	return fake.err
}

func TestCommandHandlerValidation(t *testing.T) {
	tests := []struct {
		name       string
		sensorID   string
		body       string
		wantStatus int
	}{
		{name: "valid", sensorID: "sensor-01", body: `{"command":1}`, wantStatus: fiber.StatusAccepted},
		{name: "invalid sensor id", sensorID: "-sensor", body: `{"command":1}`, wantStatus: fiber.StatusBadRequest},
		{name: "unknown field", sensorID: "sensor-01", body: `{"command":1,"other":true}`, wantStatus: fiber.StatusBadRequest},
		{name: "invalid command", sensorID: "sensor-01", body: `{"command":2}`, wantStatus: fiber.StatusBadRequest},
		{name: "multiple objects", sensorID: "sensor-01", body: `{"command":1}{"command":0}`, wantStatus: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &fakeCommandExecutor{}
			app := fiber.New()
			app.Post("/command/:sensorId", NewCommandHandler(executor).Handle)

			req := httptest.NewRequest(http.MethodPost, "/command/"+tt.sensorID, strings.NewReader(tt.body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestCommandHandlerMapsUnavailableDependency(t *testing.T) {
	executor := &fakeCommandExecutor{err: fmtUnavailable()}
	app := fiber.New()
	app.Post("/command/:sensorId", NewCommandHandler(executor).Handle)

	req := httptest.NewRequest(http.MethodPost, "/command/sensor-01", strings.NewReader(`{"command":0}`))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
	}
}

func fmtUnavailable() error {
	return errors.Join(errors.New("publish failed"), apperrors.ErrUnavailable)
}
