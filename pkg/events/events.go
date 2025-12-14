package events

import "time"

type EventType string

const (
	EventStartGossip               EventType = "START_GOSSIP"
	EventMessageSent               EventType = "MESSAGE_SENT"
	EventMessageReceived           EventType = "MESSAGE_RECEIVED"
	EventMessageDuplicated         EventType = "MESSAGE_DUPLICATED"
	EventMessageReordered          EventType = "MESSAGE_REORDERED"
	EventWinnerFound               EventType = "WINNER_FOUND"
	EventSecretWritten             EventType = "SECRET_WRITTEN"
	EventInvariantViolation        EventType = "INVARIANT_VIOLATION"
	EventQuiescenceReached         EventType = "QUIET_REACHED"
	EventSupervisorStarted         EventType = "SUPERVISOR_STARTED"
	EventWorkerRestartedAfterPanic EventType = "WORKER_RESTARTED_AFTER_PANIC"
	EventChannelCapacity           EventType = "CHANNEL_CAPACITY"
)

type Event struct {
	EventType EventType
	CreatedAt time.Time
	Payload   interface{}
}

type StartGossipEvent struct {
	Sender    int
	Receivers int
}

type MessageSentEvent struct {
	Count int
}

type MessageReceivedEvent struct {
	Count int
}

type MessageDuplicatedEvent struct {
}

type MessageReorderedEvent struct {
}

type WinnerFoundEvent struct {
}

type SecretWrittenEvent struct {
}

type InvariantViolationEvent struct {
	WorkerName string
}

type QuietReachedEvent struct {
}

type SupervisorStartedEvent struct {
}

type WorkerRestartedAfterPanicEvent struct {
}

type ChannelCapacityEvent struct {
	WorkerName string
	Capacity   int
	Length     int
}
