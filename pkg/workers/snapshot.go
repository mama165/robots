package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/errors"
	"robots/pkg/events"
	"robots/pkg/observabilities"
	"time"
)

type SnapshotWorker struct {
	name                events.WorkerName
	config              conf.Config
	log                 *slog.Logger
	telemetryEvent      chan events.Event
	observabilityWorker *observabilities.ObservabilityWorker
}

func NewSnapshotWorker(config conf.Config, log *slog.Logger, telemetryEvent chan events.Event) Worker {
	return SnapshotWorker{config: config, log: log, telemetryEvent: telemetryEvent, observabilityWorker: observabilities.NewObservabilityWorker()}
}

func (s SnapshotWorker) WithName(name string) Worker {
	s.name = events.WorkerName(name)
	return s
}

func (s SnapshotWorker) GetName() events.WorkerName {
	return s.name
}

func (s SnapshotWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.config.SnapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case event := <-s.telemetryEvent:
				s.handleEvent(event)
			default:

			}
		case <-ctx.Done():
			return nil
		default:

		}
	}
}

func (s SnapshotWorker) handleEvent(event events.Event) {
	switch event.EventType {
	case events.EventMessageSent:
		payload, ok := event.Payload.(events.MessageSentEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.IncSent(payload.SenderID.ToInt())
	case events.EventMessageReceived:
		payload, ok := event.Payload.(events.MessageReceivedEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.IncReceived(payload.ReceiverID.ToInt())
	case events.EventMessageDuplicated:
		s.observabilityWorker.IncDuplicated()
	case events.EventMessageReordered:
		s.observabilityWorker.IncReordered()
	case events.EventMessageLost:
		s.observabilityWorker.IncLost()
	case events.EventInvariantViolationSameIndexDiffWords:
		payload, ok := event.Payload.(events.InvariantViolationEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.IncInvariantViolation(payload.ID.ToInt())
	case events.EventQuiescenceDetector:
		payload, ok := event.Payload.(events.QuiescenceDetectorEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.UpdateLastActivity(payload.LastActivity.Date())
	case events.EventWorkerRestartedAfterPanic:
		payload, ok := event.Payload.(events.WorkerRestartedAfterPanicEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.IncWorkerRestart(payload.WorkerName.ToString())
	case events.EventChannelCapacity:
		payload, ok := event.Payload.(events.ChannelCapacityEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observabilityWorker.UpdateCapacity(payload.WorkerName.ToString(), payload.Capacity, payload.Length)
	}
}
