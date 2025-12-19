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

type ObservabilityWorker struct {
	name           events.WorkerName
	config         conf.Config
	log            *slog.Logger
	telemetryEvent chan events.Event
	observability  *observabilities.Observability
}

func NewObservabilityWorker(config conf.Config, log *slog.Logger, telemetryEvent chan events.Event) Worker {
	return ObservabilityWorker{config: config, log: log, telemetryEvent: telemetryEvent, observability: observabilities.NewObservability()}
}

func (s ObservabilityWorker) WithName(name string) Worker {
	s.name = events.WorkerName(name)
	return s
}

func (s ObservabilityWorker) GetName() events.WorkerName {
	return s.name
}

func (s ObservabilityWorker) Run(ctx context.Context) error {
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

func (s ObservabilityWorker) handleEvent(event events.Event) {
	switch event.EventType {
	case events.EventMessageSent:
		payload, ok := event.Payload.(events.MessageSentEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.IncSent(payload.SenderID.ToInt())
	case events.EventMessageReceived:
		payload, ok := event.Payload.(events.MessageReceivedEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.IncReceived(payload.ReceiverID.ToInt())
	case events.EventMessageDuplicated:
		s.observability.IncDuplicated()
	case events.EventMessageReordered:
		s.observability.IncReordered()
	case events.EventMessageLost:
		s.observability.IncLost()
	case events.EventInvariantViolationSameIndexDiffWords:
		payload, ok := event.Payload.(events.InvariantViolationEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.IncInvariantViolation(payload.ID.ToInt())
	case events.EventQuiescenceDetector:
		payload, ok := event.Payload.(events.QuiescenceDetectorEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.UpdateLastActivity(payload.LastActivity.Date())
	case events.EventWorkerRestartedAfterPanic:
		payload, ok := event.Payload.(events.WorkerRestartedAfterPanicEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.IncWorkerRestart(payload.WorkerName.ToString())
	case events.EventChannelCapacity:
		payload, ok := event.Payload.(events.ChannelCapacityEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.UpdateCapacity(payload.WorkerName.ToString(), payload.Capacity, payload.Length)
	case events.EventAllConverged:
		payload, ok := event.Payload.(events.AllConvergedEvent)
		if !ok {
			s.log.Error(errors.ErrInvalidPayload.Error())
		}
		s.observability.HasConverged(payload.AllConverged)
	}
}
