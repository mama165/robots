package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	"time"
)

// ConvergenceObserverWorker observes robot states and emits a signal when
// convergence is reached.
// It does not participate in gossip, state mutation, or reconciliation.
// Its sole responsibility is to detect convergence based on observed state
// and notify external observers (UI, logs, metrics, tests).
// Convergence detection may be triggered multiple times; emitted effects
// must therefore be idempotent.
// This worker is intended as an observability boundary and should not be
// relied upon by core domain logic.
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
