package workers

import (
	"context"
	"log/slog"
	"robots/pkg/events"
)

type Dispatcher struct {
	Log        *slog.Logger
	Name       events.WorkerName
	Event      chan events.Event
	processors []events.Processor
}

func NewDispatcher(log *slog.Logger, event chan events.Event) *Dispatcher {
	return &Dispatcher{Log: log, Event: event}
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
		case event := <-w.Event:
			w.Dispatch(event)
			continue
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
