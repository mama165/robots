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

type UpdateWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	robot  *robot.Robot
}

func NewUpdateWorker(config conf.Config, logger *slog.Logger, robot *robot.Robot) UpdateWorker {
	return UpdateWorker{Config: config, Log: logger, robot: robot}
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
		case updateMsg := <-w.robot.GossipUpdate:
			// On récupère les parties manquantes venant de n'importe qui
			// On sait qu'on a récupéré uniquement les parties manquantes car c'est du gossip push-pull
			var gossipUpdate robotpb.GossipUpdate
			err := proto.Unmarshal(updateMsg, &gossipUpdate)
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			// TODO il faudra transformer robot.SecretParts en chan []byte pour écrire dedans
			//TODO Doit-on contiuer à vérifier si on contient déjà le mot ?
			secretParts := robot.FromSecretPartsPb(gossipUpdate.SecretParts)
			for _, secretPart := range secretParts {
				// Updating LastUpdatedAt if the word doesn't exist
				if !robot.ContainsIndex(w.robot.SecretParts, secretPart.Index) {
					w.robot.LastUpdatedAt = time.Now().UTC()
					w.robot.SecretParts = append(w.robot.SecretParts, secretPart)
					continue
				}
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}
