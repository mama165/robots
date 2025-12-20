package events

import (
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

// MessageDuplicatedHandler handles events where a message has been duplicated.
// It is invoked whenever a duplicate message is detected in the system, which
// can occur due to network retries, gossip, or replication. The handler can
// increment metrics, log the occurrence, or trigger alerts, but must not
// alter the original message flow.
type MessageDuplicatedHandler struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageDuplicatedHandler(log *slog.Logger, counter *Counter) *MessageDuplicatedHandler {
	return &MessageDuplicatedHandler{log: log, counter: counter}
}

func (p *MessageDuplicatedHandler) Handle(event Event) {
	switch event.EventType {
	case EventMessageDuplicated:
		if _, ok := event.Payload.(MessageDuplicatedEvent); !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventMessageDuplicated)
	}
}
