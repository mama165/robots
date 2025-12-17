package events

import (
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type MessageDuplicatedProcessor struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageDuplicatedProcessor(log *slog.Logger, counter *Counter) *MessageDuplicatedProcessor {
	return &MessageDuplicatedProcessor{log: log, counter: counter}
}

func (p *MessageDuplicatedProcessor) Handle(event Event) {
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
