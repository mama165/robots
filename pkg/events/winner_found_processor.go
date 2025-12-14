package events

import "log/slog"

type WinnerFoundProcessor struct {
	log *slog.Logger
}

func NewWinnerFoundProcessor(log *slog.Logger) *WinnerFoundProcessor {
	return &WinnerFoundProcessor{log: log}
}

func (p WinnerFoundProcessor) CanProcess(event Event) bool {
	return event.EventType == EventWinnerFound
}

func (p WinnerFoundProcessor) Process(event Event) error {
	return nil
}
