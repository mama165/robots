package events

import "log/slog"

type SecretWrittenProcessor struct {
	log *slog.Logger
}

func NewSecretWrittenProcessor(log *slog.Logger) *SecretWrittenProcessor {
	return &SecretWrittenProcessor{log: log}
}

func (p SecretWrittenProcessor) CanProcess(event Event) bool {
	return event.EventType == EventSecretWritten
}

func (p SecretWrittenProcessor) Process(event Event) error {
	return nil
}
