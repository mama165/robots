package workers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"robots/pkg/events"
	"time"
)

// WriteSecretWorker Write the secret in a file
type WriteSecretWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	winner chan robot.Robot
	Event  chan events.Event
}

func NewWriteSecretWorker(config conf.Config, log *slog.Logger, winner chan robot.Robot, event chan events.Event) WriteSecretWorker {
	return WriteSecretWorker{Config: config, Log: log, winner: winner, Event: event}
}

func (w WriteSecretWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w WriteSecretWorker) GetName() string {
	return w.Name
}

func (w WriteSecretWorker) Run(ctx context.Context) error {
	for {
		select {
		case rb := <-w.winner:
			file, err := os.Create(w.Config.OutputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if _, err = file.WriteString(rb.BuildSecret()); err != nil {
				panic(err)
			}
			w.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", rb.ID, w.Config.OutputFile))
			return nil
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		default:
			w.Log.Debug("WriteSecret channel is full, dropping message")
		}
	}
}

func (w WriteSecretWorker) sendSecretWrittenEvent(ctx context.Context) {
	select {
	case w.Event <- events.Event{
		EventType: events.EventSecretWritten,
		CreatedAt: time.Now().UTC(),
		Payload:   nil,
	}:
	case <-ctx.Done():
		w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
	default:
		w.Log.Debug("Buffer is full")
	}
}
