package workers

import (
	"context"
	"fmt"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"robots/pkg/events"
	"sync/atomic"
	"time"
)

type QuiescenceDetectorWorker struct {
	Config        conf.Config
	Name          string
	log           *slog.Logger
	robot         *robot.Robot
	Event         chan events.Event
	droppedEvents uint64
}

func NewQuiescenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, event chan events.Event, droppedEvents uint64) *QuiescenceDetectorWorker {
	return &QuiescenceDetectorWorker{Config: config, log: log, robot: robot, Event: event, droppedEvents: droppedEvents}
}

func (w *QuiescenceDetectorWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w *QuiescenceDetectorWorker) GetName() string {
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

func (w *QuiescenceDetectorWorker) sendQuiescenceDetectorEvent(ctx context.Context, robotID int) {
	select {
	case w.Event <- events.Event{
		EventType: events.EventQuiescenceDetector,
		CreatedAt: time.Now().UTC(),
		Payload: events.QuiescenceDetectorEvent{
			RobotID:      robotID,
			LastActivity: events.LastActivity(w.robot.LastUpdatedAt),
		},
	}:
	case <-ctx.Done():
		w.log.Debug("Context done, stopping event send")
	default:
		atomic.AddUint64(&w.droppedEvents, 1)
		w.log.Warn(fmt.Sprintf("Quiescence event dropped for robot %d, channel full", robotID))
	}
}
