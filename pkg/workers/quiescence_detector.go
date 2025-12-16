package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"robots/pkg/events"
	"time"
)

type QuiescenceDetectorWorker struct {
	Config conf.Config
	Name   string
	log    *slog.Logger
	robot  *robot.Robot
	Event  chan events.Event
}

func NewQuiescenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, event chan events.Event) *QuiescenceDetectorWorker {
	return &QuiescenceDetectorWorker{Config: config, log: log, robot: robot, Event: event}
}

func (w QuiescenceDetectorWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w QuiescenceDetectorWorker) GetName() string {
	return w.Name
}

func (w QuiescenceDetectorWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Config.MetricInterval)
	for {
		select {
		case <-ticker.C:
			w.sendQuiescenceDetectorEvent(ctx, w.robot.ID)
			return nil
		case <-ctx.Done():
			w.log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		default:
			w.log.Debug("Buffer is full")
		}
		return nil
	}
}

func (w QuiescenceDetectorWorker) sendQuiescenceDetectorEvent(ctx context.Context, robotID int) {
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
		w.log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
	default:
		w.log.Debug("Buffer is full")
	}
}
