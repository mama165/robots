package events

import (
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type MessageReorderedHandler struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageReorderedHandler(log *slog.Logger, counter *Counter) *MessageReorderedHandler {
	return &MessageReorderedHandler{log: log, counter: counter}
}

func (p *MessageReorderedHandler) Handle(event Event) {
	switch event.EventType {
	case EventMessageReordered:
		_, ok := event.Payload.(MessageReorderedEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventMessageReordered)
	}
}
