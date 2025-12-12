package workers

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	robotpb "robots/proto/pb-go"
	"time"
)

// UpdateWorker Fetch all missing parts coming from anybody
type UpdateWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	Robot  *robot.Robot
}

func NewUpdateWorker(config conf.Config, logger *slog.Logger, robot *robot.Robot) UpdateWorker {
	return UpdateWorker{Config: config, Log: logger, Robot: robot}
}

func (w UpdateWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w UpdateWorker) GetName() string {
	return w.Name
}

func (w UpdateWorker) Run(ctx context.Context) error {
	for {
		select {
		case updateMsg := <-w.Robot.GossipUpdate:
			var gossipUpdate robotpb.GossipUpdate
			err := proto.Unmarshal(updateMsg, &gossipUpdate)
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			secretParts := robot.FromSecretPartsPb(gossipUpdate.SecretParts)
			for _, secretPart := range secretParts {
				// Updating LastUpdatedAt if the word doesn't exist
				if !robot.ContainsIndex(w.Robot.SecretParts, secretPart.Index) {
					w.Robot.LastUpdatedAt = time.Now().UTC()
					w.Robot.SecretParts = append(w.Robot.SecretParts, secretPart)
					continue
				}
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
