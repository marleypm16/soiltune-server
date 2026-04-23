package services

import (
	"soiltune-consumer/api/repository"
	"soiltune-consumer/internal/models"
)

type CommandService struct {
	repository *repository.CommandRepository
}

func NewCommandService(repository *repository.CommandRepository) *CommandService {
	return &CommandService{repository: repository}
}

func (s *CommandService) Execute(sensorID string, command models.Command) error {
	return s.repository.Publish(sensorID, command)
}
