package handlers

import (
    "context"
    "fmt"
    "strconv"
    "log"

    "github.com/gofiber/fiber/v3"
    "soiltune-consumer/api/services"
)

type SensorHandler struct {
    influx *services.InfluxService
}

func NewSensorHandler(influx *services.InfluxService) *SensorHandler {
    return &SensorHandler{influx: influx}
}

func (h *SensorHandler) GetSensorIDs(c fiber.Ctx) error {
    ctx := context.Background()
    flux := fmt.Sprintf(`from(bucket: "%s")
  |> range(start: -30d)
  |> filter(fn: (r) => r._measurement == "sensor_data")
  |> keep(columns: ["sensor_id"]) 
  |> distinct(column: "sensor_id")`, h.influx.Bucket())

    rows, err := h.influx.Query(ctx, flux)
    if err != nil {
        log.Printf("influx query error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
    }

    ids := make([]string, 0, len(rows))
    for _, r := range rows {
        if v, ok := r["sensor_id"]; ok {
            if s, ok := v.(string); ok {
                ids = append(ids, s)
            }
        }
    }

    return c.JSON(ids)
}

// GetSensorData returns the last N data points for a sensor.
// Query params:
// - limit (int, default 100)
// - range (Flux duration string, default 24h)
func (h *SensorHandler) GetSensorData(c fiber.Ctx) error {
    sensorID := c.Params("id")
    if sensorID == "" {
        return c.Status(fiber.StatusBadRequest).SendString("missing sensor id")
    }

    limitParam := c.Query("limit", "100")
    limit, err := strconv.Atoi(limitParam)
    if err != nil || limit <= 0 {
        limit = 100
    }

    rangeParam := c.Query("range", "24h")

    ctx := context.Background()
    flux := fmt.Sprintf(`from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r._measurement == "sensor_data" and r.sensor_id == "%s")
  |> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn: "_value")
  |> sort(columns:["_time"], desc: true)
  |> limit(n:%d)
  |> sort(columns:["_time"], desc: false)
`, h.influx.Bucket(), rangeParam, sensorID, limit)

    rows, err := h.influx.Query(ctx, flux)
    if err != nil {
        log.Printf("influx query error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
    }

    return c.JSON(rows)
}
