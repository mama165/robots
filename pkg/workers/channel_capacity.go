package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"time"
)

// ChannelCapacityWorker periodically reports the current channel capacity and length.
// Reading len(channel) and cap(channel) is non-blocking, so this won't interfere
// with other goroutines. It's okay if an event is dropped occasionally because
// metrics are sampled periodically.
type ChannelCapacityWorker struct {
	config conf.Config
	log    *slog.Logger
	name   events.WorkerName
	event  chan events.Event
}

func NewChannelCapacityWorker(config conf.Config, log *slog.Logger, event chan events.Event) ChannelCapacityWorker {
	return ChannelCapacityWorker{config: config, log: log, event: event}
}

func (w ChannelCapacityWorker) WithName(name string) Worker {
	w.name = events.WorkerName(name)
	return w
}

func (w ChannelCapacityWorker) GetName() events.WorkerName {
	return w.name
}

func (w ChannelCapacityWorker) Run(ctx context.Context) error {
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
			w.log.Debug("Context done, stopping event send")
			return nil
		}
	}
}
