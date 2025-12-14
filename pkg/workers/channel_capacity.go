package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/supervisor"
	"robots/pkg/events"
	"time"
)

// ChannelCapacityWorker periodically reports the current channel capacity and length.
// Reading len(channel) and cap(channel) is non-blocking, so this won't interfere
// with other goroutines. It's okay if an event is dropped occasionally because
// metrics are sampled periodically.
type ChannelCapacityWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	Event  chan events.Event
}

func NewChannelCapacityWorker(config conf.Config, log *slog.Logger, event chan events.Event) ChannelCapacityWorker {
	return ChannelCapacityWorker{Config: config, Log: log, Event: event}
}

func (w ChannelCapacityWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w ChannelCapacityWorker) GetName() string {
	return w.Name
}

func (w ChannelCapacityWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Config.MetricInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case w.Event <- events.Event{
				EventType: events.EventChannelCapacity,
				CreatedAt: time.Now().UTC(),
				Payload: events.ChannelCapacityEvent{
					WorkerName: w.Name,
					Capacity:   cap(w.Event),
					Length:     len(w.Event),
				},
			}:
			default:
				w.Log.Debug("Buffer is full, channel capacity even is lost")
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
