package events

import "log/slog"

type SupervisorStartedProcessor struct {
	log *slog.Logger
}

func NewSupervisorStartedProcessor(log *slog.Logger) *SupervisorStartedProcessor {
	return &SupervisorStartedProcessor{log: log}
}

func (p SupervisorStartedProcessor) CanProcess(event Event) bool {
	return event.EventType == EventSupervisorStarted
}

func (p SupervisorStartedProcessor) Process(event Event) error {
	return nil
}
