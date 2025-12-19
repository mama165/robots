package workers

import (
	"context"
	"log/slog"
	"robots/pkg/events"
)

type Dispatcher struct {
	Log             *slog.Logger
	Name            events.WorkerName
	CriticalEvent   chan events.Event
	ObservableEvent chan events.Event
	processors      []events.Processor
}

func NewDispatcher(log *slog.Logger, criticalEvent, observableEvent chan events.Event) *Dispatcher {
	return &Dispatcher{Log: log, CriticalEvent: criticalEvent, ObservableEvent: observableEvent}
}

func (w Dispatcher) Add(processor ...events.Processor) Dispatcher {
	w.processors = append(w.processors, processor...)
	return w
}

func (w Dispatcher) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w Dispatcher) GetName() events.WorkerName { return w.Name }

func (w Dispatcher) Run(ctx context.Context) error {
	for {
		select {
		case event := <-w.CriticalEvent:
			w.Dispatch(event)
			select {
			case w.ObservableEvent <- event:
			default:
				w.Log.Debug("Observability event lost")
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping event send")
			return nil
		}
	}
}

// Dispatch Only one processor handle the event
func (w Dispatcher) Dispatch(event events.Event) {
	for _, p := range w.processors {
		p.Handle(event)
	}
}
