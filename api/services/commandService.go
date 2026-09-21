package services

import "soiltune-consumer/internal/models"

type CommandPublisher interface {
	Publish(sensorID string, command models.Command) error
}

type CommandService struct {
	publisher CommandPublisher
}

func NewCommandService(publisher CommandPublisher) *CommandService {
	return &CommandService{publisher: publisher}
}

func (s *CommandService) Execute(sensorID string, command models.Command) error {
	return s.publisher.Publish(sensorID, command)
}
