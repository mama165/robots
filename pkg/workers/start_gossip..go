package workers

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	pb "robots/proto"
	"time"

	"google.golang.org/protobuf/proto"
)

type StartGossipWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   events.WorkerName
	Robot  *robot.Robot
	Robots []*robot.Robot
	Event  chan events.Event
}

func NewStartGossipWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, robots []*robot.Robot, event chan events.Event) StartGossipWorker {
	return StartGossipWorker{Config: config, Log: log, Robot: robot, Robots: robots, Event: event}
}

func (w StartGossipWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w StartGossipWorker) GetName() events.WorkerName {
	return w.Name
}

func (w StartGossipWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Config.GossipTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sender := w.Robot
			receiver := robot.ChooseRobot(sender, w.Robots)
			w.ExchangeMessage(ctx, sender, receiver)
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping event send")
			return nil
		}
	}
}

// ExchangeMessage r1 send a message to r2
// Simulate lost and duplicated messages
func (w StartGossipWorker) ExchangeMessage(ctx context.Context, sender, receiver *robot.Robot) {
	if sender.ID == receiver.ID {
		return
	}
	for i := 0; i < w.Config.MaxAttempts; i++ {
		// Calculate and simulate a random percentage
		isSimulated := func(percentage int) bool {
			return rand.Float32() < float32(percentage)/100.0
		}

		// Percentage of lost messages
		if isSimulated(w.Config.PercentageOfLost) {
			continue
		}

		// Percentage of duplicated messages
		var times int
		if isSimulated(w.Config.PercentageOfDuplicated) {
			times = w.Config.DuplicatedNumber
		}

		for j := 0; j <= times; j++ {
			// Sender sends his own indexes to receiver
			gossipSender := pb.GossipSummary{Indexes: sender.Indexes(), SenderId: int32(sender.ID)}
			msgSender, err := proto.Marshal(&gossipSender)
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			select {
			case receiver.GossipSummary <- msgSender:
				w.sendMessageSentEvent(ctx, sender)
			case <-ctx.Done():
				w.Log.Debug("Context done, stopping event send")
				return
			default:
				w.Log.Debug("StartGossip channel is full, dropping message")
			}
		}
	}
}

func (w StartGossipWorker) sendMessageSentEvent(ctx context.Context, sender *robot.Robot) {
	select {
	case w.Event <- events.Event{
		EventType: events.EventMessageSent,
		CreatedAt: time.Now().UTC(),
		Payload:   events.MessageSentEvent{SenderID: sender.ID},
	}:
	case <-ctx.Done():
		w.Log.Debug("Context done, stopping event send")
	default:
		w.Log.Debug("Buffer is full")
	}
}
