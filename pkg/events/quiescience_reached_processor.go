package events

import "log/slog"

type QuiescenceReachedProcessor struct {
	log *slog.Logger
}

func NewQuiescenceReachedProcessor(log *slog.Logger) *QuiescenceReachedProcessor {
	return &QuiescenceReachedProcessor{log: log}
}

func (p QuiescenceReachedProcessor) CanProcess(event Event) bool {
	return event.EventType == EventQuiescenceReached
}

func (p QuiescenceReachedProcessor) Process(event Event) error {
	return nil
}
