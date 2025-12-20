package workers

import (
	"context"
	"fmt"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	"sync/atomic"
	"time"
)

// QuiescenceDetectorWorker observes system activity and detects periods of
// quiescence (i.e. absence of meaningful state changes).
// It does not infer correctness or convergence by itself. Quiescence is used
// as a temporal signal to support higher-level decisions such as convergence
// completion or termination.
// Quiescence detection is heuristic and may trigger multiple times; consumers
// must treat its signals as advisory and idempotent.
type QuiescenceDetectorWorker struct {
	Config        conf.Config
	Name          events.WorkerName
	log           *slog.Logger
	robot         *robot.Robot
	DomainEvent   chan events.Event
	droppedEvents uint64
}

func NewQuiescenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, domainEvent chan events.Event, droppedEvents uint64) *QuiescenceDetectorWorker {
	return &QuiescenceDetectorWorker{Config: config, log: log, robot: robot, DomainEvent: domainEvent, droppedEvents: droppedEvents}
}

func (w *QuiescenceDetectorWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w *QuiescenceDetectorWorker) GetName() events.WorkerName {
	return w.Name
}

func (w *QuiescenceDetectorWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Config.MetricInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.sendQuiescenceDetectorEvent(ctx, w.robot.ID)
		case <-ctx.Done():
			w.log.Info("Stopping quiescence detector")
			return nil
		}
	}
}

func (w *QuiescenceDetectorWorker) sendQuiescenceDetectorEvent(ctx context.Context, ID robot.ID) {
	select {
	case w.DomainEvent <- events.Event{
		EventType: events.EventQuiescenceDetector,
		CreatedAt: time.Now().UTC(),
		Payload: events.QuiescenceDetectorEvent{
			ID:           ID.ToInt(),
			LastActivity: events.LastActivity(w.robot.LastUpdatedAt),
		},
	}:
	case <-ctx.Done():
		w.log.Debug("Context done, stopping domainEvent send")
	default:
		atomic.AddUint64(&w.droppedEvents, 1)
		w.log.Warn(fmt.Sprintf("Quiescence domainEvent dropped for robot %d, channel full", ID))
	}
}
