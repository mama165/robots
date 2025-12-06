package workers

import (
	"context"
	"fmt"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"time"
)

// SuperviseRobotWorker
// For each message merged with word
// Check the secret has been completed
// Only if no update since a chosen duration
type SuperviseRobotWorker struct {
	Config conf.Config
	Log    *slog.Logger
	robot  *robot.Robot
	Name   string
	winner chan robot.Robot
}

func NewSuperviseRobotWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, winner chan robot.Robot) SuperviseRobotWorker {
	return SuperviseRobotWorker{Config: config, Log: log, robot: robot, winner: winner}
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
			elapsed := w.robot.LastUpdatedAt.Add(w.Config.QuietPeriod).Before(time.Now().UTC())
			if elapsed && w.robot.IsSecretCompleted(w.Config.EndOfSecret) {
				// Send the winner in the channel without blocking any other possible winner
				select {
				case w.winner <- *w.robot:
					w.Log.Info(fmt.Sprintf("Robot %d won", w.robot.ID))
					return nil
				case <-ctx.Done():
					w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
					return nil
				default:
					w.Log.Debug(fmt.Sprintf("Robot %d wanted to win but another one won", w.robot.ID))
				}
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
