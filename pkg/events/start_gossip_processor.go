package events

import (
	"log/slog"
)

type StartGossipProcessor struct {
	log *slog.Logger
}

func NewStartGossipProcessor(log *slog.Logger) *StartGossipProcessor {
	return &StartGossipProcessor{log: log}
}

func (p StartGossipProcessor) CanProcess(event Event) bool {
	return event.EventType == EventStartGossip
}

func (p StartGossipProcessor) Handle(event Event) error {
	return nil
}
