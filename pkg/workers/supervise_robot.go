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

// SuperviseRobotWorker
// For each message merged with word
// Check the secret has been completed
// Only if no update since a chosen duration
type SuperviseRobotWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Robot  *robot.Robot
	Name   string
	Winner chan robot.Robot
	Event  chan events.Event
}

func NewSuperviseRobotWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, winner chan robot.Robot, event chan events.Event) SuperviseRobotWorker {
	return SuperviseRobotWorker{Config: config, Log: log, Robot: robot, Winner: winner, Event: event}
}

func (w SuperviseRobotWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w SuperviseRobotWorker) GetName() string {
	return w.Name
}

func (w SuperviseRobotWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := w.Robot.LastUpdatedAt.Add(w.Config.QuietPeriod).Before(time.Now().UTC())
			if elapsed && w.Robot.IsSecretCompleted(w.Config.EndOfSecret) {
				w.sendSecretWrittenEvent(ctx)
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}

// Send the winner in the channel without blocking any other possible winner
func (w SuperviseRobotWorker) sendSecretWrittenEvent(ctx context.Context) {
	select {
	case w.Event <- events.Event{
		EventType: events.EventSecretWritten,
		CreatedAt: time.Now().UTC(),
		Payload:   events.SecretWrittenEvent{Robot: *w.Robot},
	}:
	case <-ctx.Done():
		w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
	default:
		w.Log.Debug("Buffer is full")
	}
}
