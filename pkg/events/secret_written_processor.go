package events

import "log/slog"

type SecretWrittenProcessor struct {
	log *slog.Logger
}

func NewSecretWrittenProcessor(log *slog.Logger) *SecretWrittenProcessor {
	return &SecretWrittenProcessor{log: log}
}

func (p SecretWrittenProcessor) Handle(event Event) error {
	return nil
}
