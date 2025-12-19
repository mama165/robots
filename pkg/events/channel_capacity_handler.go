package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
)

// ChannelCapacityHandler turns raw channel capacity metrics into warnings.
// It logs a warning when the remaining buffer size falls below a defined threshold.
type ChannelCapacityHandler struct {
	log                  *slog.Logger
	LowCapacityThreshold int
}

func NewChannelCapacityHandler(log *slog.Logger, LowCapacityThreshold int) *ChannelCapacityHandler {
	return &ChannelCapacityHandler{log: log, LowCapacityThreshold: LowCapacityThreshold}
}

func (p ChannelCapacityHandler) Handle(event Event) {
	switch event.EventType {
	case EventChannelCapacity:
		payload, ok := event.Payload.(ChannelCapacityEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
			return
		}
		p.log.Debug(fmt.Sprintf("Channel usage: %d / %d", payload.Length, payload.Capacity))
		if payload.Capacity <= 0 { // In case of unbuffered channel
			return
		}
		capacityLeft := payload.Capacity - payload.Length
		if capacityLeft > 0 && capacityLeft <= p.LowCapacityThreshold {
			p.log.Warn(fmt.Sprintf("metric event channel capacity left : %d", capacityLeft))
		}
	}
}
