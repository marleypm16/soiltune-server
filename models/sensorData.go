package models

type SensorData struct {
	SensorID    string  `json:"sensor_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Weight      float64 `json:"weight"`
}
