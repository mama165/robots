package events

import (
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

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
