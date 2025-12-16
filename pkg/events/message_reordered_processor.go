package events

import (
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type MessageReorderedProcessor struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageReorderedProcessor(log *slog.Logger, counter *Counter) *MessageReorderedProcessor {
	return &MessageReorderedProcessor{log: log, counter: counter}
}

func (p *MessageReorderedProcessor) Handle(event Event) {
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
