package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type MessageSentHandler struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageSentHandler(log *slog.Logger, counter *Counter) *MessageSentHandler {
	return &MessageSentHandler{log: log, counter: counter}
}

func (p *MessageSentHandler) Handle(event Event) {
	switch event.EventType {
	case EventMessageSent:
		payload, ok := event.Payload.(MessageSentEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventMessageSent)
		p.log.Debug(fmt.Sprintf("Robot %d sent a message", payload.SenderID))
	}
}
