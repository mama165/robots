package workers

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
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
	Config      conf.Config
	Log         *slog.Logger
	Robot       *robot.Robot
	Name        events.WorkerName
	DomainEvent chan events.Event
}

func NewConvergenceDetectorWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, DomainEvent chan events.Event) ConvergenceDetectorWorker {
	return ConvergenceDetectorWorker{Config: config, Log: log, Robot: robot, DomainEvent: DomainEvent}
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
				w.sendWinnerElectedEvent(ctx, w.Robot.ID)
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping domainEvent send")
			return nil
		}
	}
}

func (w ConvergenceDetectorWorker) sendWinnerElectedEvent(ctx context.Context, id robot.ID) {
	select {
	case w.DomainEvent <- events.Event{
		EventType: events.EventWinnerElected,
		CreatedAt: time.Now().UTC(),
		Payload:   events.WinnerElectedEvent{ID: id.ToInt()},
	}:
	case <-ctx.Done():
		w.Log.Debug("Context done, stopping domainEvent send")
		return
	default:
		w.Log.Debug("ConvergenceDetector channel is full, dropping message")
	}
}
