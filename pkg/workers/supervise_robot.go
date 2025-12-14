package workers

import (
	"context"
	"fmt"
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
				// Send the winner in the channel without blocking any other possible winner
				select {
				case w.Winner <- *w.Robot:
					w.Log.Info(fmt.Sprintf("Robot %d won", w.Robot.ID))
					return nil
				case <-ctx.Done():
					w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
					return nil
				default:
					// TODO on envoie un event ici aussi ?
					w.Log.Debug(fmt.Sprintf("Robot %d wanted to win but another one won", w.Robot.ID))
				}
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
