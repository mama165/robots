package workers

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log/slog"
	"math/rand"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/internal/supervisor"
	robotpb "robots/proto/pb-go"
	"time"
)

type StartGossipWorker struct {
	Config conf.Config
	Log    *slog.Logger
	Name   string
	Robot  *robot.Robot
	Robots []*robot.Robot
}

func NewStartGossipWorker(config conf.Config, log *slog.Logger, robot *robot.Robot, robots []*robot.Robot) StartGossipWorker {
	return StartGossipWorker{Config: config, Log: log, Robot: robot, Robots: robots}
}

func (w StartGossipWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w StartGossipWorker) GetName() string {
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
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
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
	messageSent := 0
	for i := 0; i < w.Config.MaxAttempts; i++ {
		w.Log.Debug(fmt.Sprintf("Robot %d communicates with robot %d", sender.ID, receiver.ID))

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
			gossipSender := robotpb.GossipSummary{Indexes: sender.Indexes(), SenderId: int32(sender.ID)}
			msgSender, err := proto.Marshal(&gossipSender)
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			select {
			case receiver.GossipSummary <- msgSender:
				messageSent++
			case <-ctx.Done():
				w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
				return
			default:
				w.Log.Debug(fmt.Sprintf("Robot %d : buffer is full, message is ignored", receiver.ID))
			}
		}
	}
}
