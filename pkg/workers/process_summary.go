package workers

import (
	"context"
	"fmt"
	"log/slog"
	"robots/pkg/events"
	"robots/pkg/robot"
	pb "robots/proto"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/samber/lo"
)

// ProcessSummaryWorker handles incoming gossip summaries from other robots.
// It tries to send the corresponding updates to the target robots without blocking.
// If the receiver channel is full, the message is dropped to keep the system responsive.
// Channel capacity can be monitored via metrics if needed.
type ProcessSummaryWorker struct {
	Log    *slog.Logger
	Name   events.WorkerName
	robot  *robot.Robot
	Robots []*robot.Robot
	Event  chan events.Event
}

func NewProcessSummaryWorker(logger *slog.Logger, robot *robot.Robot, robots []*robot.Robot, event chan events.Event) ProcessSummaryWorker {
	return ProcessSummaryWorker{Log: logger, robot: robot, Robots: robots, Event: event}
}

func (w ProcessSummaryWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w ProcessSummaryWorker) GetName() events.WorkerName {
	return w.Name
}

func (w ProcessSummaryWorker) Run(ctx context.Context) error {
	for {
		select {
		case summaryMsg := <-w.robot.GossipSummary:
			var gossipSummary pb.GossipSummary
			if err := proto.Unmarshal(summaryMsg, &gossipSummary); err != nil {
				w.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			indexes := lo.Map(gossipSummary.Indexes, func(item int64, _ int) int {
				return int(item)
			})
			secretParts := w.robot.GetWordsToSend(indexes)
			msg, err := proto.Marshal(&pb.GossipUpdate{SecretParts: robot.ToSecretPartsPb(secretParts)})
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
				w.sendMessageReceivedEvent(ctx, receiver.ID)
			default:
				w.Log.Debug("GossipUpdate channel is full, dropping message")
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping event send")
			return nil
		}
	}
}

func (w ProcessSummaryWorker) sendMessageReceivedEvent(ctx context.Context, receiverID robot.ID) {
	select {
	case w.Event <- events.Event{
		EventType: events.EventMessageReceived,
		CreatedAt: time.Now().UTC(),
		Payload:   events.MessageReceivedEvent{ReceiverID: receiverID},
	}:
	case <-ctx.Done():
		w.Log.Debug("Context done, stopping event send")
	default:
		w.Log.Debug("Buffer is full")
	}
}
