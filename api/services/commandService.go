package services

import (
	"soiltune-consumer/api/repository"
)

func CommandService(sensorID string, comando []byte) error {

	return repository.CommandRepository(sensorID, comando)
}
