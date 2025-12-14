package events

import "log/slog"

type MessageReceivedProcessor struct {
	log *slog.Logger
}

func NewMessageReceivedProcessor(log *slog.Logger) *MessageReceivedProcessor {
	return &MessageReceivedProcessor{log: log}
}

func (p MessageReceivedProcessor) CanProcess(event Event) bool {
	return event.EventType == EventMessageReceived
}

func (p MessageReceivedProcessor) Process(event Event) error {
	return nil
}
