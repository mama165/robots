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
)

// MergeSecretWorker merges incoming secret parts into the local robot state.
// It is responsible to apply merging rules (deduplication, ordering,
// versioning) to ensure that secret reconstruction is monotonic and idempotent.
// This worker mutates local state but does not perform convergence detection
// or trigger side effects. It operates independently of observability
// concerns such as UI, logging, or notifications.
type MergeSecretWorker struct {
	Log         *slog.Logger
	Name        events.WorkerName
	Robot       *robot.Robot
	DomainEvent chan events.Event
}

func NewMergeSecretWorker(logger *slog.Logger, robot *robot.Robot, DomainEvent chan events.Event) MergeSecretWorker {
	return MergeSecretWorker{Log: logger, Robot: robot, DomainEvent: DomainEvent}
}

func (w MergeSecretWorker) WithName(name string) Worker {
	w.Name = events.WorkerName(name)
	return w
}

func (w MergeSecretWorker) GetName() events.WorkerName {
	return w.Name
}

// Run processes incoming GossipUpdate messages for a robot.
// Responsibilities:
// - Merge new SecretParts into the robot's state.
// - Update LastUpdatedAt when new parts are added.
// Invariant enforcement (delegated to Robot.MergeSecretPart):
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
			for _, secretPart := range secretParts {
				w.mergeSecretPart(sendInvariantViolationEvent)(ctx, secretPart)
			}
		case <-ctx.Done():
			w.Log.Debug("Context done, stopping domainEvent send")
			return nil
		}
	}
}

func (w MergeSecretWorker) mergeSecretPart(
	recoverFunc func(ctx context.Context, r *robot.Robot, event chan events.Event),
) func(ctx context.Context, secretPart robot.SecretPart) {
	return func(ctx context.Context, part robot.SecretPart) {
		defer func() {
			if r := recover(); r != nil {
				recoverFunc(ctx, w.Robot, w.DomainEvent)
			}
		}()
		w.Robot.MergeSecretPart(part)
		return
	}
}

func sendInvariantViolationEvent(ctx context.Context, r *robot.Robot, event chan events.Event) {
	select {
	case event <- events.Event{
		EventType: events.EventInvariantViolationSameIndexDiffWords,
		CreatedAt: time.Now().UTC(),
		Payload:   events.InvariantViolationEvent{ID: r.ID},
	}:
	case <-ctx.Done():
		return
	default:
		// TODO à gérer
	}
}
