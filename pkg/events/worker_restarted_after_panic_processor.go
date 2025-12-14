package events

import "log/slog"

type WorkerRestartedAfterPanicProcessor struct {
	log *slog.Logger
}

func NewWorkerRestartedAfterPanicProcessor(log *slog.Logger) *WorkerRestartedAfterPanicProcessor {
	return &WorkerRestartedAfterPanicProcessor{log: log}
}

func (p WorkerRestartedAfterPanicProcessor) CanProcess(event Event) bool {
	return event.EventType == EventWorkerRestartedAfterPanic
}

func (p WorkerRestartedAfterPanicProcessor) Process(event Event) error {
	return nil
}
