package workers

import (
	"context"
	"log/slog"
	"robots/internal/supervisor"
	"robots/pkg/events"
)

type MetricWorker struct {
	Log        *slog.Logger
	Name       string
	Event      chan events.Event
	processors []events.Processor
}

func NewMetricWorker(log *slog.Logger, event chan events.Event) *MetricWorker {
	return &MetricWorker{Log: log, Event: event}
}

func (w MetricWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w MetricWorker) GetName() string { return w.Name }

func (w MetricWorker) Add(processor ...events.Processor) MetricWorker {
	w.processors = append(w.processors, processor...)
	return w
}

func (w MetricWorker) Run(ctx context.Context) error {
	for {
		select {
		case event := <-w.Event:
			w.Process(event)
			continue
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			continue
		default:
			// TODO à gérer avec un retry maybe ?
			w.Log.Info("Too bad to loose so many events...")
		}
	}
}

// Process Only one processor handle the event
func (w MetricWorker) Process(event events.Event) {
	for _, p := range w.processors {
		if !p.CanProcess(event) {
			continue
		}
		err := p.Process(event)
		if err != nil {
			w.Log.Error(err.Error())
		}
		break
	}
}
