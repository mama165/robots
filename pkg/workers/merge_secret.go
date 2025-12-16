package workers

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"robots/internal/robot"
	"robots/internal/supervisor"
	"robots/pkg/events"
	pb "robots/proto"
	"time"
)

// MergeSecretWorker Fetch all missing parts coming from anybody
type MergeSecretWorker struct {
	Log   *slog.Logger
	Name  string
	Robot *robot.Robot
	Event chan events.Event
}

func NewMergeSecretWorker(logger *slog.Logger, robot *robot.Robot, event chan events.Event) MergeSecretWorker {
	return MergeSecretWorker{Log: logger, Robot: robot, Event: event}
}

func (w MergeSecretWorker) WithName(name string) supervisor.Worker {
	w.Name = name
	return w
}

func (w MergeSecretWorker) GetName() string {
	return w.Name
}

// Run processes incoming GossipUpdate messages for a robot.
// Responsibilities:
// - Merge new SecretParts into the robot's state.
// - Update LastUpdatedAt when new parts are added.
// Invariant enforcement:
// - Monotonicity: robot never loses a SecretPart.
// - Uniqueness: each index maps to exactly one word; conflicting parts trigger panic.
// - Duplicate messages with the same word are ignored (idempotence).
// Resilience:
// - Runs until context cancellation (timeout or CTRL+C).
// - Panics are caught by the Supervisor and the worker is restarted.
// This worker ensures the robot's local state grows correctly and consistently,
// enabling the gossip protocol to achieve eventual convergence.
func (w MergeSecretWorker) Run(ctx context.Context) error {
	for {
		select {
		case updateMsg := <-w.Robot.GossipUpdate:
			var gossipUpdate pb.GossipUpdate
			err := proto.Unmarshal(updateMsg, &gossipUpdate)
			if err != nil {
				w.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			secretParts := robot.FromSecretPartsPb(gossipUpdate.SecretParts)
			before := len(w.Robot.SecretParts)
			for _, secretPart := range secretParts {
				w.Robot.MergeSecretPart(ctx, secretPart, w.Log, w.Event)
			}
			after := len(w.Robot.SecretParts)
			if after < before {
				robot.SendInvariantViolationEvent(ctx, w.Log, events.EventInvariantViolationSecretPartDecreased, w.Event)
				select {
				case w.Event <- events.Event{
					EventType: events.EventInvariantViolationSecretPartDecreased,
					CreatedAt: time.Time{},
				}:
					panic("INVARIANT VIOLATION: secret parts count decreased")
				}
			}
		case <-ctx.Done():
			w.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
			return nil
		}
	}
}
