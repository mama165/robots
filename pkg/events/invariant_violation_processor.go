package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
)

type InvariantViolationProcessor struct {
	log *slog.Logger
}

func NewInvariantViolationProcessor(log *slog.Logger) *InvariantViolationProcessor {
	return &InvariantViolationProcessor{log: log}
}

func (p InvariantViolationProcessor) CanProcess(event Event) bool {
	return event.EventType == EventInvariantViolation
}

func (p InvariantViolationProcessor) Process(event Event) error {
	payload, ok := event.Payload.(InvariantViolationEvent)
	if !ok {
		return errors.ErrInvalidPayload
	}
	p.log.Debug(fmt.Sprintf("%s buffer is full", payload.WorkerName))
	return nil
}
