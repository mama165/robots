package workers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"robots/internal/conf"
	"robots/internal/robot"
)

// WriteSecretWorker Write the secret in a file
type WriteSecretWorker struct {
	Config conf.Config
	Log    *slog.Logger
	winner chan robot.Robot
}

func NewWriteSecretWorker(config conf.Config, log *slog.Logger, winner chan robot.Robot) WriteSecretWorker {
	return WriteSecretWorker{Config: config, Log: log, winner: winner}
}

func (w WriteSecretWorker) Run(ctx context.Context) error {
	for {
		select {
		case robot := <-w.winner:
			file, err := os.Create(w.Config.OutputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if _, err = file.WriteString(robot.BuildSecret()); err != nil {
				panic(err)
			}
			w.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", robot.ID, w.Config.OutputFile))
			return nil
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}
