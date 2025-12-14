package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
)

// ChannelCapacityProcessor turns raw channel capacity metrics into warnings.
// It logs a warning when the remaining buffer size falls below a defined threshold.
type ChannelCapacityProcessor struct {
	log                  *slog.Logger
	LowCapacityThreshold int
}

func NewChannelCapacityProcessor(log *slog.Logger, LowCapacityThreshold int) *ChannelCapacityProcessor {
	return &ChannelCapacityProcessor{log: log, LowCapacityThreshold: LowCapacityThreshold}
}

func (p ChannelCapacityProcessor) CanProcess(event Event) bool {
	return event.EventType == EventChannelCapacity
}

func (p ChannelCapacityProcessor) Process(event Event) error {
	payload, ok := event.Payload.(ChannelCapacityEvent)
	if !ok {
		return errors.ErrInvalidPayload
	}
	p.log.Debug(fmt.Sprintf("Channel usage: %d / %d", payload.Length, payload.Capacity))
	if payload.Capacity <= 0 { // In case of unbuffered channel
		return nil
	}
	capacityLeft := payload.Capacity - payload.Length
	if capacityLeft > 0 && capacityLeft <= p.LowCapacityThreshold {
		p.log.Warn(fmt.Sprintf("metric event channel capacity left : %d", capacityLeft))
	}
	return nil
}
