package events

import "log/slog"

type MessageReorderedProcessor struct {
	log *slog.Logger
}

func NewMessageReorderedProcessor(log *slog.Logger) *MessageReorderedProcessor {
	return &MessageReorderedProcessor{log: log}
}

func (p MessageReorderedProcessor) CanProcess(event Event) bool {
	return event.EventType == EventMessageReordered
}

func (p MessageReorderedProcessor) Process(event Event) error {
	return nil
}
