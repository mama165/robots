package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	"time"
)

type ConvergenceObserverWorker struct {
	config      conf.Config
	Log         *slog.Logger
	Name        events.WorkerName
	Robots      []*robot.Robot
	DomainEvent chan events.Event
}

func NewConvergenceObserverWorker(config conf.Config, log *slog.Logger, robots []*robot.Robot, domainEvent chan events.Event) Worker {
	return ConvergenceObserverWorker{
		config:      config,
		Log:         log,
		Robots:      robots,
		DomainEvent: domainEvent,
	}
}

func (w ConvergenceObserverWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w ConvergenceObserverWorker) GetName() events.WorkerName {
	return w.Name
}

func (w ConvergenceObserverWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			allConverged := true
			for _, r := range w.Robots {
				if !r.IsSecretCompleted(w.config.EndOfSecret) {
					allConverged = false
				}
			}
			select {
			case w.DomainEvent <- events.Event{
				EventType: events.EventAllConverged,
				CreatedAt: time.Now().UTC(),
				Payload:   events.AllConvergedEvent{AllConverged: allConverged},
			}:
			case <-ctx.Done():
				w.Log.Debug("Context done, stopping domainEvent send")
			default:
				w.Log.Debug("ConvergenceObserver channel is full, dropping message")
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping domainEvent send")
			return nil
		}
	}
}
