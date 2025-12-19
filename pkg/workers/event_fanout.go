package workers

import (
	"context"
	"log/slog"
	"robots/pkg/events"
)

type EventFanout struct {
	Log            *slog.Logger
	Name           events.WorkerName
	DomainEvent    chan events.Event
	TelemetryEvent chan events.Event
	handlers       []events.EventHandler
}

func NewEventFanout(log *slog.Logger, domainEvent, telemetryEvent chan events.Event) *EventFanout {
	return &EventFanout{Log: log, DomainEvent: domainEvent, TelemetryEvent: telemetryEvent}
}

func (w EventFanout) Add(handlers ...events.EventHandler) EventFanout {
	w.handlers = append(w.handlers, handlers...)
	return w
}

func (w EventFanout) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w EventFanout) GetName() events.WorkerName { return w.Name }

func (w EventFanout) Run(ctx context.Context) error {
	for {
		select {
		case event := <-w.DomainEvent:
			w.Fanout(event)
			select {
			case w.TelemetryEvent <- event:
			default:
				w.Log.Debug("Observability telemetry event lost")
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping domainEvent send")
			return nil
		}
	}
}

// Fanout One handler for each event
func (w EventFanout) Fanout(event events.Event) {
	for _, p := range w.handlers {
		p.Handle(event)
	}
}
