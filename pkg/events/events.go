package events

import (
	"sync"
	"time"
)

type EventType string

const (
	EventMessageSent                           EventType = "MESSAGE_SENT"
	EventMessageReceived                       EventType = "MESSAGE_RECEIVED"
	EventMessageDuplicated                     EventType = "MESSAGE_DUPLICATED"
	EventMessageReordered                      EventType = "MESSAGE_REORDERED"
	EventSecretWritten                         EventType = "SECRET_WRITTEN"
	EventInvariantViolationSecretPartDecreased EventType = "INVARIANT_VIOLATION_SECRET_PART_DECREASED"
	EventInvariantViolationSameIndexDiffWords  EventType = "INVARIANT_VIOLATION_SAME_INDEX_DIFF_WORDS"
	EventQuiescenceDetector                    EventType = "QUIESCENCE_DETECTOR"
	EventWorkerRestartedAfterPanic             EventType = "WORKER_RESTARTED_AFTER_PANIC"
	EventChannelCapacity                       EventType = "CHANNEL_CAPACITY"
)

type Event struct {
	EventType EventType
	CreatedAt time.Time
	Payload   any
}

type MessageSentEvent struct {
	SenderID int
}

type MessageReceivedEvent struct {
	ReceiverID int
}

type MessageDuplicatedEvent struct{}
type MessageReorderedEvent struct{}
type InvariantViolationEvent struct{}

type SecretWrittenEvent struct {
	ID int
}

type QuiescenceDetectorEvent struct {
	RobotID      int
	LastActivity LastActivity
}

type WorkerRestartedAfterPanicEvent struct {
	WorkerName string
}

type ChannelCapacityEvent struct {
	WorkerName string
	Capacity   int
	Length     int
}

type LastActivity time.Time

func (l LastActivity) Date() time.Time {
	return time.Time(l)
}

type Counter struct {
	mu     sync.Mutex
	counts map[EventType]int
}

func NewCounter() *Counter {
	return &Counter{counts: make(map[EventType]int)}
}

func (c *Counter) Increment(evt EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[evt]++
}

func (c *Counter) Get(evt EventType) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[evt]
}
