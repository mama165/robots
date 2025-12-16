package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/supervisor"
	"robots/pkg/events"
	"time"
)

// MetricWorker periodically reports the current channel capacity and length.
// Reading len(channel) and cap(channel) is non-blocking, so this won't interfere
// with other goroutines. It's okay if an event is dropped occasionally because
// metrics are sampled periodically.
type MetricWorker struct {
	config conf.Config
	log    *slog.Logger
	name   string
	event  chan events.Event
}

func NewMetricWorker(config conf.Config, log *slog.Logger, event chan events.Event) MetricWorker {
	return MetricWorker{config: config, log: log, event: event}
}

func (w MetricWorker) WithName(name string) supervisor.Worker {
	w.name = name
	return w
}

func (w MetricWorker) GetName() string {
	return w.name
}

func (w MetricWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.config.MetricInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case w.event <- events.Event{
				EventType: events.EventChannelCapacity,
				CreatedAt: time.Now().UTC(),
				Payload: events.ChannelCapacityEvent{
					WorkerName: w.name,
					Capacity:   cap(w.event),
					Length:     len(w.event),
				},
			}:
			default:
				w.log.Debug("Buffer is full, channel capacity even is lost")
			}
		case <-ctx.Done():
			w.log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
