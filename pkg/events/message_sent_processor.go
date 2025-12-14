package events

import "log/slog"

type MessageSentProcessor struct {
	log *slog.Logger
}

func NewMessageSentProcessor(log *slog.Logger) *MessageSentProcessor {
	return &MessageSentProcessor{log: log}
}

func (p MessageSentProcessor) CanProcess(event Event) bool {
	return event.EventType == EventMessageSent
}

func (p MessageSentProcessor) Process(event Event) error {
	return nil
}
