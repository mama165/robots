package workers

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/samber/lo"
	"log/slog"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"robots/pkg/events"
	robotpb "robots/proto/pb-go"
)

// ProcessSummaryWorker handles incoming gossip summaries from other robots.
// It tries to send the corresponding updates to the target robots without blocking.
// If the receiver channel is full, the message is dropped to keep the system responsive.
// Channel capacity can be monitored via metrics if needed.
type ProcessSummaryWorker struct {
	Log    *slog.Logger
	Name   string
	robot  *robot.Robot
	Robots []*robot.Robot
	Event  chan events.Event
}

func NewProcessSummaryWorker(logger *slog.Logger, robot *robot.Robot, robots []*robot.Robot, event chan events.Event) ProcessSummaryWorker {
	return ProcessSummaryWorker{Log: logger, robot: robot, Robots: robots, Event: event}
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
			if gossipSummary.SenderId < 0 || int(gossipSummary.SenderId) >= len(w.Robots) {
				w.Log.Debug(fmt.Sprintf("Robot %d doesn't exist", gossipSummary.SenderId))
				continue
			}
			receiver := w.Robots[gossipSummary.SenderId]
			select {
			case receiver.GossipUpdate <- msg:
			default:
				w.Log.Debug("GossipUpdate channel is full, dropping message")
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
