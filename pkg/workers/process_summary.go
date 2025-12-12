package workers

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/samber/lo"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	robotpb "robots/proto/pb-go"
)

type ProcessSummaryWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	robot  *robot.Robot
	Robots []*robot.Robot
}

func NewProcessSummaryWorker(config conf.Config, logger *slog.Logger, robot *robot.Robot, robots []*robot.Robot) ProcessSummaryWorker {
	return ProcessSummaryWorker{Config: config, Log: logger, robot: robot, Robots: robots}
}

func (w ProcessSummaryWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w ProcessSummaryWorker) GetName() string {
	return w.Name
}

func (w ProcessSummaryWorker) Run(ctx context.Context) error {
	for {
		select {
		case summaryMsg := <-w.robot.GossipSummary:
			// On doit donc retourner les secretParts manquant
			var gossipSummary robotpb.GossipSummary
			if err := proto.Unmarshal(summaryMsg, &gossipSummary); err != nil {
				w.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			indexes := lo.Map(gossipSummary.Indexes, func(item int64, _ int) int {
				return int(item)
			})
			secretParts := w.robot.GetWordsToSend(indexes)
			msg, err := proto.Marshal(&robotpb.GossipUpdate{SecretParts: robot.ToSecretPartsPb(secretParts)})
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			// ⚠️ Don't forget to add a select case and default (not just writing)
			// ⚠️ If the channel robot.GossipUpdate is slowly dequeued
			// ⚠️ Can block the process
			// Check if senderId exists
			// Find the receiver
			if gossipSummary.SenderId < 0 || int(gossipSummary.SenderId) > len(w.Robots) {
				w.Log.Debug(fmt.Sprintf("Robot %d doesn't exist", gossipSummary.SenderId))
				continue
			}
			receiver := w.Robots[gossipSummary.SenderId]
			select {
			case receiver.GossipUpdate <- msg:
				// Successfully sent the message
			default:
				w.Log.Debug(fmt.Sprintf("Robot %d : buffer is full, message is ignored", w.robot.ID))
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
