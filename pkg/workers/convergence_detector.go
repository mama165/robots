package workers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	"sync"
	"time"
)

// ConvergenceDetectorWorker continuously monitors a robot's progress to detect when the secret
// has been fully reconstructed. Once a robot has completed the secret and no updates have been
// received for a configured quiet period, it attempts to write the secret to the output file.
//
// Only one robot should write the secret. A channel (Winner) is used to coordinate this, but
// because multiple robots can attempt to write concurrently, the first robot to succeed becomes
// the "winner". Other robots detect that a winner already exists and do not write.
//
// ⚠️ Important concurrency note:
// If the Winner channel is unbuffered or incorrectly handled, robots can fall into the default
// case and fail to write the secret. This can cause no file to be created even though a robot
// has completed the secret. Using a buffered channel or a sync.Once ensures that the secret
// is reliably written exactly once.
type ConvergenceDetectorWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Robot  *robot.Robot
	Name   events.WorkerName
	Winner chan *robot.Robot
	once   *sync.Once
	writer io.Writer
}

func NewConvergenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, winner chan *robot.Robot, once *sync.Once, writer io.Writer) ConvergenceDetectorWorker {
	if writer == nil {
		panic("ConvergenceDetectorWorker: writer must be injected")
	}
	return ConvergenceDetectorWorker{Config: config, Log: log, Robot: robot, Winner: winner, once: once, writer: writer}
}

func (w ConvergenceDetectorWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w ConvergenceDetectorWorker) GetName() events.WorkerName {
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
			w.Log.Debug("Context done, stopping domainEvent send")
			return nil
		}
	}
}

// tryWriteSecret Send the robot in the channel
func (w ConvergenceDetectorWorker) tryWriteSecret() {
	// Tenter d'écrire le secret exactement une fois
	w.once.Do(func() {
		w.writeSecret()
		// Send winner in channel to notify others
		select {
		case w.Winner <- w.Robot:
		default:
			// jamais bloquant
		}
	})
	// Notify other robots they lost
	select {
	case winner := <-w.Winner:
		if winner.ID != w.Robot.ID {
			w.Log.Info(fmt.Sprintf(
				"Robot %d tried to win but robot %d has already written the secret",
				w.Robot.ID, winner.ID,
			))
		}
		// ⚠️Winner is sent again in channel for other robots to read it
		select {
		case w.Winner <- winner:
		default:
		}
	default:
		// le channel peut être vide si aucun gagnant n'a encore été mis
		w.Log.Info(fmt.Sprintf("Robot %d tried to win but no winner registered yet", w.Robot.ID))
	}
}

// Send the winner in the channel without blocking any other possible winner
func (w ConvergenceDetectorWorker) writeSecret() {
	if _, err := w.writer.Write([]byte(w.Robot.BuildSecret())); err != nil {
		w.Log.Error("failed to write secret", "err", err)
	}
	w.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", w.Robot.ID, w.Config.OutputFile))
}
