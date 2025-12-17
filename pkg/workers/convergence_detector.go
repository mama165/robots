package workers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"time"
)

// ConvergenceDetectorWorker
// For each message merged with word
// Check the secret has been completed
// Only if no update since a chosen duration
type ConvergenceDetectorWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Robot  *robot.Robot
	Name   string
	Winner chan robot.Robot
}

func NewConvergenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, winner chan robot.Robot) ConvergenceDetectorWorker {
	return ConvergenceDetectorWorker{Config: config, Log: log, Robot: robot, Winner: winner}
}

func (w ConvergenceDetectorWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w ConvergenceDetectorWorker) GetName() string {
	return w.Name
}

func (w ConvergenceDetectorWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			elapsed := w.Robot.LastUpdatedAt.Add(w.Config.QuietPeriod).Before(time.Now().UTC())
			if elapsed && w.Robot.IsSecretCompleted(w.Config.EndOfSecret) {
				w.tryWriteSecret()
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}

// tryWriteSecret Send the robot in the channel
func (w ConvergenceDetectorWorker) tryWriteSecret() {
	select {
	case w.Winner <- *w.Robot:
		w.writeSecret()
	default:
		w.Log.Debug(fmt.Sprintf("Robot %d could have won but an other one has already written the secret", w.Robot.ID))
	}
}

// Send the winner in the channel without blocking any other possible winner
func (w ConvergenceDetectorWorker) writeSecret() {
	file, err := os.Create(w.Config.OutputFile)
	if err != nil {
		w.Log.Error(err.Error())
	}
	defer file.Close()
	if _, err = file.WriteString(w.Robot.BuildSecret()); err != nil {
		w.Log.Error(err.Error())
	}
	w.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", w.Robot.ID, w.Config.OutputFile))
}
