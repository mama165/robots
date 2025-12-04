package workers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
)

// WriteSecretWorker Write the secret in a file
type WriteSecretWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	winner chan robot.Robot
}

func (w WriteSecretWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w WriteSecretWorker) GetName() string {
	return w.Name
}

func NewWriteSecretWorker(config conf.Config, log *slog.Logger, winner chan robot.Robot) WriteSecretWorker {
	return WriteSecretWorker{Config: config, Log: log, winner: winner}
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
		}
	}
}
