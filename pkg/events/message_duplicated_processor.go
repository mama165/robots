package events

import "log/slog"

type MessageDuplicatedProcessor struct {
	log *slog.Logger
}

func NewMessageDuplicatedProcessor(log *slog.Logger) *MessageDuplicatedProcessor {
	return &MessageDuplicatedProcessor{log: log}
}

func (p MessageDuplicatedProcessor) CanProcess(event Event) bool {
	return event.EventType == EventMessageDuplicated
}

func (p MessageDuplicatedProcessor) Process(event Event) error {
	return nil
}
