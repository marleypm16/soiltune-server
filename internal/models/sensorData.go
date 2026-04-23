package models

type SensorData struct {
	SensorID    string  `json:"sensor_id"`
	Temperature float64 `json:"temperatura"`
	Humidity    float64 `json:"umidade"`
	Weight      float64 `json:"peso"`
	State       string  `json:"estado"`
}
