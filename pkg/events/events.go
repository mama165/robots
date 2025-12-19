package events

import (
	"robots/pkg/robot"
	"sync"
	"time"
)

type WorkerName string

func (name WorkerName) ToString() string {
	return string(name)
}

type EventType string

const (
	EventMessageSent                          EventType = "MESSAGE_SENT"
	EventMessageReceived                      EventType = "MESSAGE_RECEIVED"
	EventMessageDuplicated                    EventType = "MESSAGE_DUPLICATED"
	EventMessageReordered                     EventType = "MESSAGE_REORDERED"
	EventMessageLost                          EventType = "MESSAGE_LOST"
	EventInvariantViolationSameIndexDiffWords EventType = "INVARIANT_VIOLATION_SAME_INDEX_DIFF_WORDS"
	EventQuiescenceDetector                   EventType = "QUIESCENCE_DETECTOR"
	EventWorkerRestartedAfterPanic            EventType = "WORKER_RESTARTED_AFTER_PANIC"
	EventChannelCapacity                      EventType = "CHANNEL_CAPACITY"
	EventAllConverged                         EventType = "ALL_CONVERGED"
)

type Event struct {
	EventType EventType
	CreatedAt time.Time
	Payload   any
}

type MessageSentEvent struct {
	SenderID robot.ID
}

type MessageReceivedEvent struct {
	ReceiverID robot.ID
}

type MessageDuplicatedEvent struct{}
type MessageReorderedEvent struct{}
type InvariantViolationEvent struct {
	ID robot.ID
}

type SecretWrittenEvent struct {
	ID int
}

type QuiescenceDetectorEvent struct {
	ID           int
	LastActivity LastActivity
}

type WorkerRestartedAfterPanicEvent struct {
	WorkerName WorkerName
}

type ChannelCapacityEvent struct {
	WorkerName WorkerName
	Capacity   int
	Length     int
}

type AllConvergedEvent struct {
	AllConverged bool
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
